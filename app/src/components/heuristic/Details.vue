<template>
  <v-bottom-sheet scrollable v-model="inputVal">
    <v-card style="max-height: 500px">
      <v-card-text style="height: 80%">
        <div class="d-flex flex-wrap" style="align-items: flex-start;">
          <v-card outlined
                  class="mx-auto my-12"
                  max-width="500">
            <v-card-title>
              Heuristic Properties
            </v-card-title>
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
                    This heuristic has not been executed, thus no results are available.
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
                  <IconItem title="Number of addresses"
                            :icon="icon.mdiPoundBoxOutline">
                    {{ addressMap === undefined ? 0 : addressMap.size }}
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
          <v-card outlined class="mx-auto my-12" v-if="dataItems.length > 0" max-width="800px">
            <svg id="heuristic_details_canvas" :class="!enoughDataForGraph?'hide':''"/>
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
          <v-card outlined class="mx-auto my-12" v-if="dataItems.length > 0">
            <v-data-table :headers="dataHeaders"
                          :items="dataItems"
                          :items-per-page="5"
            ></v-data-table>
          </v-card>
        </div>
      </v-card-text>
    </v-card>
  </v-bottom-sheet>
</template>

<script>
import {
  mdiIframeVariableOutline, mdiTune, mdiPoundBoxOutline,
} from '@mdi/js';
import * as d3 from 'd3';
import IconItem from '../common/IconItem.vue';

// addPercentageToDate returns a new date which has a percentage of duration added
function addPercentageToDate(date, duration, percentage) {
  const newDate = new Date(date);
  newDate.setTime(newDate.getTime() + duration * percentage);
  return newDate;
}

