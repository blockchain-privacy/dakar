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
                  :value="timeoutData.percent" :indeterminate="timeoutData.updating" rotate="270" size="40">
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
                      <v-icon v-bind:class="{ 'green--text': data.iscrawling, 'red--text': !data.iscrawling }">
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
                        {{ data.iscrawling ? "Active" : "Inactive" }}
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
                        <router-link :to="'search/' + data.lowestblockid">{{ data.lowestblockid }}</router-link>
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
                        <router-link :to="'search/' + data.lastblockid">{{ data.lastblockid }}</router-link>
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-pound</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>
                        Block count
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
                          <span>{{ tooltips.blockCount }}</span>
                        </v-tooltip>
                      </v-list-item-title>
                      <v-list-item-subtitle>
                        {{ data.blkcount }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-pound</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>
                        Transaction count
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
                          <span>{{ tooltips.transactionCount }}</span>
                        </v-tooltip>
                      </v-list-item-title>
                      <v-list-item-subtitle>
                        {{ data.txcount }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-pound</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>
                        Output count
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
                          <span>{{ tooltips.outputCount }}</span>
                        </v-tooltip>
                      </v-list-item-title>
                      <v-list-item-subtitle>
                        {{ data.outputcount }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-pound</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>
                        Address count
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
                          <span>{{ tooltips.addressCount }}</span>
                        </v-tooltip>
                      </v-list-item-title>
                      <v-list-item-subtitle>
                        {{ data.addresscount }}
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
        lastBlockId: "The last block which was completely saved in the database",
        lowestBlockId: "The lowest block ID in the database",
        blockCount: "Number of blocks in the database",
        transactionCount: "Number of transactions in the database",
        outputCount: "Number of outputs in the database. Note that an output is only saved once, even if it is used as an input.",
        addressCount: "Number of addresses in the database",
      },
      timeoutData: {
        start: 0,
        refreshStep: 10000,
        progressStep: 600,
        remaining: 0,
        updating: false,
        percent: 0,
      }
    };
  },
  computed: {
    data() {
      return this.$store.getters.getMetaData;
    },
  },

  methods: {
    startTimer: function () {
      this.startProgressTimer();

      this.timer = setInterval(async () => {
        //this.timeoutData.updating = true;
        clearInterval(this.remainderTimer);
        await this.$store.dispatch('updateMetaData');
        // this.timeoutData.updating = false;
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