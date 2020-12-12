<template>
  <v-bottom-sheet scrollable v-model="inputVal">
    <v-card style="max-height: 400px">
      <v-subheader>Heuristic Properties</v-subheader>
      <v-card-text style="height: 80%">
        <div class="d-flex flex-wrap" style="align-items: flex-start;">
          <v-card
              class="mx-auto my-12"
              max-width="300">
            <v-card-title>
              Heuristic Properties
            </v-card-title>
            <v-card-subtitle>
              <p>Type: {{ heuristicData.heuristicType }}</p>
              <p v-if="heuristicData.heuristicParameter">
                Parameter: {{ heuristicData.heuristicParameter }}
              </p>
              <p v-if="heuristicData.resultCount">
                Number of results: {{ heuristicData.resultCount }}
              </p>
            </v-card-subtitle>
          </v-card>
          <svg id="heuristic_details_canvas" ></svg>
          <v-card
              class="mx-auto my-12"
              v-for="[key, v] of addressMap"
              :key="key"
              max-width="300">
            <v-card-title>
              {{ key }}
            </v-card-title>
            <v-card-subtitle>
              Number of origins: {{ v.length }}
            </v-card-subtitle>
          </v-card>
        </div>
      </v-card-text>
    </v-card>
  </v-bottom-sheet>
</template>

<script>

import * as d3 from 'd3';

export default {
  name: 'HeuristicDetails',
  props: {
    // v-model
    value: Boolean,
    heuristicData: Object,
    // map[addresshash]array[origins]
    addressMap: Map,
  },
  data() {
    return {
      chart: null,
      chartData: [{ letter: 'A', frequency: 3 }, { letter: 'B', frequency: 10 },
        { letter: 'C', frequency: 15 }, { letter: 'D', frequency: 8 }, { letter: 'E', frequency: 1 }],
    };
  },
  computed: {
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
  updated() {
    // do nothing if sheet is not open
    if (!this.value) return;
    const svgCanvasId = 'heuristic_details_canvas';
    const detailArray = [];
    this.details.get(this.heuristicData.heuristicUid).results.forEach((v) => {
      v.dateTime = new Date(v.ts);
      detailArray.push(v);
    });

    console.log(detailArray);

    const margin = {
      top: 10, right: 30, bottom: 30, left: 50,
    };
    const width = 960 - margin.left - margin.right;
    const height = 500 - margin.top - margin.bottom;

    const parseDate = d3.isoParse;
    const formatDate = d3.isoFormat;

    const x = d3.scaleTime().domain([]).range([0, width]);
    const y = d3.scaleLinear().range([height, 0]);

    const xAxis = (g) => g
      .attr('transform', `translate(0,${height - margin.bottom})`)
      .call(d3.axisBottom(x).tickFormat(formatDate));

    const yAxis = (g) => g
      .attr('transform', `translate(${margin.left},0)`)
      .call(d3.axisLeft(y).ticks(6))
      .call((d) => d.select('.domain').remove());

    // Create the SVG drawing area
    const svg = d3.select(`#${svgCanvasId}`)
      .attr('width', width + margin.left + margin.right)
      .attr('height', height + margin.top + margin.bottom)
      .append('g')
      .attr('transform', `translate(${margin.left},${margin.top})`);

    // Parse the date strings into javascript dates
    detailArray.forEach((d) => {
      d.created_date = parseDate(d.ts);
    });

    // Determine the first and list dates in the data set
    const monthExtent = d3.extent(detailArray, (d) => d.created_date);

    // Create one bin per month, use an offset to include the first and last months
    const monthBins = d3.timeMonth(d3.timeMonth.offset(monthExtent[0], -1),
      d3.timeMonth.offset(monthExtent[1], 1));

    // Use the histogram layout to create a function that will bin the data
    const binByMonth = d3.layout.histogram()
      .value((d) => d.created_date)
      .bins(monthBins);

    // Bin the data by month
    const histData = binByMonth(detailArray);

    // Scale the range of the data by setting the domain
    x.domain(d3.extent(monthBins));
    y.domain([0, d3.max(histData, (d) => d.y)]);

    // Set up one bar for each bin
    // Months have slightly different lengths so calculate the width for each bin
    // Note: dx, the width of the histogram bin, is in milliseconds so convert the x value
    // into UTC time and convert the sum back to a date in order to help calculate the width
    // Thanks to npdoty for pointing this out in this stack overflow post:
    // http://stackoverflow.com/questions/17745682/d3-histogram-date-based
    svg.selectAll('.bar')
      .data(histData)
      .enter().append('rect')
      .attr('class', 'bar')
      .attr('x', (d) => x(d.x))
      .attr('width', (d) => x(new Date(d.x.getTime() + d.dx)) - x(d.x) - 1)
      .attr('y', (d) => y(d.y))
      .attr('height', (d) => height - y(d.y));

    // Add the X Axis
    svg.append('g')
      .attr('class', 'x axis')
      .attr('transform', `translate(0,${height})`)
      .call(xAxis);

    // Add the Y Axis and label
    svg.append('g')
      .attr('class', 'y axis')
      .call(yAxis)
      .append('text')
      .attr('transform', 'rotate(-90)')
      .attr('y', 6)
      .attr('dy', '.71em')
      .style('text-anchor', 'end')
      .text('Number of Sightings');
  },
};
</script>

<style scoped>

</style>
