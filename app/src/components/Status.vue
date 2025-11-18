<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-container fluid>
    <v-row
      align="center"
      justify="center"
    >
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
            <v-btn
              icon
              variant="text"
              @click="refreshData"
            >
              <v-icon>
                {{ mdiRefresh }}
              </v-icon>
            </v-btn>
          </div>
          <v-card-text>
            <v-skeleton-loader
              v-if="!data"
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
                    <v-col v-if="data.status.lastclusteredhmiid > 0">
                      <icon-item
                        :icon="mdiDatabaseSearch"
                        title="Hierarchical Multi-Input Clustering"
                        :tooltip="tooltips.databaseClusteringHMI"
                        is-color
                        :is-red="!data.status.isclusteringhmi"
                      >
                        <v-progress-linear
                          :color="clusteringHMISyncProgress > 98?'green'
                            :clusteringHMISyncProgress > 90?'light-green':'light-blue'"
                          height="17"
                          :model-value="clusteringHMISyncProgress"
                          rounded
                        >
                          {{ Math.round(clusteringHMISyncProgress) }}%
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
                    <v-col v-if="data.status.lastclusteredhmiid">
                      <icon-item
                        :icon="mdiCounter"
                        title="Last HMI Block"
                      >
                        <router-link
                          :to="{ name: ROUTE_NAME_BLOCK_PAGE,
                                 params: { id: data.status.lastclusteredhmiid, blockchainMode: blockchainMode }}"
                        >
                          {{ data.status.lastclusteredhmiid.toLocaleString() }}
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
	mdiRefresh, mdiDatabase, mdiDatabaseSync, mdiDatabaseSearch, mdiCounter,
} from '@mdi/js';
import {PAGE_TITLE, ROUTE_NAME_BLOCK_PAGE} from '@/constants';
import IconItem from './common/IconItem.vue';
import {getDakarClient, handleError} from '@/utilities';
import IconTitle from '@/components/common/IconTitle.vue';
import {
	computed, ref, onMounted, onBeforeUnmount,
} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const props = defineProps({
	title: {type: String, required: true},
	blockchainMode: {type: String, required: true},
});

const context = {$route: useRoute(), addMessage: useMsgStore().addMessage};
const dakar = getDakarClient(props.blockchainMode);

const tooltips = {
	databaseSync: 'Percentage of blocks synced from the RPC client to the database. The crawler is active if the icon is green.',
	databaseClassification: 'Percentage of classified blocks in the database. The classifier is active if the icon is green.',
	databaseClusteringHMI: 'Percentage of hierarchical multi-input clustered blocks in the database. '
		+ 'Clustering is ongoing if the icon is green.',
	databaseClusteringFMI: 'Percentage of flat multi-input clustered blocks in the database. '
		+ 'Clustering is ongoing if the icon is green.',
	rpcDifficulty: 'Current mining difficulty',
	rpcPruned: 'Whether the RPC client prunes blocks',
	rpcVerificationProgress: 'Estimate of verification progress of the RPC client',
	rpcBlockchainSize: 'The estimated size of the block and undo files on disk',
};

const refreshStep = 10000;
let timer = null;
const data = ref(null);

// Computed

const crawlerSyncProgress = computed(() => {
	if (!data.value) {
		return 0.0;
	}

	return data.value.status.lastblockid / data.value.blocks * 100;
});

const classifierSyncProgress = computed(() => {
	if (!data.value) {
		return 0.0;
	}

	const percentage = data.value.status.lastclassifiedid / data.value.status.lastblockid * 100;

	return percentage > 100 ? 100 : percentage;
});

const clusteringHMISyncProgress = computed(() => {
	if (!data.value) {
		return 0.0;
	}

	const percentage = data.value.status.lastclusteredhmiid / data.value.status.lastblockid * 100;

	return percentage > 100 ? 100 : percentage;
});

const clusteringFMISyncProgress = computed(() => {
	if (!data.value) {
		return 0.0;
	}

	const percentage = data.value.status.lastclusteredfmiid / data.value.status.lastblockid * 100;

	return percentage > 100 ? 100 : percentage;
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
	} catch (e) {
		handleError(context, e);
		return false;
	}
}

async function refreshData() {
	resetTimers();
	if (await loadStatusData()) {
		startTimer();
	}
}

onMounted(() => {
	document.title = `Status - ${PAGE_TITLE}`;
});

onBeforeUnmount(() => {
	resetTimers();
});

// Initially get data
refreshData();

</script>

<style scoped>

</style>
