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
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon
                          v-bind:class="{ 'green--text': data.status.iscrawling, 'red--text': !data.status.iscrawling }">
                        mdi-robot
                      </v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>
                        Crawler status
                        <v-tooltip right>
                          <template v-slot:activator="{ on }">
                            <v-hover
                                v-slot:default="{ hover }"
                                open-delay="0">
                              <v-icon small v-on="on">
                                {{ hover ? "mdi-help-circle" : "mdi-help-circle-outline" }}
                              </v-icon>
                            </v-hover>
                          </template>
                          <span>{{ tooltips.crawler }}</span>
                        </v-tooltip>
                      </v-list-item-title>
                      <v-list-item-subtitle>
                        {{ data.status.iscrawling ? "Active" : "Inactive" }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>
                        mdi-database-sync
                      </v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>
                        Database synchronisation
                        <v-tooltip right>
                          <template v-slot:activator="{ on }">
                            <v-hover
                                v-slot:default="{ hover }"
                                open-delay="0">
                              <v-icon small v-on="on">
                                {{ hover ? "mdi-help-circle" : "mdi-help-circle-outline" }}
                              </v-icon>
                            </v-hover>
                          </template>
                          <span>{{ tooltips.databaseSync }}</span>
                        </v-tooltip>
                      </v-list-item-title>
                      <v-list-item-subtitle>
                        <v-progress-linear
                            :color="databaseSyncProgress > 98?'green':databaseSyncProgress > 90?'light-green':'light-blue'"
                            height="17"
                            :value="databaseSyncProgress"
                            rounded>
                          {{ Math.round(databaseSyncProgress) }} %
                        </v-progress-linear>
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-arrow-down-circle-outline</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>
                        Lowest block ID
                        <v-tooltip right>
                          <template v-slot:activator="{ on }">
                            <v-hover
                                v-slot:default="{ hover }"
                                open-delay="0">
                              <v-icon small v-on="on">
                                {{ hover ? "mdi-help-circle" : "mdi-help-circle-outline" }}
                              </v-icon>
                            </v-hover>
                          </template>
                          <span>{{ tooltips.lowestBlockId }}</span>
                        </v-tooltip>
                      </v-list-item-title>
                      <v-list-item-subtitle>
                        <router-link :to="'search/' + data.status.lowestblockid">
                          {{ data.status.lowestblockid }}
                        </router-link>
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-timeline-clock-outline</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>
                        Last fully crawled block
                        <v-tooltip right>
                          <template v-slot:activator="{ on }">
                            <v-hover
                                v-slot:default="{ hover }"
                                open-delay="0">
                              <v-icon small v-on="on">
                                {{ hover ? "mdi-help-circle" : "mdi-help-circle-outline" }}
                              </v-icon>
                            </v-hover>
                          </template>
                          <span>{{ tooltips.lastBlockId }}</span>
                        </v-tooltip>
                      </v-list-item-title>
                      <v-list-item-subtitle>
                        <router-link :to="'search/' + data.status.lastblockid">
                          {{ data.status.lastblockid }}
                        </router-link>
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
              </v-row>
              <v-divider v-if="data.rpcinfo"></v-divider>
              <v-row>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-message-text-clock-outline</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>
                        RPC Version
                        <v-tooltip right>
                          <template v-slot:activator="{ on }">
                            <v-hover
                                v-slot:default="{ hover }"
                                open-delay="0">
                              <v-icon small v-on="on">
                                {{ hover ? "mdi-help-circle" : "mdi-help-circle-outline" }}
                              </v-icon>
                            </v-hover>
                          </template>
                          <span>{{ tooltips.rpcVersion }}</span>
                        </v-tooltip>
                      </v-list-item-title>
                      <v-list-item-subtitle>
                        {{ data.rpcinfo.version }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-file-clock-outline</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>
                        Protocol version
                        <v-tooltip right>
                          <template v-slot:activator="{ on }">
                            <v-hover
                                v-slot:default="{ hover }"
                                open-delay="0">
                              <v-icon small v-on="on">
                                {{ hover ? "mdi-help-circle" : "mdi-help-circle-outline" }}
                              </v-icon>
                            </v-hover>
                          </template>
                          <span>{{ tooltips.rpcProtocolVersion }}</span>
                        </v-tooltip>
                      </v-list-item-title>
                      <v-list-item-subtitle>
                        {{ data.rpcinfo.protocolversion }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-format-list-numbered</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>
                        Block Height
                        <v-tooltip right>
                          <template v-slot:activator="{ on }">
                            <v-hover
                                v-slot:default="{ hover }"
                                open-delay="0">
                              <v-icon small v-on="on">
                                {{ hover ? "mdi-help-circle" : "mdi-help-circle-outline" }}
                              </v-icon>
                            </v-hover>
                          </template>
                          <span>{{ tooltips.rpcBlockHeight }}</span>
                        </v-tooltip>
                      </v-list-item-title>
                      <v-list-item-subtitle>
                        {{ data.rpcinfo.blocks }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-weight</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>
                        Difficulty
                        <v-tooltip right>
                          <template v-slot:activator="{ on }">
                            <v-hover
                                v-slot:default="{ hover }"
                                open-delay="0">
                              <v-icon small v-on="on">
                                {{ hover ? "mdi-help-circle" : "mdi-help-circle-outline" }}
                              </v-icon>
                            </v-hover>
                          </template>
                          <span>{{ tooltips.rpcDifficulty }}</span>
                        </v-tooltip>
                      </v-list-item-title>
                      <v-list-item-subtitle>
                        {{ Math.trunc(data.rpcinfo.difficulty) }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-lan</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>
                        Connections
                        <v-tooltip right>
                          <template v-slot:activator="{ on }">
                            <v-hover
                                v-slot:default="{ hover }"
                                open-delay="0">
                              <v-icon small v-on="on">
                                {{ hover ? "mdi-help-circle" : "mdi-help-circle-outline" }}
                              </v-icon>
                            </v-hover>
                          </template>
                          <span>{{ tooltips.rpcConnections }}</span>
                        </v-tooltip>
                      </v-list-item-title>
                      <v-list-item-subtitle>
                        {{ data.rpcinfo.connections }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
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

export default {
  name: 'EntryView',
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