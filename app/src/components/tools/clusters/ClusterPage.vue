<template>
  <div>
    <v-card
      variant="text"
      class="mx-auto"
      max-width="1200"
    >
      <icon-title
        title="Custom Clusters"
        :icon="icon.mdiMerge"
        :one-line="true"
      >
        <v-menu location="bottom">
          <template #activator="{ props }">
            <v-btn
              icon
              v-bind="props"
              variant="text"
            >
              <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="addClusterDialog = true">
              <template #prepend>
                <v-icon>{{ icon.mdiFileImport }}</v-icon>
              </template>
              <v-list-item-title>Import Clusters</v-list-item-title>
            </v-list-item>
            <v-list-item
              :disabled="items.length === 0"
              @click="deleteAllClustersDialog = true"
            >
              <template #prepend>
                <v-icon>{{ icon.mdiDelete }}</v-icon>
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
                @click="addClusterDialog = true"
              >
                <v-icon>{{ icon.mdiFileImport }}</v-icon>
                Import Clusters
              </v-btn>
            </div>
          </v-col>
        </v-row>
      </v-card-text>
      <import-cluster-dialog
        v-model="addClusterDialog"
        @added="loadData"
      />
      <delete-all-clusters-dialog
        v-model="deleteAllClustersDialog"
        @deleted="loadData"
      />
      <delete-cluster-dialog
        v-model="deleteClusterDialog"
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
                  <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
                </v-btn>
              </template>
              <v-list>
                <v-list-item @click="deleteItem(item.uid, item.address_count)">
                  <template #prepend>
                    <v-icon>{{ icon.mdiDelete }}</v-icon>
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
            :to="{ name: routes.addressRoute, params: { id: address }}"
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

<script>
import {mdiMerge, mdiDelete, mdiDotsVertical, mdiFileImport} from '@mdi/js';
import {PAGE_TITLE, ROUTE_NAME_ADDRESS_PAGE} from '@/constants';
import {handleError} from '@/utilities';
import ImportClusterDialog from './ImportClustersDialog.vue';
import DeleteClusterDialog from './DeleteClusterDialog.vue';
import DeleteAllClustersDialog from './DeleteAllClustersDialog.vue';
import WikiTooltip from '../../wiki/WikiTooltip.vue';
import IconTitle from '@/components/common/IconTitle.vue';

export default {
	name: 'ClusterPage',
	components: {
		IconTitle, WikiTooltip, DeleteAllClustersDialog, DeleteClusterDialog, ImportClusterDialog,
	},
	data() {
		return {
			icon: {
				mdiMerge, mdiDelete, mdiDotsVertical, mdiFileImport,
			},
			routes: {
				addressRoute: ROUTE_NAME_ADDRESS_PAGE,
			},
			addClusterDialog: false,
			deleteClusterDialog: false,
			deleteAllClustersDialog: false,
			isLoading: false,
			deleteClusterUid: '',
			deleteClusterSize: -1,
			items: [],
		};
	},
	async mounted() {
		document.title = `Custom Clusters - ${PAGE_TITLE}`;
		await this.loadData();
	},
	methods: {
		async loadData() {
			this.items = [];
			this.isLoading = true;

			try {
				const response = await 	this.dakar.cluster.clusterOverviewGet();

				if (response.clusters) {
					// Parse date
					response.clusters = response.clusters.map(d => {
						d.ts = new Date(d.ts);
						return d;
					});

					// Sort clusters by time stamp
					this.items = response.clusters.sort((a, b) => b.ts - a.ts);
				}
			} catch (e) {
				handleError(this, e);
			}

			this.isLoading = false;
		},
		deleteItem(clusterUid, clusterSize) {
			this.deleteClusterUid = clusterUid;
			this.deleteClusterSize = clusterSize;
			this.deleteClusterDialog = true;
		},
		handleClusterDeletion(clusterUid) {
			this.items = this.items.filter(d => d.uid !== clusterUid);
		},
	},
};
</script>

<style scoped>

</style>
