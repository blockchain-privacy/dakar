<template>
  <v-bottom-sheet scrollable v-model="inputVal">
    <v-card style="max-height: 600px">
      <v-card-title>
        <v-icon class="mr-2">{{ icon.mdiChartBar }}</v-icon>
        Heuristic Properties
      </v-card-title>
      <v-divider/>
      <v-card-text style="height: 80%">
        <v-row>
          <v-col>
            <v-card outlined class="mr-auto my-4" max-width="500">
              <v-card-subtitle>
                <v-row>
                  <v-col>
                    <IconItem title="Type" :icon="icon.mdiIframeVariableOutline">
                      {{ heuristicData.heuristicType }}
                    </IconItem>
                  </v-col>
                  <v-col>
                    <IconItem v-if="heuristicData.heuristicParameter"
                              title="Parameter"
                              :icon="icon.mdiTune">
                      {{ heuristicData.heuristicParameter }}
                    </IconItem>
                  </v-col>
                </v-row>
                <v-row v-if="isHollow">
                  <v-col>
                    <v-card-title class="text-h5">
                      Not executed
                    </v-card-title>
                    <v-card-subtitle>
                      This heuristic has not been executed, therefore no results are available.
                    </v-card-subtitle>
                  </v-col>
                </v-row>
                <v-row v-else-if="dataItems.length > 0">
                  <v-col>
                    <IconItem title="Number of origins"
                              :icon="icon.mdiPoundBoxOutline">
                      {{ heuristicData.resultCount ? heuristicData.resultCount : 0 }}
                    </IconItem>
                  </v-col>
                  <v-col>
                    <IconItem title="Number of clusters"
                              :icon="icon.mdiPoundBoxOutline">
                      {{
                        heuristicData.transactions === undefined ? 0
                            : heuristicData.transactions.length
                      }}
                    </IconItem>
                  </v-col>
                </v-row>
                <v-row v-else>
                  <v-col>
                    <v-card-title class="text-h5">
                      No results
                    </v-card-title>
                    <v-card-subtitle>
                      This heuristic returned no results. Try different parameters,
                      other heuristics or a different combination of heuristics.
                    </v-card-subtitle>
                  </v-col>
                </v-row>
              </v-card-subtitle>
            </v-card>
          </v-col>
          <v-col>
            <v-card outlined class="mx-auto my-4" v-if="dataItems.length > 0" min-width="400px">
              <div v-if="enoughDataForGraph" class="text-subtitle-1" style="text-align: center">
                Origin Transactions
              </div>
              <svg id="heuristic_details_canvas" :class="{'hide':!enoughDataForGraph}"/>
              <v-card-title class="text-h5" v-if="!enoughDataForGraph">
                Not enough data to display diagram
              </v-card-title>
              <v-card-subtitle v-if="!enoughDataForGraph && durationInMinutes > 0">
                {{
                  `Only ${durationInMinutes} minute${durationInMinutes > 1 ? 's' : ''}
                between earliest and latest origin.`
                }}
              </v-card-subtitle>
              <v-card-subtitle v-if="!enoughDataForGraph && durationInMinutes === 0">
                All origins occur in the same point of time.
              </v-card-subtitle>
            </v-card>
          </v-col>
          <v-col>
            <v-card outlined class="ml-auto my-4" v-if="dataItems.length > 0">
              <v-data-table :headers="dataHeaders"
                            :items="dataItems"
                            :items-per-page="5"
                            :sort-by.sync="sortBy"
                            :sort-desc.sync="sortDesc"
                            item-key="cluster"
                            show-expand>
                <template v-slot:expanded-item="{ headers, item }">
                  <td :colspan="headers.length" class="py-3">
                    <v-list dense>
                      <v-list-item v-for="a in item.addresses" :key="a"
                                   :to="{ name: addressRoute, params: { id: a }}">
                        <v-list-item-content>
                          <v-list-item-title>{{ a }}</v-list-item-title>
                        </v-list-item-content>
                      </v-list-item>
                    </v-list>
                  </td>
                </template>
              </v-data-table>
            </v-card>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-bottom-sheet>
</template>

<script>
import {
  mdiIframeVariableOutline, mdiTune, mdiPoundBoxOutline, mdiChartBar,
} from '@mdi/js';
import IconItem from '../common/IconItem.vue';
import { ROUTE_NAME_ADDRESS_PAGE } from '../../constants';
import Histogram from '../../d3Documents/histogram';

export default {
  name: 'Details',
  components: { IconItem },
  props: {
    // v-model
    value: { type: Boolean, required: true },
    heuristicData: { type: Object, required: true },
    // array[origins]
    newHeuristicPrefix: { type: String, required: true },
  },
  data() {
    return {
      icon: {
        mdiIframeVariableOutline, mdiTune, mdiPoundBoxOutline, mdiChartBar,
      },
      addressRoute: ROUTE_NAME_ADDRESS_PAGE,
      chart: null,
      sortBy: 'count',
      sortDesc: false,
      svgHistogram: null,
      enoughDataForGraph: true,
      durationInMinutes: 0,
    };
  },
  computed: {
    dataItems() {
      if (!this.heuristicData.transactions) {
        return [];
      }

      let i = 1;
      this.heuristicData.transactions.forEach((v) => {
        v.id = i;
        i += 1;
        v.txCount = v.txs.length;
        // get unique addresses
        const addressSet = new Set();
        v.txs.forEach((d) => addressSet.add(d.addresshash));
        v.addresses = [...addressSet];
        v.address_count = v.addresses.length;
      });

      return this.heuristicData.transactions;
    },
    isHollow() {
      return this.heuristicData.heuristicUid.startsWith(this.newHeuristicPrefix);
    },
    inputVal: {
      get() {
        return this.value;
      },
      set(val) {
        this.$emit('input', val);
      },
    },
    dataHeaders() {
      const idHeader = {
        text: 'ID', align: 'start', sortable: false, value: 'id',
      };
      const addressCountHeader = { text: 'Cluster Address Count', value: 'address_count' };

      const transactionCountHeader = { text: 'Origin Tx Count', value: 'txCount' };
      const destinationHeader = { text: 'Destination Tx Count', value: 'count' };
      const expansionHeader = { value: 'data-table-expand' };

      // check if destination counts from forward lookup are set
      if (this.heuristicData.transactions.some((d) => d.count)) {
        return [idHeader, addressCountHeader, transactionCountHeader,
          destinationHeader, expansionHeader];
      }
      return [idHeader, addressCountHeader, transactionCountHeader, expansionHeader];
    },
  },
  methods: {
    updateData(graphData) {
      // flatten
      const detailArray = [];
      graphData.forEach((d) => {
        detailArray.push(...d.txs);
      });

      this.svgHistogram.draw(detailArray);
      this.enoughDataForGraph = !this.svgHistogram.empty;
      this.durationInMinutes = this.svgHistogram.getDurationInMinutes;
    },
  },
  updated() {
    // do nothing if sheet is not open
    if (!this.value || !this.heuristicData.transactions) return;

    this.svgHistogram.reset();

    this.updateData(this.heuristicData.transactions);
  },
  mounted() {
    this.svgHistogram = new Histogram('heuristic_details_canvas', 600, 300, 'Origin Transactions');
  },
};
</script>

<style>

.bar {
  fill: #008ee5;
}

.hide {
  display: none;
  height: 0;
}

</style>
