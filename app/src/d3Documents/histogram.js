import * as d3 from 'd3';

// addPercentageToDate returns a new date which has a percentage of duration added
function addPercentageToDate(date, duration, percentage) {
  const newDate = new Date(date);
  newDate.setTime(newDate.getTime() + duration * percentage);
  return newDate;
}

export default class Histogram {
  constructor(svgId, width, height, title) {
    this.svgId = svgId;
    this.width = width;
    this.height = height;
    this.title = title;
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
    this.drawStacked(graphData, [], null);
  }

  drawStacked(graphData, categories, colorMap, chartTitle) {
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
    if (duration < 180000) {
      this.isEmpty = true;
      this.durationInMinutes = Math.floor(duration / 1000 / 60);
      return;
    }
    this.isEmpty = false;

    // add a percentage of time to the date limitations,
    // so all rectangles can be displayed in their full width
    const lowestRange = addPercentageToDate(lowestDate, duration, -0.03);
    const highestRange = addPercentageToDate(highestDate, duration, 0.03);

    // 1000 / 60 -> minutes
    // /40 -> get about 40 bars
    let numTicks = Math.floor(duration / 1000 / 60 / 40);

    if (numTicks === 0) numTicks = 1;

    let timeScale = d3.timeMinute.every(numTicks);
    let timeUnit = 'minute';
    let timeFactor = 1 / 1000 / 60;

    if (numTicks > 30 && numTicks < 60 * 24 * 5) {
      // hour
      numTicks = Math.floor(numTicks / 60);
      timeScale = d3.timeHour.every(numTicks);
      timeUnit = 'hour';
      timeFactor /= 60;
    } else if (numTicks >= 60 * 24 * 5) {
      // day
      numTicks = Math.floor(numTicks / 60 / 24);
      timeScale = d3.timeDay.every(numTicks);
      timeUnit = 'day';
      timeFactor /= (60 * 24);
    }

    const svg = d3.select(`#${this.svgId}`);

    const margin = {
      top: 25, right: 30, bottom: 50, left: 50,
    };
    const width = this.width - margin.left - margin.right;
    const height = this.height - margin.top - margin.bottom;

    // set the ranges
    const x = d3.scaleTime().domain([lowestRange, highestRange]).rangeRound([0, width]);
    const y = d3.scaleLinear().range([height, 0]);

    const xTicks = x.ticks(timeScale);
    const binSize = Math.max(Math.floor((xTicks[1] - xTicks[0]) * timeFactor),
      Math.floor((xTicks[2] - xTicks[1]) * timeFactor),
      Math.floor((xTicks[3] - xTicks[2]) * timeFactor));

    // set the parameters for the histogram
    const histogram = d3.bin()
      .value((d) => d.dateTime)
      .domain(x.domain())
      .thresholds(xTicks);

    const svgGroup = svg
      .attr('viewBox', `0 0 ${this.width} ${this.height}`)
      .append('g')
      .attr('transform', `translate(${margin.left},${margin.top})`);

    // group the data for the bars
    const bins = histogram(detailArray);

    // Scale the range of the data in the y domain
    y.domain([0, d3.max(bins, (d) => d.length)]);

    const t = d3.transition().duration(300).ease(d3.easeLinear);

    if (categories.length === 0) {
      // append bar rectangles to the svg element
      const bars = svgGroup.selectAll('rect')
        .data(bins)
        .join('rect')
        .attr('class', 'bar')
        .attr('x', 1)
        .attr('width', (d) => x(d.x1) - x(d.x0) - 1)
        .attr('height', (d) => height - y(d.length))
        .attr('transform', (d) => `translate(${0},${y(d.length)})`);

      bars
        .transition(t)
        .attr('transform', (d) => `translate(${x(d.x0)},${y(d.length)})`);
    } else {
      // append stacked bar rectangles to the svg element
      const bars = svgGroup.selectAll('.bar')
        .data(bins)
        .join('g')
        .attr('class', 'bar')
        .attr('transform', (d) => `translate(${0},${y(d.length)})`);
      bars
        .transition(t)
        .attr('transform', (d) => `translate(${x(d.x0)},${y(d.length)})`);

      bars.selectAll('.subBar')
        .data((d) => {
          const elements = [];
          let parentSize = 0;

          const privacyGroups = d3.group(d, (e) => e.privacytype);

          if (privacyGroups.size === 0) {
            return elements;
          }

          colorMap.forEach((v, k) => {
            const g = privacyGroups.get(k);
            if (g === undefined) {
              return;
            }

            elements.push({
              parentSize,
              width: x(d.x1) - x(d.x0) - 1,
              height: g.length,
              color: colorMap.get(g[0].privacytype),
            });
            parentSize += g.length;
          });

          return elements;
        })
        .join('rect')
        .attr('class', 'subBar')
        .attr('x', 1)
        .attr('fill', (d) => d.color)
        .attr('width', (d) => d.width)
        .attr('y', (d) => height - y(d.parentSize))
        .attr('height', (d) => height - y(d.height));
    }

    // add the x Axis
    svgGroup.append('g')
      .attr('transform', `translate(0,${height})`)
      .call(d3.axisBottom(x));

    // add x title description
    svgGroup.append('text')
      .attr('fill', 'currentColor')
      .attr('font-family', 'sans-serif')
      .attr('font-size', '1em')
      .attr('transform',
        `translate(${(width / 2)} ,${
          height + margin.top + 20})`)
      .style('text-anchor', 'middle')
      .text(`${lowestDate.toLocaleString()} - ${highestDate.toLocaleString()}`);

    // only allow integer on scale
    const yAxisTicks = y.ticks().filter((tick) => Number.isInteger(tick));

    // add the y Axis
    svgGroup.append('g')
      .call(d3.axisLeft(y).tickValues(yAxisTicks)
        .tickFormat(d3.format('d')));

    // add y title
    svgGroup.append('text')
      .attr('fill', 'currentColor')
      .attr('font-family', 'sans-serif')
      .attr('font-size', '1em')
      .attr('transform', 'rotate(-90)')
      .attr('y', 0 - margin.left)
      .attr('x', 0 - (height / 2) - 20)
      .attr('dy', '1em')
      .style('text-anchor', 'middle')
      .text('Occurrences');

    let title = `${this.title} per `;

    if (chartTitle !== undefined) {
      title = `${chartTitle} per `;
    }

    if (binSize > 1) {
      title += `${binSize} ${timeUnit}s`;
    } else {
      title += timeUnit;
    }

    // title
    svgGroup.append('text')
      .attr('fill', 'currentColor')
      .attr('font-family', 'sans-serif')
      .attr('font-size', '1.2em')
      .attr('transform',
        `translate(${(width / 2)} ,-8)`)
      .style('text-anchor', 'middle')
      .text(title);
  }
}
