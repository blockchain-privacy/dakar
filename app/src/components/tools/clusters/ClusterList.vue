<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div>
    <v-card
      variant="text"
      class="mx-auto"
      max-width="1200"
    >
      <icon-title
        :title="`Custom ${title} Clusters`"
        :icon="mdiMerge"
        one-line
      >
        <v-menu location="bottom">
          <template #activator="item">
            <v-icon-btn
              v-bind="item.props"
              variant="text"
            >
              <v-icon>{{ mdiDotsVertical }}</v-icon>
            </v-icon-btn>
          </template>
          <v-list>
            <v-list-item @click="addClusterDialogModel = true">
              <template #prepend>
                <v-icon>{{ mdiFileImport }}</v-icon>
              </template>
              <v-list-item-title>Import Clusters</v-list-item-title>
            </v-list-item>
            <v-list-item
              :disabled="items.length === 0"
              @click="deleteAllClustersDialogModel = true"
            >
              <template #prepend>
                <v-icon>{{ mdiDelete }}</v-icon>
              </template>
              <v-list-item-title>Delete All Custom Clusters</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
      </icon-title>
      <v-card-text>
        <p class="text-body-large mb-3">
          <wiki-tooltip description-url="addressCluster.md">
            Clusters
          </wiki-tooltip>
          created here can be used to refine transaction heuristics.
        </p>
        <v-progress-linear
          v-if="isLoading"
          indeterminate
        />
        <alert
          v-else-if="failedLoading"
          text="Failed loading data. Please try again later."
        />
        <v-row v-else-if="items.length === 0">
          <v-col>
            <div class="d-flex justify-center">
              <v-btn
                variant="text"
                @click="addClusterDialogModel = true"
              >
                <v-icon>{{ mdiFileImport }}</v-icon>
                Import Clusters
              </v-btn>
            </div>
          </v-col>
        </v-row>
      </v-card-text>
      <import-cluster-dialog
        v-model="addClusterDialogModel"
        :blockchain-mode="blockchainMode"
        :title="title"
        @added="loadData"
      />
      <delete-all-clusters-dialog
        v-model="deleteAllClustersDialogModel"
        :blockchain-mode="blockchainMode"
        :title="title"
        @deleted="loadData"
      />
      <delete-cluster-dialog
        v-model="deleteClusterDialogModel"
        :cluster-uid="deleteClusterUid"
        :num-addresses="deleteClusterSize"
        :blockchain-mode="blockchainMode"
        :title="title"
        @deleted="handleClusterDeletion"
      />
    </v-card>
    <v-row
      v-if="items.length > 0"
      class="mt-3 mx-auto"
      style="max-width: 1200px"
    >
      <div
        class="d-flex flex-wrap align-baseline"
        style="gap: 20px 20px"
      >
        <v-card
          v-for="(item, i) in items"
          :key="i"
        >
          <div class="mx-4 mt-2 d-flex align-center">
            <v-list-item-title class="me-auto">
              {{ item.address_count.toLocaleString() }} Addresses
            </v-list-item-title>
            <v-list-item-subtitle>
              {{ item.ts.toLocaleDateString() }}
            </v-list-item-subtitle>
            <v-menu location="bottom">
              <template #activator="activatorItem">
                <v-icon-btn
                  v-bind="activatorItem.props"
                  variant="plain"
                >
                  <v-icon>{{ mdiDotsVertical }}</v-icon>
                </v-icon-btn>
              </template>
              <v-list>
                <v-list-item @click="deleteItem(item.uid, item.address_count)">
                  <template #prepend>
                    <v-icon>{{ mdiDelete }}</v-icon>
                  </template>
                  <v-list-item-title>Delete</v-list-item-title>
                </v-list-item>
              </v-list>
            </v-menu>
          </div>
          <v-divider />
          <v-list-item
            v-for="address in item.addresses"
            :key="address"
            :to="{ name: ROUTE_NAME_ADDRESS_PAGE, params: { id: address, blockchainMode: blockchainMode }}"
          >
            <div>
              {{ address }}
            </div>
          </v-list-item>
        </v-card>
      </div>
    </v-row>
  </div>
</template>

<script setup>
import {
	mdiMerge,
	mdiDelete,
	mdiDotsVertical,
	mdiFileImport,
} from '@mdi/js';
import {onMounted, ref} from 'vue';
import WikiTooltip from '../../wiki/WikiTooltip.vue';
import ImportClusterDialog from './ImportClustersDialog.vue';
import DeleteClusterDialog from './DeleteClusterDialog.vue';
import DeleteAllClustersDialog from './DeleteAllClustersDialog.vue';
import {PAGE_TITLE, ROUTE_NAME_ADDRESS_PAGE} from '@/constants';
import {getDakarClient} from '@/utilities';
import IconTitle from '@/components/common/IconTitle.vue';
import Alert from '@/components/common/Alert.vue';

const props = defineProps({
	title: {type: String, required: true},
	blockchainMode: {type: String, required: true},
});

const dakar = getDakarClient(props.blockchainMode);

const addClusterDialogModel = ref(false);
const deleteClusterDialogModel = ref(false);
const deleteAllClustersDialogModel = ref(false);
const isLoading = ref(false);
const failedLoading = ref(false);
const deleteClusterUid = ref('');
const deleteClusterSize = ref(-1);
const items = ref([]);

// Hooks
onMounted(async () => {
	document.title = `Custom Clusters - ${PAGE_TITLE}`;
	await loadData();
});

// Functions
async function loadData() {
	items.value = [];
	isLoading.value = true;
	failedLoading.value = false;

	try {
		const response = await dakar.cluster.clustersGet();

		if (response.clusters) {
			// Parse date
			response.clusters = response.clusters.map(d => {
				d.ts = new Date(d.ts);
				return d;
			});

			// Sort clusters by time stamp
			items.value = response.clusters.toSorted((a, b) => b.ts - a.ts);
		}
	} catch {
		failedLoading.value = true;
	}

	isLoading.value = false;
}

function deleteItem(clusterUid, clusterSize) {
	deleteClusterUid.value = clusterUid;
	deleteClusterSize.value = clusterSize;
	deleteClusterDialogModel.value = true;
}

function handleClusterDeletion(clusterUid) {
	items.value = items.value.filter(d => d.uid !== clusterUid);
}

</script>

<style scoped>

</style>
