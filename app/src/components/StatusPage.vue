<template>
  <v-container :fluid="true">
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
            <IconTitle
              title="Server Status"
              :icon="icon.mdiDatabase"
            />
            <v-spacer />
            <v-btn
              icon
              variant="text"
              @click="refreshData"
            >
              <v-icon>
                {{ icon.mdiRefresh }}
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
                      <IconItem
                        :icon="icon.mdiDatabaseSync"
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
                      </IconItem>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col>
                      <IconItem
                        :icon="icon.mdiDatabaseSearch"
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
                      </IconItem>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col v-if="data.status.lastclusteredhmiid > 0">
                      <IconItem
                        :icon="icon.mdiDatabaseSearch"
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
                      </IconItem>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col v-if="data.status.lastclusteredfmiid > 0">
                      <IconItem
                        :icon="icon.mdiDatabaseSearch"
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
                      </IconItem>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col>
                      <IconItem
                        :icon="icon.mdiProgressWrench"
                        title="RPC Verification Progress"
                        :tooltip="tooltips.rpcVerificationProgress"
                      >
                        <v-progress-linear
                          :color="data.rpcinfo.verificationprogress > 98?'green'
                            :data.rpcinfo.verificationprogress > 90?'light-green':'light-blue'"
                          height="17"
                          :model-value="data.rpcinfo.verificationprogress"
                          rounded
                        >
                          {{ Math.round(data.rpcinfo.verificationprogress) }}%
                        </v-progress-linear>
                      </IconItem>
                    </v-col>
                  </v-row>
                </v-col>
                <v-col
                  cols="12"
                  md="4"
                >
                  <v-row>
                    <v-col>
                      <IconItem
                        :icon="icon.mdiCounter"
                        title="Last crawled Block"
                      >
                        <router-link
                          :to="{ name: blockRoute,
                                 params: { id: data.status.lastblockid }}"
                        >
                          {{ data.status.lastblockid.toLocaleString() }}
                        </router-link>
                      </IconItem>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col v-if="data.status.lastclassifiedid">
                      <IconItem
                        :icon="icon.mdiCounter"
                        title="Last classified Block"
                      >
                        <router-link
                          :to="{ name: blockRoute,
                                 params: { id: data.status.lastclassifiedid }}"
                        >
                          {{ data.status.lastclassifiedid.toLocaleString() }}
                        </router-link>
                      </IconItem>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col v-if="data.status.lastclusteredhmiid">
                      <IconItem
                        :icon="icon.mdiCounter"
                        title="Last HMI Block"
                      >
                        <router-link
                          :to="{ name: blockRoute,
                                 params: { id: data.status.lastclusteredhmiid }}"
                        >
                          {{ data.status.lastclusteredhmiid.toLocaleString() }}
                        </router-link>
                      </IconItem>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col v-if="data.status.lastclusteredfmiid">
                      <IconItem
                        :icon="icon.mdiCounter"
                        title="Last FMI Block"
                      >
                        <router-link
                          :to="{ name: blockRoute,
                                 params: { id: data.status.lastclusteredfmiid }}"
                        >
                          {{ data.status.lastclusteredfmiid.toLocaleString() }}
                        </router-link>
                      </IconItem>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col>
                      <IconItem
                        :icon="icon.mdiCounter"
                        title="RPC Client Block Height"
                      >
                        {{ data.rpcinfo.blocks.toLocaleString() }}
                      </IconItem>
                    </v-col>
                  </v-row>
                </v-col>
              </v-row>
              <named-divider
                v-if="data.rpcinfo"
                title="RPC-Client Data"
              />
              <v-row>
                <v-col
                  cols="12"
                  sm="6"
                  lg="4"
                >
                  <IconItem
                    :icon="icon.mdiCubeOffOutline"
                    title="Pruned"
                    :tooltip="tooltips.rpcPruned"
                  >
                    {{ data.rpcinfo.pruned ? 'Yes' : 'No' }}
                  </IconItem>
                </v-col>
                <v-col
                  cols="12"
                  sm="6"
                  lg="4"
                >
                  <IconItem
                    :icon="icon.mdiWeight"
                    title="Difficulty"
                    :tooltip="tooltips.rpcDifficulty"
                  >
                    {{ data.rpcinfo.difficulty.toLocaleString() }}
                  </IconItem>
                </v-col>
                <v-col
                  v-if="data.rpcinfo.size_on_disk && data.rpcinfo.size_on_disk > 0"
                  cols="12"
                  sm="6"
                  lg="4"
                >
                  <IconItem
                    :icon="icon.mdiHarddisk"
                    title="Blockchain Size"
                    :tooltip="tooltips.rpcBlockchainSize"
                  >
                    {{ (data.rpcinfo.size_on_disk / 1073741824).toLocaleString() }} GiB
                  </IconItem>
                </v-col>
              </v-row>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import {
	mdiRefresh, mdiDatabase, mdiDatabaseSync, mdiDatabaseSearch,
	mdiCounter, mdiProgressWrench, mdiCubeOffOutline, mdiWeight,
	mdiHarddisk,
} from '@mdi/js';
import {PAGE_TITLE, ROUTE_NAME_BLOCK_PAGE} from '@/constants';
import IconItem from './common/IconItem.vue';
import {handleError} from '@/utilities';
import NamedDivider from './common/NamedDivider.vue';
import IconTitle from '@/components/common/IconTitle.vue';

export default {
	name: 'StatusPage',
	components: {IconTitle, NamedDivider, IconItem},
	data() {
		return {
			icon: {
				mdiRefresh,
				mdiDatabase,
				mdiDatabaseSync,
				mdiDatabaseSearch,
				mdiCounter,
				mdiProgressWrench,
				mdiCubeOffOutline,
				mdiWeight,
				mdiHarddisk,
			},
			blockRoute: ROUTE_NAME_BLOCK_PAGE,
			tooltips: {
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
			},
			timer: null,
			refreshStep: 10000,
			data: null,
		};
	},
	computed: {
		crawlerSyncProgress() {
			if (!this.data) {
				return 0.0;
			}

			return this.data.status.lastblockid / this.data.rpcinfo.blocks * 100;
		},
		classifierSyncProgress() {
			if (!this.data) {
				return 0.0;
			}

			const percentage = this.data.status.lastclassifiedid / this.data.status.lastblockid * 100;

			return percentage > 100 ? 100 : percentage;
		},
		clusteringHMISyncProgress() {
			if (!this.data) {
				return 0.0;
			}

			const percentage = this.data.status.lastclusteredhmiid / this.data.status.lastblockid * 100;

			return percentage > 100 ? 100 : percentage;
		},
		clusteringFMISyncProgress() {
			if (!this.data) {
				return 0.0;
			}

			const percentage = this.data.status.lastclusteredfmiid / this.data.status.lastblockid * 100;

			return percentage > 100 ? 100 : percentage;
		},
	},
	created() {
		this.refreshData();
	},
	mounted() {
		document.title = `Status - ${PAGE_TITLE}`;
	},
	beforeUnmount() {
		this.resetTimers();
	},
	methods: {
		startTimer() {
			this.timer = setInterval(async () => {
				await this.loadStatusData();
			}, this.refreshStep);
		},
		resetTimers() {
			clearInterval(this.timer);
		},
		async loadStatusData() {
			try {
				this.data = await this.dakar.meta.metaGet();
				return true;
			} catch (e) {
				handleError(this, e);
				return false;
			}
		},
		async refreshData() {
			this.resetTimers();
			if (await this.loadStatusData()) {
				this.startTimer();
			}
		},
	},
};
</script>

<style scoped>

</style>
