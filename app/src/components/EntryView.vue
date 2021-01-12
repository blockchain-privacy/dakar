<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <v-card class="elevation-12">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title>
              <v-icon>mdi-database</v-icon>
              Server status
            </v-toolbar-title>
            <v-spacer></v-spacer>
            <v-btn icon v-on:click="refreshData">
              <v-progress-circular
                  :value="timeoutData.percent" rotate="270" size="40">
                <v-icon>
                  mdi-refresh
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
                  <IconItem icon="mdi-database-sync" title="Chain Synchronisation"
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
                <v-col>
                  <IconItem icon="mdi-database-search" title="Database analysis"
                            :tooltip="tooltips.databaseAnalysation" is-color
                            :is-red="!data.status.isanalyzing">
                    <v-progress-linear
                        :color="analyzerSyncProgress > 98?'green'
                        :analyzerSyncProgress > 90?'light-green':'light-blue'"
                        height="17"
                        :value="analyzerSyncProgress"
                        rounded>
                      {{ Math.round(analyzerSyncProgress) }}%
                    </v-progress-linear>
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <IconItem icon="mdi-arrow-down-circle-outline" title="Lowest block ID"
                            :tooltip="tooltips.lowestBlockId">
                    <router-link :to="'block/' + data.status.lowestblockid">
                      {{ data.status.lowestblockid }}
                    </router-link>
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem icon="mdi-timeline-clock-outline" title="Last crawled block"
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
                  <IconItem icon="mdi-format-list-numbered"
                            title="Block Height" :tooltip="tooltips.rpcBlockHeight">
                    {{ data.rpcinfo.blocks }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem icon="mdi-progress-wrench"
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
                      icon="mdi-cube-off-outline"
                      title="Pruned"
                      :tooltip="tooltips.rpcPruned">
                    {{ data.rpcinfo.pruned ? 'Yes' : 'No' }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem icon="mdi-weight" title="Difficulty" :tooltip="tooltips.rpcDifficulty">
                    {{ data.rpcinfo.difficulty.toFixed() }}
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
import { PAGE_TITLE, ROUTE_NAME_BLOCK_PAGE } from '../constants';
import IconItem from './common/IconItem.vue';

export default {
  name: 'EntryView',
  components: { IconItem },
  data() {
    return {
      blockRoute: ROUTE_NAME_BLOCK_PAGE,
      tooltips: {
        lastBlockId: 'Last block which was completely saved in the database',
        lowestBlockId: 'Lowest block ID in the database',
        databaseSync: 'Percentage of blocks synced from the RPC client to the database. The crawler is active if the icon is green.',
        databaseAnalysation: 'Percentage of analyzed blocks in the database. The analyzer is active if the icon is green.',
        rpcBlockHeight: 'Current block height of the RPC client',
        rpcDifficulty: 'Current mining difficulty',
        rpcPruned: 'Whether the RPC client prunes blocks',
        rpcVerificationProgress: 'Estimate of verification progress of the RPC client',
      },
      timer: null,
      timeoutData: {
        start: 0,
        refreshStep: 10000,
        progressStep: 600,
        remaining: 0,
        percent: 0,
      },
    };
  },
  computed: {
    data() {
      return this.$store.getters.getMetaData;
    },
    crawlerSyncProgress() {
      if (!this.data) {
        return 0.0;
      }

      return (1 + (this.data.status.lastblockid - this.data.status.lowestblockid))
          / this.data.rpcinfo.blocks * 100;
    },
    analyzerSyncProgress() {
      if (!this.data) {
        return 0.0;
      }
      const percentage = ((1 + (this.data.status.lastanalysedid - this.data.status.lowestblockid))
          / (1 + (this.data.status.lastblockid - this.data.status.lowestblockid))) * 100;

      return percentage > 100 ? 100 : percentage;
    },
  },

  methods: {
    startTimer() {
      this.startProgressTimer();

      this.timer = setInterval(async () => {
        clearInterval(this.remainderTimer);
        await this.$store.dispatch('updateMetaData');
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
    async refreshData() {
      this.resetTimers();
      await this.$store.dispatch('updateMetaData');
      this.startTimer();
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
