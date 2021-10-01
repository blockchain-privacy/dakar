import * as d3 from 'd3';

// addPercentageToDate returns a new date which has a percentage of duration added
function addPercentageToDate(date, duration, percentage) {
  const newDate = new Date(date);
  newDate.setTime(newDate.getTime() + duration * percentage);
  return newDate;
}

export default class Histogram {
  constructor(svgId) {
    this.svgId = svgId;
    this.isEmpty = false;
    this.durationInMinutes = 0;
  }

  get empty() {
    return this.isEmpty;
  }

  get getDurationInMinutes() {
    return this.durationInMinutes;
  }

  // reset removes all content from the root svg
  reset() {
    this.isEmpty = false;
    this.durationInMinutes = 0;

    // check if svg exists yet
    const documentSvg = document.getElementById(this.svgId);
    if (documentSvg === null) return;
    // reset svg
    documentSvg.innerHTML = '';
  }

  draw(graphData) {
    let lowestDate = null;
    let highestDate = null;

    const detailArray = graphData.map((d) => {
      d.dateTime = new Date(d.ts);
      if (lowestDate === null || lowestDate > d.dateTime) lowestDate = d.dateTime;
      if (highestDate === null || highestDate < d.dateTime) {
        highestDate = d.dateTime;
      }
      return d;
    });

    const duration = highestDate - lowestDate;

    // check if there is enough data to draw the diagram; 1000 * 60 * 60 * 3 = 10800000
    if (duration < 10800000) {
      this.isEmpty = true;
      this.durationInMinutes = Math.floor(duration / 1000 / 60);
      return;
    }
    this.isEmpty = false;

    // add a percentage of time to the date limitations,
    // so all rectangles can be displayed in their full width
    const lowestRange = addPercentageToDate(lowestDate, duration, -0.03);
    const highestRange = addPercentageToDate(highestDate, duration, 0.03);

    // 1000*60*60*24*2.5 = 2.5 days
    // 216000000 = 2.5 days
    let numTicks = Math.floor(duration / 216000000);
    if (numTicks === 0) numTicks = 1;

    const svg = d3.select(`#${this.svgId}`);
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
  }
}