export default {
  name: 'Details',
  components: { IconItem },
  props: {
    // v-model
    value: { type: Boolean, required: true },
    heuristicData: { type: Object, required: true },
    // map[addresshash]array[origins]
    addressMap: { type: Map, required: false },
    newHeuristicPrefix: { type: String, required: true },
  },
  data() {
    return {
      icon: {
        mdiIframeVariableOutline, mdiTune, mdiPoundBoxOutline,
      },
      chart: null,
      dataHeaders: [
        {
          text: 'Address', align: 'start', sortable: false, value: 'address',
        },
        { text: 'Number of Origins', value: 'len' },
      ],
      enoughDataForGraph: true,
      durationInMinutes: 0,
    };
  },
  computed: {
    dataItems() {
      const dataItems = [];
      if (this.addressMap !== undefined) {
        this.addressMap.forEach((v, k) => {
          dataItems.push({ address: k, len: v.length });
        });
      }

      return dataItems;
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
    details: {
      get() {
        return this.$store.getters.getHeuristicDetails;
      },
      set(value) {
        if (value === null) {
          this.$store.dispatch('resetHeuristicDetails');
          return;
        }
        this.$store.dispatch('setHeuristicDetails', value);
      },
    },
  },
  methods: {
    updateData(graphData) {
      const svgCanvasId = 'heuristic_details_canvas';
      const detailArray = [];
      let lowestDate = null;
      let highestDate = null;

      graphData.forEach((v) => {
        v.dateTime = new Date(v.ts);
        if (lowestDate === null || lowestDate > v.dateTime) lowestDate = v.dateTime;
        if (highestDate === null || highestDate < v.dateTime) highestDate = v.dateTime;

        detailArray.push(v);
      });

      const duration = highestDate - lowestDate;

      // check if there is enough data to draw the diagram
      if (duration < 1000 * 60 * 60 * 3) {
        this.enoughDataForGraph = false;
        this.durationInMinutes = Math.floor(duration / 1000 / 60);
        return;
      }
      this.enoughDataForGraph = true;

      // add a percentage of time to the date limitations,
      // so all rectangles can be displayed in their full width
      const lowestRange = addPercentageToDate(lowestDate, duration, -0.03);
      const highestRange = addPercentageToDate(highestDate, duration, 0.03);

      // 1000*60*60*24*2.5 = 2.5 days
      // 216000000 = 2.5 days
      const smallestDuration = 216000000;
      let numTicks = Math.floor(duration / smallestDuration);
      if (numTicks === 0) numTicks = 1;

      const svg = d3.select(`#${svgCanvasId}`);
      const margin = {
        top: 20, right: 30, bottom: 45, left: 40,
      };
      const width = 600 - margin.left - margin.right;
      const height = 300 - margin.top - margin.bottom;

      // set the ranges
      const x = d3.scaleTime()
        .domain([lowestRange, highestRange])
        .rangeRound([0, width]);

      const y = d3.scaleLinear()
        .range([height, 0]);

      // set the parameters for the histogram
      const histogram = d3.bin()
        .value((d) => d.dateTime)
        .domain(x.domain())
        .thresholds(x.ticks(d3.timeHour.every(numTicks)));

      // append the svg object to the body of the page
      // append a 'group' element to 'svg'
      // moves the 'group' element to the top left margin
      const svgGroup = svg
        .attr('width', width + margin.left + margin.right)
        .attr('height', height + margin.top + margin.bottom)
        .append('g')
        .attr('transform', `translate(${margin.left},${margin.top})`);

      // group the data for the bars
      const bins = histogram(detailArray);

      // Scale the range of the data in the y domain
      y.domain([0, d3.max(bins, (d) => d.length)]);

      // append the bar rectangles to the svg element
      svgGroup.selectAll('rect')
        .data(bins)
        .enter().append('rect')
        .attr('class', 'bar')
        .attr('x', 1)
        .attr('transform', (d) => `translate(${x(d.x0)},${y(d.length)})`)
        .attr('width', (d) => x(d.x1) - x(d.x0) - 1)
        .attr('height', (d) => height - y(d.length));

      // add the x Axis
      svgGroup.append('g')
        .attr('transform', `translate(0,${height})`)
        .call(d3.axisBottom(x));
      // .call(d3.axisBottom(x).tickFormat(d3.timeFormat('%b %d %I:%M')));

      // add x title
      svgGroup.append('text')
        .attr('fill', 'currentColor')
        .attr('font-family', 'sans-serif')
        .attr('font-size', 10)
        .attr('transform',
          `translate(${(width / 2)} ,${
            height + margin.top + 10})`)
        .style('text-anchor', 'middle')
        .text('Time');

      // add x title description
      svgGroup.append('text')
        .attr('fill', 'currentColor')
        .attr('font-family', 'sans-serif')
        .attr('font-size', 10)
        .attr('transform',
          `translate(${(width / 2)} ,${
            height + margin.top + 22})`)
        .style('text-anchor', 'middle')
        .text(`${lowestDate.toLocaleString()} - ${highestDate.toLocaleString()}`);

      // add the y Axis
      svgGroup.append('g')
        .call(d3.axisLeft(y));

      // add y title
      svgGroup.append('text')
        .attr('fill', 'currentColor')
        .attr('font-family', 'sans-serif')
        .attr('font-size', 10)
        .attr('transform', 'rotate(-90)')
        .attr('y', 0 - margin.left)
        .attr('x', 0 - (height / 2) - 20)
        .attr('dy', '1em')
        .style('text-anchor', 'middle')
        .text('Occurrences');

      let title = 'Number of origins per ';

      if (numTicks > 1) {
        title += `${numTicks} hours`;
      } else {
        title += 'hour';
      }

      // title
      svgGroup.append('text')
        .attr('fill', 'currentColor')
        .attr('font-family', 'sans-serif')
        .attr('font-size', 12)
        .attr('y', -5)
        .attr('x', width / 2 - 50)
        .text(title);
    },
  },
  updated() {
    // do nothing if sheet is not open
    if (!this.value || this.details.size === 0) return;
    const svgCanvasId = 'heuristic_details_canvas';

    // check if svg exists yet
    const documentSvg = document.getElementById(svgCanvasId);
    if (documentSvg === null) return;
    // reset svg
    documentSvg.innerHTML = '';
    if (!this.details.has(this.heuristicData.heuristicUid)) return;

    // this.updateData(this.addressMap.entries().next().value[1]);
    this.updateData(this.details.get(this.heuristicData.heuristicUid).results);
  },
};
</script>

<style>

.bar {
  fill: #008ee5;
}

.hide {
  display: none;
}

</style>
