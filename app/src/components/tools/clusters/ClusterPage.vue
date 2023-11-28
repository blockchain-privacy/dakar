<template>
  <div>
    <v-card
      variant="text"
      class="mx-auto"
      max-width="1200"
    >
      <icon-title
        title="Custom Clusters"
        :icon="mdiMerge"
        :one-line="true"
      >
        <v-menu location="bottom">
          <template #activator="{ props }">
            <v-btn
              icon
              v-bind="props"
              variant="text"
            >
              <v-icon>{{ mdiDotsVertical }}</v-icon>
            </v-btn>
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
        <p class="text-subtitle-1 mb-3">
          <wiki-tooltip description-url="addressCluster.md">
            Clusters
          </wiki-tooltip>
          created here can be used to refine transaction heuristics.
        </p>
        <v-progress-linear
          v-if="isLoading"
          :indeterminate="true"
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
        @added="loadData"
      />
      <delete-all-clusters-dialog
        v-model="deleteAllClustersDialogModel"
        @deleted="loadData"
      />
      <delete-cluster-dialog
        v-model="deleteClusterDialogModel"
        :cluster-uid="deleteClusterUid"
        :num-addresses="deleteClusterSize"
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
              <template #activator="{ props }">
                <v-btn
                  icon
                  v-bind="props"
                  variant="plain"
                >
                  <v-icon>{{ mdiDotsVertical }}</v-icon>
                </v-btn>
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
            :to="{ name: ROUTE_NAME_ADDRESS_PAGE, params: { id: address }}"
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
import {mdiMerge, mdiDelete, mdiDotsVertical, mdiFileImport} from '@mdi/js';
import {PAGE_TITLE, ROUTE_NAME_ADDRESS_PAGE} from '@/constants';
import {handleError} from '@/utilities';
import ImportClusterDialog from './ImportClustersDialog.vue';
import DeleteClusterDialog from './DeleteClusterDialog.vue';
import DeleteAllClustersDialog from './DeleteAllClustersDialog.vue';
import WikiTooltip from '../../wiki/WikiTooltip.vue';
import IconTitle from '@/components/common/IconTitle.vue';
import {inject, onMounted, ref} from 'vue';
import {useRoute} from 'vue-router';
import {useStore} from 'vuex';

const dakar = inject('dakar');
const route = useRoute();
const store = useStore();
const context = {$store: store, $route: route};

const addClusterDialogModel = ref(false);
const deleteClusterDialogModel = ref(false);
const deleteAllClustersDialogModel = ref(false);
const isLoading = ref(false);
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

	try {
		const response = await 	dakar.cluster.clusterOverviewGet();

		if (response.clusters) {
			// Parse date
			response.clusters = response.clusters.map(d => {
				d.ts = new Date(d.ts);
				return d;
			});

			// Sort clusters by time stamp
			items.value = response.clusters.sort((a, b) => b.ts - a.ts);
		}
	} catch (e) {
		handleError(context, e);
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
