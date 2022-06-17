<template>
  <div>
    <v-card flat>
      <v-card-text>
        <v-row v-if="(activities && activities.length > 0) || showTooManyAddressesMsg">
          <v-col>
            <v-select
                multiple
                chips
                deletable-chips
                item-value="id"
                item-text="text"
                label="Filter by Privacy Type"
                v-model="selectedPrivacyLabel"
                style="max-width: 300px; min-width: 200px;"
                :items="privacyLabels"
                @change="updateSvgData(false)">
              <template v-slot:item="{ item }">
                <span>
                   <v-chip label :outlined="item.color === undefined"
                           :color="item.color?item.color:'black'" small/>
                  {{ item.text }}
                </span>
              </template>
            </v-select>
          </v-col>
          <v-col>
            <v-menu
                v-model="menu"
                :close-on-content-click="false"
                :nudge-right="40"
                transition="scale-transition"
                offset-y
                min-width="auto"
                @input="handleMenuChange">
              <template v-slot:activator="{ on, attrs }">
                <v-text-field
                    style="min-width: 100px"
                    :value="dateRangeString"
                    label="Filter by Date"
                    :prepend-icon="icons.mdiCalendarRange"
                    readonly
                    v-bind="attrs"
                    v-on="on"/>
              </template>
              <v-date-picker
                  scrollable
                  color="primary"
                  first-day-of-week="1"
                  no-title
                  range
                  :events="datePicker.events"
                  :min="datePicker.min"
                  :max="datePicker.max"
                  v-model="datePicker.range"/>
            </v-menu>
          </v-col>
          <v-col>
            <v-switch
                label="Include cluster addresses"
                v-model="includeCusterAddresses"
                @click="updateSvgData(true)"/>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
    <v-card class="my-3" flat v-if="hasLoaded && showEmptyResponseMsg && !isLoading">
      <v-card-text class="text-h6" style="text-align:center">
        No mixing activity available
      </v-card-text>
    </v-card>
    <v-card class="my-3" flat v-if="hasLoaded && showTooManyAddressesMsg && !isLoading">
      <v-card-text class="text-h6" style="text-align:center">
        Mixing activity lookup is not possible because the cluster
        of this address is connected to too many addresses
      </v-card-text>
    </v-card>
    <v-card class="my-3" flat v-show="this.activities && this.activities.length > 0">
      <v-card-text>
        <v-tabs v-model="graphTabs" grow>
          <v-tab key="histogram" @change="onTabChange('histogram')">Histogram</v-tab>
          <v-tab key="graph" @change="onTabChange('graph')">
            Force Graph
          </v-tab>
        </v-tabs>
        <v-tabs-items v-model="graphTabs">
          <v-tab-item eager key="histogram">
            <v-card flat>
              <v-card-text>
                <p v-if="showNotEnoughDataMsg && !isLoading"
                   class="text-h6" style="text-align: center">
                  Not enough data available to draw chart
                </p>
                <div v-if="showHistogram" class="text-subtitle-1" style="text-align: center">
                  {{
                    selectedPrivacyLabel.length === 0 ? 'All Privacy'
                        : selectedPrivacyLabel.map(capitalize).join(', ')
                  }}
                  Transactions
                </div>
                <svg id="mixing_activity_histogram" v-show="showHistogram"/>
                <v-card outlined v-if="showHistogram">
                  <v-card-text>
                    <v-row>
                      <v-col>
                        <div class="text-subtitle-1 font-weight-black">Legend</div>
                      </v-col>
                      <v-col v-for="item in privacyLabels" :key="item.text">
                        <div class="legendBox">
                          <v-chip label :outlined="item.color === undefined"
                                  :color="item.color?item.color:'black'" small
                                  class="mr-1"/>
                          {{ item.text }}
                        </div>
                      </v-col>
                    </v-row>
                  </v-card-text>
                </v-card>
                <v-expand-transition>
                  <v-card outlined v-if="barTable.transactions.length > 0" class="mt-2">
                    <v-card-text>
                      <v-data-table
                          :headers="barTable.headers"
                          :items="barTable.transactions">
                        <template v-slot:top>
                          <v-toolbar flat class="hidden-sm-and-up">
                            <v-toolbar-title>
                              Privacy Transactions from {{ barTable.transactions.startDate }}
                              to {{ barTable.transactions.endDate }}
                            </v-toolbar-title>
                          </v-toolbar>
                          <v-toolbar flat class="hidden-xs-only">
                            <v-toolbar-title>
                              Privacy Transactions from {{ barTable.transactions.startDate }}
                              to {{ barTable.transactions.endDate }}
                            </v-toolbar-title>
                          </v-toolbar>
                        </template>
                        <template v-slot:[`item.txhash`]="{ item }">
                          <router-link :to="{ name: txRoute, params: { id: item.txhash }}">
                            {{ item.txhash }}
                          </router-link>
                        </template>
                        <template v-slot:[`item.dateTime`]="{ item }">
                          <span>{{ item.dateTime.toLocaleString() }}</span>
                        </template>
                      </v-data-table>
                    </v-card-text>
                  </v-card>
                </v-expand-transition>
              </v-card-text>
            </v-card>
          </v-tab-item>
          <v-tab-item eager key="graph">
            <v-card flat>
              <v-card-text>
                <div v-if="!overrideTooManyTransactionsWarning && showTooManyTransactionsMsg"
                     style="text-align:center">
                  <v-alert text
                           prominent type="warning">
                    The mixing activity results have more than
                    {{ tooManyTransactionsThreshold }} transactions.
                    Displaying a large number of items in a force graph may severely degrade
                    the performance of your browser.
                    Consider filtering the results by time or privacy type.
                  </v-alert>
                  <v-btn color="primary" @click="showForceGraphDespiteWarning">
                    Display force graph anyway
                  </v-btn>
                </div>

                <p v-if="!showGraph && !isLoading && !showTooManyTransactionsMsg"
                   class="text-h6" style="text-align: center">
                  No data available
                </p>
                <svg id="mixing_activity_force_graph" v-show="showGraph"/>
                <v-card outlined v-if="showGraph">
                  <v-card-text>
                    <v-row>
                      <v-col>
                        <div class="text-subtitle-1 font-weight-black">Legend</div>
                      </v-col>
                      <v-col v-for="item in privacyLabels" :key="item.text">
                        <div class="legendBox">
                          <v-chip label :outlined="item.color === undefined"
                                  :color="item.color?item.color:'black'" small
                                  class="mr-1"/>
                          {{ item.text }}
                        </div>
                      </v-col>
                    </v-row>
                  </v-card-text>
                </v-card>
                <v-card outlined v-if="this.clickedNode" class="mt-2">
                  <v-card-title>
                    Transaction
                    <router-link class="ml-1"
                                 :to="{ name: txRoute, params: { id: clickedNode.txhash }}">
                      {{ clickedNode.txhash }}
                    </router-link>
                  </v-card-title>
                  <v-card-text>
                    <p class="text-subtitle-1">
                      Privacy Type: {{ clickedNode.privacytype }}
                    </p>
                    <p class="text-subtitle-1">
                      Timestamp: {{ clickedNode.dateTime.toLocaleString() }}
                    </p>
                    <p class="text-subtitle-1" v-if="clickedNode.input_txs">
                      Input Transactions:
                    </p>
                    <v-expand-transition>
                      <v-list v-if="clickedNode.input_txs">
                        <v-list-item v-for="(t) in clickedNode.input_txs" :key="t.txhash"
                                     :to="{ name: txRoute, params: { id: t.txhash }}">
                          <v-list-item-content>
                            <v-list-item-title>{{ t.txhash }}</v-list-item-title>
                          </v-list-item-content>
                        </v-list-item>
                      </v-list>
                    </v-expand-transition>
                  </v-card-text>
                </v-card>
              </v-card-text>
            </v-card>
          </v-tab-item>
        </v-tabs-items>
      </v-card-text>
    </v-card>
    <v-progress-linear class="mt-10" v-if="isLoading" indeterminate/>
  </div>
