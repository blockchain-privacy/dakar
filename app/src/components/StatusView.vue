<template>
  <v-container fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <v-card class="elevation-4">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title>
              <v-icon>{{ icon.mdiDatabase }}</v-icon>
              Server status
            </v-toolbar-title>
            <v-spacer></v-spacer>
            <v-btn icon v-on:click="refreshData">
              <v-progress-circular
                  :value="timeoutData.percent" rotate="270" size="40">
                <v-icon>
                  {{ icon.mdiRefresh }}
                </v-icon>
              </v-progress-circular>
            </v-btn>
          </v-toolbar>
          <v-card-text>
            <v-container v-if="!data">
              <v-skeleton-loader type="table-tbody"></v-skeleton-loader>
            </v-container>
            <v-container v-if="data">
              <v-row>
                <v-col>
                  <IconItem :icon="icon.mdiDatabaseSync" title="Chain Synchronisation"
                            :tooltip="tooltips.databaseSync"
                            is-color :is-red="!data.status.iscrawling">
                    <v-progress-linear
                        :color="crawlerSyncProgress > 98?'green'
                        :crawlerSyncProgress > 90?'light-green':'light-blue'"
                        height="17"
                        :value="crawlerSyncProgress"
                        rounded>
                      {{ Math.round(crawlerSyncProgress) }}%
                    </v-progress-linear>
                  </IconItem>
                </v-col>
                <v-col v-if="data.status.lastclassifiedid > 0">
                  <IconItem :icon="icon.mdiDatabaseSearch" title="Transaction Classification"
                            :tooltip="tooltips.databaseClassification" is-color
                            :is-red="!data.status.isclassifying">
                    <v-progress-linear
                        :color="classifierSyncProgress > 98?'green'
                        :classifierSyncProgress > 90?'light-green':'light-blue'"
                        height="17"
                        :value="classifierSyncProgress"
                        rounded>
                      {{ Math.round(classifierSyncProgress) }}%
                    </v-progress-linear>
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col v-if="data.status.lastclusteredhmiid > 0">
                  <IconItem :icon="icon.mdiDatabaseSearch"
                            title="Hierarchical Multi-Input Clustering"
                            :tooltip="tooltips.databaseClusteringHMI" is-color
                            :is-red="!data.status.isclusteringhmi">
                    <v-progress-linear
                        :color="clusteringHMISyncProgress > 98?'green'
                        :clusteringHMISyncProgress > 90?'light-green':'light-blue'"
                        height="17"
                        :value="clusteringHMISyncProgress"
                        rounded>
                      {{ Math.round(clusteringHMISyncProgress) }}%
                    </v-progress-linear>
                  </IconItem>
                </v-col>
                <v-col v-if="data.status.lastclusteredfmiid > 0">
                  <IconItem :icon="icon.mdiDatabaseSearch" title="Flat Multi-Input Clustering"
                            :tooltip="tooltips.databaseClusteringFMI" is-color
                            :is-red="!data.status.isclusteringfmi">
                    <v-progress-linear
                        :color="clusteringFMISyncProgress > 98?'green'
                        :clusteringFMISyncProgress > 90?'light-green':'light-blue'"
                        height="17"
                        :value="clusteringFMISyncProgress"
                        rounded>
                      {{ Math.round(clusteringFMISyncProgress) }}%
                    </v-progress-linear>
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <IconItem :icon="icon.mdiArrowDownCircleOutline" title="Lowest block ID"
                            :tooltip="tooltips.lowestBlockId">
                    <router-link :to="'block/' + data.status.lowestblockid">
                      {{ data.status.lowestblockid }}
                    </router-link>
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem :icon="icon.mdiTimelineClockOutline" title="Last crawled block"
                            :tooltip="tooltips.lastBlockId">
                    <router-link :to="{ name: blockRoute,
                    params: { id: data.status.lastblockid }}">
                      {{ data.status.lastblockid }}
                    </router-link>
                  </IconItem>
                </v-col>
              </v-row>
              <v-divider v-if="data.rpcinfo"/>
              <v-row>
                <v-col>
                  <IconItem :icon="icon.mdiFormatListNumbered"
                            title="Block Height" :tooltip="tooltips.rpcBlockHeight">
                    {{ data.rpcinfo.blocks }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem :icon="icon.mdiProgressWrench"
                            title="Verification Progress"
                            :tooltip="tooltips.rpcVerificationProgress">
                    <v-progress-linear
                        :color="data.rpcinfo.verificationprogress > 98?'green'
                        :data.rpcinfo.verificationprogress > 90?'light-green':'light-blue'"
                        height="17"
                        :value="data.rpcinfo.verificationprogress"
                        rounded>
                      {{ Math.round(data.rpcinfo.verificationprogress) }}%
                    </v-progress-linear>
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <IconItem
                      :icon="icon.mdiCubeOffOutline"
                      title="Pruned"
                      :tooltip="tooltips.rpcPruned">
                    {{ data.rpcinfo.pruned ? 'Yes' : 'No' }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem
                      :icon="icon.mdiWeight"
                      title="Difficulty"
                      :tooltip="tooltips.rpcDifficulty">
                    {{ data.rpcinfo.difficulty.toFixed() }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-row v-if="data.rpcinfo.size_on_disk && data.rpcinfo.size_on_disk > 0">
                <v-col>
                  <IconItem
                      :icon="icon.mdiHarddisk"
                      title="Blockchain Size"
                      :tooltip="tooltips.rpcBlockchainSize">
                    {{  (data.rpcinfo.size_on_disk / 1073741824).toFixed(2) }} GiB
                  </IconItem>
                </v-col>
              </v-row>
            </v-container>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import {
  mdiRefresh, mdiDatabase, mdiDatabaseSync, mdiDatabaseSearch,
  mdiArrowDownCircleOutline, mdiTimelineClockOutline,
  mdiFormatListNumbered, mdiProgressWrench, mdiCubeOffOutline,
  mdiWeight, mdiHarddisk,
} from '@mdi/js';
import { PAGE_TITLE, ROUTE_NAME_BLOCK_PAGE, ROUTE_META } from '../constants';
import IconItem from './common/IconItem.vue';
import { doGet, handleError } from '../utilities';

export default {
  name: 'StatusView',
  components: { IconItem },
  data() {
    return {
      icon: {
        mdiRefresh,
        mdiDatabase,
        mdiDatabaseSync,
        mdiDatabaseSearch,
        mdiArrowDownCircleOutline,
        mdiTimelineClockOutline,
        mdiFormatListNumbered,
        mdiProgressWrench,
        mdiCubeOffOutline,
        mdiWeight,
        mdiHarddisk,
      },
      blockRoute: ROUTE_NAME_BLOCK_PAGE,
      tooltips: {
        lastBlockId: 'Last block which was completely saved in the database',
        lowestBlockId: 'Lowest block ID in the database',
        databaseSync: 'Percentage of blocks synced from the RPC client to the database. The crawler is active if the icon is green.',
        databaseClassification: 'Percentage of classified blocks in the database. The classifier is active if the icon is green.',
        databaseClusteringHMI: 'Percentage of hierarchical multi-input clustered blocks in the database. '
            + 'Clustering is ongoing if the icon is green.',
        databaseClusteringFMI: 'Percentage of flat multi-input clustered blocks in the database. '
            + 'Clustering is ongoing if the icon is green.',
        rpcBlockHeight: 'Current block height of the RPC client',
        rpcDifficulty: 'Current mining difficulty',
        rpcPruned: 'Whether the RPC client prunes blocks',
        rpcVerificationProgress: 'Estimate of verification progress of the RPC client',
        rpcBlockchainSize: 'The estimated size of the block and undo files on disk',
      },
      timer: null,
      timeoutData: {
        start: 0,
        refreshStep: 10000,
        progressStep: 600,
        remaining: 0,
        percent: 0,
      },
      data: null,
    };
  },
  computed: {
    crawlerSyncProgress() {
      if (!this.data) {
        return 0.0;
      }

      return (1 + (this.data.status.lastblockid - this.data.status.lowestblockid))
          / this.data.rpcinfo.blocks * 100;
    },
    classifierSyncProgress() {
      if (!this.data) {
        return 0.0;
      }
      const percentage = ((1 + (this.data.status.lastclassifiedid - this.data.status.lowestblockid))
          / (1 + (this.data.status.lastblockid - this.data.status.lowestblockid))) * 100;

      return percentage > 100 ? 100 : percentage;
    },
    clusteringHMISyncProgress() {
      if (!this.data) {
        return 0.0;
      }
      const percentage = ((1 + (this.data.status.lastclusteredhmiid
              - this.data.status.lowestblockid))
          / (1 + (this.data.status.lastblockid - this.data.status.lowestblockid))) * 100;

      return percentage > 100 ? 100 : percentage;
    },
    clusteringFMISyncProgress() {
      if (!this.data) {
        return 0.0;
      }
      const percentage = ((1 + (this.data.status.lastclusteredfmiid
              - this.data.status.lowestblockid))
          / (1 + (this.data.status.lastblockid - this.data.status.lowestblockid))) * 100;

      return percentage > 100 ? 100 : percentage;
    },
  },

  methods: {
    startTimer() {
      this.startProgressTimer();

      this.timer = setInterval(async () => {
        clearInterval(this.remainderTimer);
        await this.loadStatusData();
        this.startProgressTimer();
      }, this.timeoutData.refreshStep);
    },
    startProgressTimer() {
      this.timeoutData.percent = 100;
      this.timeoutData.start = new Date().getTime();
      this.remainderTimer = setInterval(this.updateRemainingTime, this.timeoutData.progressStep);
    },
    updateRemainingTime() {
      this.timeoutData.remaining = this.timeoutData.refreshStep
          - (new Date().getTime() - this.timeoutData.start);
      if (this.timeoutData.remaining < 0) {
        this.timeoutData.percent = 0;
        return;
      }
      this.timeoutData.percent = this.timeoutData.remaining / this.timeoutData.refreshStep * 100;
    },
    resetTimers() {
      this.timeoutData.percent = 100;
      clearInterval(this.timer);
      clearInterval(this.remainderTimer);
    },
    loadStatusData() {
      return doGet(ROUTE_META, this.$router, this.$store).then((data) => {
        this.data = data;
        this.$store.dispatch('resetMessages');
      }).catch((e) => {
        handleError(this.$store, e);
        return e;
      });
    },
    async refreshData() {
      this.resetTimers();
      this.loadStatusData().then((err) => {
        if (err !== undefined) {
          return;
        }
        this.startTimer();
      });
    },
  },
  created() {
    this.refreshData();
  },
  mounted() {
    document.title = `Status - ${PAGE_TITLE}`;
  },
  beforeDestroy() {
    this.resetTimers();
  },
};
</script>

<style scoped>

</style>
