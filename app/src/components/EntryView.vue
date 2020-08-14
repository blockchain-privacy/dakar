<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <v-card class="elevation-12">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title>
              <v-icon>mdi-database</v-icon>
              Database Information
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
                  <IconItem icon="mdi-robot" title="Crawler status"
                            :tooltip="tooltips.crawler" is-color :is-red="!data.status.iscrawling">
                    {{ data.status.iscrawling ? "Active" : "Inactive" }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem icon="mdi-database-sync" title="Database synchronisation" :tooltip="tooltips.databaseSync">
                    <v-progress-linear
                        :color="databaseSyncProgress > 98?'green':databaseSyncProgress > 90?'light-green':'light-blue'"
                        height="17"
                        :value="databaseSyncProgress"
                        rounded>
                      {{ Math.round(databaseSyncProgress) }} %
                    </v-progress-linear>
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <IconItem icon="mdi-arrow-down-circle-outline" title="Lowest block ID"
                            :tooltip="tooltips.lowestBlockId">
                    <router-link :to="'search/' + data.status.lowestblockid">
                      {{ data.status.lowestblockid }}
                    </router-link>
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem icon="mdi-timeline-clock-outline" title="Last fully crawled block"
                            :tooltip="tooltips.lastBlockId">
                    <router-link :to="'search/' + data.status.lastblockid">
                      {{ data.status.lastblockid }}
                    </router-link>
                  </IconItem>
                </v-col>
              </v-row>
              <v-divider v-if="data.rpcinfo"></v-divider>
              <v-row>
                <v-col>
                  <IconItem icon="mdi-message-text-clock-outline" title="RPC Version" :tooltip="tooltips.rpcVersion">
                    {{ data.rpcinfo.version }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem icon="mdi-file-clock-outline" title="Protocol version"
                            :tooltip="tooltips.rpcProtocolVersion">
                    {{ data.rpcinfo.protocolversion }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <IconItem icon="mdi-format-list-numbered" title="Block Height" :tooltip="tooltips.rpcBlockHeight">
                    {{ data.rpcinfo.blocks }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem icon="mdi-weight" title="Difficulty" :tooltip="tooltips.rpcDifficulty">
                    {{ data.rpcinfo.difficulty.toFixed() }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <IconItem icon="mdi-lan" title="Connections" :tooltip="tooltips.rpcConnections">
                    {{ data.rpcinfo.connections }}
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
import {PAGE_TITLE} from "@/constants";
import IconItem from "@/components/common/IconItem";

export default {
  name: 'EntryView',
  components: {IconItem},
  data: function () {
    return {
      tooltips: {
        crawler: "Displays if the crawler is currently active",
        lastBlockId: "Last block which was completely saved in the database",
        lowestBlockId: "Lowest block ID in the database",
        databaseSync: "Percentage of available blocks included in database",
        rpcVersion: "Version of the RPC client",
        rpcProtocolVersion: "Version of the protocol",
        rpcBlockHeight: "Current block height of the RPC client",
        rpcConnections: "Number of Nodes connected to the RPC client",
        rpcDifficulty: "Current mining difficulty",
      },
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
    databaseSyncProgress() {
      if (!this.data) {
        return 0.0;
      }

      return (1 + (this.data.status.lastblockid - this.data.status.lowestblockid)) / this.data.rpcinfo.blocks * 100;
    }
  },

  methods: {
    startTimer: function () {
      this.startProgressTimer();

      this.timer = setInterval(async () => {
        clearInterval(this.remainderTimer);
        await this.$store.dispatch('updateMetaData');
        this.startProgressTimer();
      }, this.timeoutData.refreshStep);
    },
    startProgressTimer: function () {
      this.timeoutData.percent = 100;
      this.timeoutData.start = new Date().getTime();
      this.remainderTimer = setInterval(this.updateRemainingTime, this.timeoutData.progressStep);
    },
    updateRemainingTime: function () {
      this.timeoutData.remaining = this.timeoutData.refreshStep - (new Date().getTime() - this.timeoutData.start);
      if (this.timeoutData.remaining < 0) {
        this.timeoutData.percent = 0;
        return
      }
      this.timeoutData.percent = this.timeoutData.remaining / this.timeoutData.refreshStep * 100;
    },
    resetTimers: function () {
      this.timeoutData.percent = 100;
      clearInterval(this.timer);
      clearInterval(this.remainderTimer);
    },
    refreshData: function () {
      this.resetTimers();
      this.$store.dispatch('updateMetaData');
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
  }
}
</script>

<style scoped>

</style>