<template>
  <v-bottom-sheet scrollable v-model="inputVal">
    <v-card style="max-height: 800px">
      <v-card-title>
        <v-icon class="mr-2">{{ icon.mdiChartBar }}</v-icon>
        Heuristic Properties
      </v-card-title>
      <v-divider/>
      <v-card-text style="height: 80%">
        <v-row>
          <!-- side bar -->
          <v-col cols="12" sm="12" md="4" lg="4">
            <!-- properties -->
            <v-row>
              <v-col>
                <v-card outlined class="mr-auto my-4" max-width="500">
                  <v-card-subtitle>
                    <v-row>
                      <v-col>
                        <IconItem title="Type" :icon="icon.mdiIframeVariableOutline">
                          {{ heuristicData.heuristicTypeTitle }}
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
                    <v-row>
                      <v-col>
                        <IconItem title="Custom clusters" :icon="icon.mdiMerge">
                          {{ heuristicData.heuristicCustomClusters ? 'yes' : 'no' }}
                        </IconItem>
                      </v-col>
                      <v-col>
                        <IconItem title="Exclude Addresses" :icon="icon.mdiPlaylistRemove">
                          {{ heuristicData.heuristicExcludeAddresses ? 'yes' : 'no' }}
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
                          {{ transactionCount }}
                        </IconItem>
                      </v-col>
                      <v-col>
                        <IconItem title="Number of clusters"
                                  :icon="icon.mdiPoundBoxOutline">
                          {{ heuristicData.clusterCount ? heuristicData.clusterCount : 0 }}
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
            </v-row>
            <!-- timeline -->
            <v-row>
              <v-col>
                <v-card outlined class="mr-auto my-4" v-if="dataItems.length > 0"
                        min-width="400px" max-width="600px">
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
            </v-row>
          </v-col>
          <!-- iterator view -->
          <v-col v-if="dataItems">
            <results :items="dataItems"/>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-bottom-sheet>
</template>

<script>
import {
  mdiIframeVariableOutline, mdiTune, mdiPoundBoxOutline, mdiChartBar, mdiMerge, mdiPlaylistRemove,
} from '@mdi/js';
import IconItem from '../common/IconItem.vue';
import Histogram from '../../d3Documents/histogram';
import Results from './Results.vue';

export default {
  name: 'Details',
  components: { Results, IconItem },
  props: {
    // v-model
    value: { type: Boolean, required: true },
    heuristicData: { type: Object, required: true },
    newHeuristicPrefix: { type: String, required: true },
  },
  data() {
    return {
      icon: {
        mdiIframeVariableOutline, mdiTune, mdiPoundBoxOutline, mdiChartBar, mdiMerge, mdiPlaylistRemove,
      },
      chart: null,
      svgHistogram: null,
      enoughDataForGraph: true,
      durationInMinutes: 0,
    };
  },
  computed: {
    isHollow() {
      return this.heuristicData.heuristicUid.startsWith(this.newHeuristicPrefix);
    },
    transactionCount() {
      if (!this.heuristicData.clusters) return 0;
      let count = 0;

      this.heuristicData.clusters.forEach((cluster) => {
        count += cluster.txs.length;
      });
      return count;
    },
    dataItems() {
      if (!this.heuristicData.clusters) return [];
      return this.heuristicData.clusters;
    },
    inputVal: {
      get() {
        return this.value;
      },
      set(val) {
        this.$emit('input', val);
      },
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
    if (!this.value || !this.heuristicData.clusters) return;

    this.svgHistogram.reset();

    this.updateData(this.heuristicData.clusters);
  },
  mounted() {
    this.svgHistogram = new Histogram('heuristic_details_canvas', 600, 300, 'Origin Transactions');
  },
};
</script>

<style>

/* css for d3 graph  */
.bar {
  fill: #008ee5;
}

.hide {
  display: none;
  height: 0;
}

</style>
