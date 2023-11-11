<template>
  <div class="my-2 mx-1">
    <v-card variant="text">
      <v-card-text>
        <v-progress-linear
          v-if="loading"
          :indeterminate="true"
        />
        <div v-else>
          <v-row>
            <v-col
              v-if="items.length > 0"
              class="d-flex"
            >
              <p class="text-subtitle-1 my-auto mr-auto">
                Attributions help to easier identify addresses belonging to the same
                <WikiTooltip description-url="addressCluster.md">
                  address cluster
                </WikiTooltip>.
              </p>
              <v-menu location="bottom">
                <template #activator="{ props }">
                  <v-btn
                    icon
                    variant="text"
                    v-bind="props"
                  >
                    <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
                  </v-btn>
                </template>
                <v-list>
                  <v-list-item @click="addAttributionDialog = true">
                    <template #prepend>
                      <v-icon>{{ icon.mdiTagPlus }}</v-icon>
                    </template>
                    <v-list-item-title>Import Attributions</v-list-item-title>
                  </v-list-item>
                  <v-list-item @click="deleteAllAttributionsDialog = true">
                    <template #prepend>
                      <v-icon>{{ icon.mdiDelete }}</v-icon>
                    </template>
                    <v-list-item-title>Delete All Attributions</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>
            </v-col>
            <v-col v-else>
              <div class="d-flex justify-center">
                <v-btn
                  variant="text"
                  @click="addAttributionDialog = true"
                >
                  <v-icon>{{ icon.mdiFileImport }}</v-icon>
                  Import attributions
                </v-btn>
              </div>
            </v-col>
          </v-row>
        </div>
      </v-card-text>
      <import-attribution-dialog
        v-model="addAttributionDialog"
        @added="loadOverviewData"
      />
      <delete-all-attributions-dialog
        v-model="deleteAllAttributionsDialog"
        @deleted="loadOverviewData"
      />
    </v-card>
    <v-row
      v-if="items.length > 0"
      class="mt-3 mx-auto mb-2"
    >
      <div
        class="d-flex flex-wrap align-baseline"
        style="gap: 20px 20px"
      >
        <attribution-details
          v-for="(item, i) in items"
          :key="i"
          :attribution="item"
          @deleted="handleAttributionDeletion"
        />
      </div>
    </v-row>
  </div>
</template>

<script>
import {
	mdiMerge, mdiDelete, mdiDotsVertical,
	mdiFileImport, mdiTagPlus, mdiClose,
} from '@mdi/js';
import {PAGE_TITLE} from '@/constants';
import {handleError} from '@/utilities';
import ImportAttributionDialog from './ImportAttributionsDialog.vue';
import DeleteAllAttributionsDialog from './DeleteAllAttributionsDialog.vue';
import AttributionDetails from './AttributionDetails.vue';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';

export default {
	name: 'AttributionOverview',
	components: {
		WikiTooltip,
		AttributionDetails, DeleteAllAttributionsDialog, ImportAttributionDialog,
	},
	data() {
		return {
			icon: {
				mdiMerge,
				mdiDelete,
				mdiDotsVertical,
				mdiFileImport,
				mdiTagPlus,
				mdiClose,
			},
			loading: false,
			addAttributionDialog: false,
			deleteAllAttributionsDialog: false,
			items: [],
			fab: false,
		};
	},
	async mounted() {
		document.title = `Attribution Overview - ${PAGE_TITLE}`;
		await this.loadOverviewData();
	},
	methods: {
		async loadOverviewData() {
			this.loading = true;
			this.items = [];

			try {
				const response = await this.dakar.attribution.attributionOverviewGet();

				if (response.attributions) {
					// Parse date
					response.attributions = response.attributions.map(d => {
						d.ts = new Date(d.ts);
						return d;
					});

					// Sort attributions by time stamp
					this.items = response.attributions.sort((a, b) => b.ts - a.ts);
				}
			} catch (e) {
				handleError(this, e);
			}

			this.loading = false;
		},
		handleAttributionDeletion(attributionUid) {
			this.items = this.items.filter(d => d.uid !== attributionUid);
		},
	},
};
</script>

<style scoped>

</style>
