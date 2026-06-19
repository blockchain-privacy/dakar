<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-container fluid>
    <v-row class="align-center justify-center">
      <v-col
        cols="12"
        md="10"
        lg="9"
        xl="8"
      >
        <v-card variant="text">
          <div class="d-flex align-center">
            <icon-title
              :title="`${title} Server Status`"
              :icon="mdiDatabase"
            />
            <v-spacer />
            <v-icon-btn
              :icon="mdiRefresh"
              variant="text"
              @click="refreshData"
            />
          </div>
          <v-card-text>
            <alert
              v-if="errorMsg"
              :text="errorMsg"
            />
            <v-skeleton-loader
              v-else-if="!data"
              type="table-tbody"
            />
            <div v-else>
              <v-row>
                <v-col
                  cols="12"
                  md="8"
                >
                  <v-row>
                    <v-col>
                      <icon-item
                        :icon="mdiDatabaseSync"
                        title="Chain Synchronisation"
                        :tooltip="tooltips.databaseSync"
                        is-color
                        :is-red="!data.status.iscrawling"
                      >
                        <v-progress-linear
                          :color="crawlerSyncProgress > 98?'green'
                            :crawlerSyncProgress > 90?'light-green':'light-blue'"
                          height="17"
                          :model-value="crawlerSyncProgress"
                          rounded
                        >
                          {{ Math.round(crawlerSyncProgress) }}%
                        </v-progress-linear>
                      </icon-item>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col>
                      <icon-item
                        :icon="mdiDatabaseSearch"
                        title="Transaction Classification"
                        :tooltip="tooltips.databaseClassification"
                        is-color
                        :is-red="!data.status.isclassifying"
                      >
                        <v-progress-linear
                          :color="classifierSyncProgress > 98?'green'
                            :classifierSyncProgress > 90?'light-green':'light-blue'"
                          height="17"
                          :model-value="classifierSyncProgress"
                          rounded
                        >
                          {{ Math.round(classifierSyncProgress) }}%
                        </v-progress-linear>
                      </icon-item>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col v-if="data.status.lastclusteredfmiid > 0">
                      <icon-item
                        :icon="mdiDatabaseSearch"
                        title="Flat Multi-Input Clustering"
                        :tooltip="tooltips.databaseClusteringFMI"
                        is-color
                        :is-red="!data.status.isclusteringfmi"
                      >
                        <v-progress-linear
                          :color="clusteringFMISyncProgress > 98?'green'
                            :clusteringFMISyncProgress > 90?'light-green':'light-blue'"
                          height="17"
                          :model-value="clusteringFMISyncProgress"
                          rounded
                        >
                          {{ Math.round(clusteringFMISyncProgress) }}%
                        </v-progress-linear>
                      </icon-item>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col />
                  </v-row>
                </v-col>
                <v-col
                  cols="12"
                  md="4"
                >
                  <v-row>
                    <v-col>
                      <icon-item
                        :icon="mdiCounter"
                        title="Last crawled Block"
                      >
                        <router-link
                          :to="{ name: ROUTE_NAME_BLOCK_PAGE,
                                 params: { id: data.status.lastblockid, blockchainMode: blockchainMode }}"
                        >
                          {{ data.status.lastblockid.toLocaleString() }}
                        </router-link>
                      </icon-item>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col v-if="data.status.lastclassifiedid">
                      <icon-item
                        :icon="mdiCounter"
                        title="Last classified Block"
                      >
                        <router-link
                          :to="{ name: ROUTE_NAME_BLOCK_PAGE,
                                 params: { id: data.status.lastclassifiedid, blockchainMode: blockchainMode }}"
                        >
                          {{ data.status.lastclassifiedid.toLocaleString() }}
                        </router-link>
                      </icon-item>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col v-if="data.status.lastclusteredfmiid">
                      <icon-item
                        :icon="mdiCounter"
                        title="Last FMI Block"
                      >
                        <router-link
                          :to="{ name: ROUTE_NAME_BLOCK_PAGE,
                                 params: { id: data.status.lastclusteredfmiid, blockchainMode: blockchainMode }}"
                        >
                          {{ data.status.lastclusteredfmiid.toLocaleString() }}
                        </router-link>
                      </icon-item>
                    </v-col>
                  </v-row>
                  <v-row v-if="data.blocks > 0">
                    <v-col>
                      <icon-item
                        :icon="mdiCounter"
                        title="RPC Client Block Height"
                      >
                        {{ data.blocks.toLocaleString() }}
                      </icon-item>
                    </v-col>
                  </v-row>
                </v-col>
              </v-row>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>
<script setup>
import {
	mdiRefresh,
	mdiDatabase,
	mdiDatabaseSync,
	mdiDatabaseSearch,
	mdiCounter,
} from '@mdi/js';
import {
	computed,
	ref,
	onMounted,
	onBeforeUnmount,
} from 'vue';
import IconItem from './common/IconItem.vue';
import {ROUTE_NAME_BLOCK_PAGE} from '@/constants';
import {getDakarClient} from '@/utilities';
import IconTitle from '@/components/common/IconTitle.vue';
import Alert from '@/components/common/Alert.vue';

const props = defineProps({
	title: {type: String, required: true},
	blockchainMode: {type: String, required: true},
});

const dakar = getDakarClient(props.blockchainMode);

const data = ref(null);
const errorMsg = ref('');
const tooltips = {
	databaseSync: 'Percentage of blocks synced from the RPC client to the database. The crawler is active if the icon is green.',
	databaseClassification: 'Percentage of classified blocks in the database. The classifier is active if the icon is green.',
	databaseClusteringFMI: 'Percentage of flat multi-input clustered blocks in the database. '
		+ 'Clustering is ongoing if the icon is green.',
	rpcDifficulty: 'Current mining difficulty',
	rpcPruned: 'Whether the RPC client prunes blocks',
	rpcVerificationProgress: 'Estimate of verification progress of the RPC client',
	rpcBlockchainSize: 'The estimated size of the block and undo files on disk',
};
const refreshStep = 10_000;
let timer = null;

// Computed

const crawlerSyncProgress = computed(() => {
	if (!data.value) {
		return 0;
	}

	return data.value.status.lastblockid / data.value.blocks * 100;
});

const classifierSyncProgress = computed(() => {
	if (!data.value) {
		return 0;
	}

	const percentage = data.value.status.lastclassifiedid / data.value.status.lastblockid * 100;

	return Math.min(percentage, 100);
});

const clusteringFMISyncProgress = computed(() => {
	if (!data.value) {
		return 0;
	}

	const percentage = data.value.status.lastclusteredfmiid / data.value.status.lastblockid * 100;

	return Math.min(percentage, 100);
});

// Hooks
onMounted(() => {
	// Initially get data
	refreshData();
});

onBeforeUnmount(() => {
	resetTimers();
});

// Functions

function startTimer() {
	timer = setInterval(async () => {
		await loadStatusData();
	}, refreshStep);
}

function resetTimers() {
	clearInterval(timer);
}

async function loadStatusData() {
	try {
		data.value = await dakar.meta.metaGet();
		return true;
	} catch (error) {
		errorMsg.value = error.message;
		return false;
	}
}

async function refreshData() {
	errorMsg.value = '';
	resetTimers();
	if (await loadStatusData()) {
		startTimer();
	}
}

</script>

<style scoped>

</style>