</template>

<script>
import { mdiCalendarRange } from '@mdi/js';
import Histogram from '../../../d3Documents/histogram';
import ForceGraph from '../../../d3Documents/forceGraph';
import {
  doPost, getPrivacyTypeLabel, handleError,
} from '../../../utilities';
import {
  ROUTE_MIXING_ACTIVITY,
  ROUTE_NAME_TRANSACTION_PAGE,
} from '../../../constants';

// capitalize returns the first letter of each word (separated by a space) in str capitalized
function capitalize(str) {
  return str.split(' ').map((d) => d[0].toUpperCase() + d.slice(1)).join(' ');
}

export default {
  name: 'MixingActivity',
  props: {
    addressHash: { type: String, required: true },
  },
  data() {
    return {
      icons: { mdiCalendarRange },
      txRoute: ROUTE_NAME_TRANSACTION_PAGE,
      lastQuery: '',
      includeCluster: false,
      selectedPrivacyLabel: [],
      showHistogram: false,
      showGraph: false,
      colorMap: new Map(),
      svgHistogram: null,
      svgGraph: null,
      isLoading: false,
      showEmptyResponseMsg: false,
      showTooManyAddressesMsg: false,
      showNotEnoughDataMsg: false,
      showTooManyTransactionsMsg: false,
      overrideTooManyTransactionsWarning: false,
      tooManyTransactionsThreshold: 500,
      activities: null,
      initialLoadDone: false,
      includeCusterAddresses: false,
      datePicker: {
        range: null,
        oldRange: null,
        min: '',
        max: '',
        events: [],
      },
      menu: null,
      graphTabs: null,
      barTable: {
        headers: [{
          text: 'Transaction', align: 'start', value: 'txhash',
        },
        { text: 'Timestamp', value: 'dateTime' },
        { text: 'Transaction Type', value: 'privacytype' },
        ],
        transactions: [],
      },
      clickedNode: null,
      graphMode: false,
      hasLoaded: false,
    };
  },
  computed: {
    privacyLabels() {
      const labels = [];

      this.colorMap.forEach((v, k) => {
        labels.push({ text: capitalize(k), color: v, id: k });
      });

      return labels;
    },
    dateRangeString() {
      let dateString = '';

      if (!this.datePicker.range) {
        return dateString;
      }

      dateString = new Date(this.datePicker.range[0]).toLocaleDateString();

      if (this.datePicker.range.length === 2) {
        dateString += ` to ${new Date(this.datePicker.range[1]).toLocaleDateString()}`;
      }

      return dateString;
    },
  },
  methods: {
    capitalize,
    setErrorMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'error', temporary: true });
    },
    onBarClick(data) {
      if (data.x0.getHours() === data.x1.getHours()
          && data.x0.getMinutes() === data.x1.getMinutes()) {
        data.startDate = data.x0.toLocaleDateString();
        data.endDate = data.x1.toLocaleDateString();
      } else {
        data.startDate = data.x0.toLocaleString();
        data.endDate = data.x1.toLocaleString();
      }

      this.barTable.transactions = data;
    },
    onNodeClick(data) {
      this.clickedNode = data;
    },
    handleMenuChange(open) {
      if (open) {
        this.datePicker.oldRange = this.datePicker.range;
      } else if (this.datePicker.oldRange !== this.datePicker.range) {
        this.updateSvgData(false);
      }
    },
    getMixingActivity() {
      this.lastQuery = this.addressHash;
      return doPost(ROUTE_MIXING_ACTIVITY, this.$router, this.$store,
        { addressHash: this.addressHash, isClusterLookup: this.includeCusterAddresses })
        .then((data) => data)
        .catch((e) => {
          handleError(this.$store, e);
          return [];
        });
    },
    getFilteredData(withLinks) {
      let fromDate = null;
      let toDate = null;
      let considerDate = true;

      if (this.datePicker.range !== null) {
        fromDate = new Date(this.datePicker.range[0]);
        if (this.datePicker.range.length === 2) {
          toDate = new Date(this.datePicker.range[1]);
        } else {
          toDate = new Date(this.datePicker.range[0]);
        }

        // check if dates are in the right order
        if (toDate < fromDate) {
          const mem = toDate;
          toDate = fromDate;
          fromDate = mem;
        }

        // increase to toDate by one day, so the filter includes
        // all events on that day and does cut off the day before
        toDate.setDate(toDate.getDate() + 1);
      } else {
        considerDate = false;
      }

      const ret = { items: [], links: [] };

      ret.items = this.activities.filter((d) => {
        if (this.selectedPrivacyLabel.length > 0
            && !this.selectedPrivacyLabel.includes(d.privacytype)) return false;

        if (!considerDate) return true;

        return d.dateTime <= toDate && d.dateTime >= fromDate;
      });

      if (withLinks) {
        const filteredHashes = new Set(ret.items.map((d) => d.txhash));

        ret.items.forEach((d) => {
          if (d.input_txs === undefined || d.input_txs.length === 0) return;

          d.input_txs.forEach((it) => {
            if (!filteredHashes.has(it.txhash)) return;
            ret.links.push({ source: it.txhash, target: d.txhash });
          });
        });
      }

      return ret;
    },
    getCategories(filtered) {
      if (this.selectedPrivacyLabel.length > 0) {
        return this.selectedPrivacyLabel;
      }

      return [...new Set(filtered.map((d) => d.privacytype))];
    },
    showForceGraphDespiteWarning() {
      this.overrideTooManyTransactionsWarning = true;
      this.onTabChange('graph');
    },
    onTabChange(tab) {
      this.graphMode = tab === 'graph';

      if (this.graphMode) {
        if (!this.showTooManyTransactionsMsg
            || this.overrideTooManyTransactionsWarning) {
          this.updateSvgData();
        }
        return;
      }

      this.updateSvgData();
    },
    async updateSvgData(pullNewData) {
      this.showHistogram = false;
      this.showGraph = false;
      this.isLoading = true;
      this.showTooManyAddressesMsg = false;

      // check if new data has to be loaded
      if (pullNewData || !this.initialLoadDone) {
        const mixingActivity = await this.getMixingActivity();
        this.hasLoaded = true;

        if (!mixingActivity.success) {
          this.isLoading = false;
          this.activities = [];
          this.setErrorMessage('error getting mixing activity');
          return;
        }

        if (mixingActivity.msg && mixingActivity.msg === 'too_many_addresses') {
          this.showTooManyAddressesMsg = true;
          this.initialLoadDone = true;
          this.isLoading = false;
          this.activities = [];
          return;
        }

        if (mixingActivity.activities === undefined) {
          this.showEmptyResponseMsg = true;
          this.isLoading = false;
          this.activities = [];
          return;
        }

        // used to set boundaries for the date picker
        let maxDate = null;
        let minDate = null;

        // reset events
        const events = [];

        this.activities = mixingActivity.activities.map((d) => {
          d.privacytype = getPrivacyTypeLabel(d.privacytype);
          d.dateTime = new Date(d.block[0].ts);

          events.push(d.dateTime.toISOString().substring(0, 10));

          if (maxDate === null || d.dateTime > maxDate) maxDate = new Date(d.dateTime);
          if (minDate === null || d.dateTime < minDate) minDate = new Date(d.dateTime);
          return d;
        });

        this.datePicker.events = events;

        // date picker needs iso strings
        // cut off time portion, so it gets displayed prettier
        this.datePicker.min = minDate.toISOString().substring(0, 10);
        this.datePicker.max = maxDate.toISOString().substring(0, 10);

        this.datePicker.range = [this.datePicker.min, this.datePicker.max];

        this.initialLoadDone = true;
      }

      this.showEmptyResponseMsg = false;

      const filtered = this.getFilteredData(this.graphMode);

      if (filtered.items.length === 0) {
        this.isLoading = false;
        this.showGraph = false;
        this.showHistogram = false;
        this.showNotEnoughDataMsg = true;
        return;
      }

      this.showTooManyTransactionsMsg = filtered.items.length > this.tooManyTransactionsThreshold;

      // draw
      if (this.graphMode) {
        if (!this.showTooManyTransactionsMsg
              || this.overrideTooManyTransactionsWarning) {
          this.svgGraph.draw(filtered.items, filtered.links);
          this.showGraph = true;
        }
      } else {
        this.svgHistogram.reset();
        this.svgHistogram.drawStacked(filtered.items,
          this.getCategories(filtered.items), this.colorMap);
        this.showHistogram = !this.svgHistogram.empty;
        this.showNotEnoughDataMsg = this.svgHistogram.empty;
      }

      this.isLoading = false;
    },
  },
  beforeMount() {
    this.includeCusterAddresses = this.includeCluster;
    this.colorMap.set('destination', '#0072B2');
    this.colorMap.set('collateral creation', '#E69F00');
    this.colorMap.set('collateral payment', '#009E73');
    this.colorMap.set('origin', '#D55E00');
    this.colorMap.set('mixing', '#56B4E9');

    this.svgHistogram = new Histogram('mixing_activity_histogram',
      1200, 300, 'Privacy transactions');
    this.svgHistogram.setClickHandler(this.onBarClick);
  },
  mounted() {
    // has to be called after the SVG is included in the DOM
    this.svgGraph = new ForceGraph(1200, 500, 'mixing_activity_force_graph', this.colorMap);
    this.svgGraph.setClickHandler(this.onNodeClick);
  },
  created() {
    this.updateSvgData(true);
  },
  watch: {
    addressHash() {
      // prop was changed -> pull new data
      this.updateSvgData(true);
    },
  },
};
</script>

<style scoped>
>>> .overlay {
  stroke-width: 2px;
  stroke: #1976D2;
  fill: #1976D2;
  cursor: pointer;
}

>>> .nodeMouseOver {
  cursor: pointer;
}

.legendBox {
  display: flex;
  align-items: center;
  justify-content: center;
  word-wrap: unset;
  word-break: unset;
  white-space: nowrap
}
</style>
