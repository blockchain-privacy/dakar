import {isFunction} from '@/utilities';
import {select as d3Select} from 'd3-selection';
import {scaleTime, scaleLinear} from 'd3-scale';
import {timeTickInterval} from 'd3-time';
import {bin, max, group} from 'd3-array';
import {transition} from 'd3-transition';
import {easeLinear} from 'd3-ease';
import {axisBottom, axisLeft} from 'd3-axis';
import {format} from 'd3-format';

// AddPercentageToDate returns a new date which has a percentage of duration added
function addPercentageToDate(date, duration, percentage) {
	const newDate = new Date(date);
	newDate.setTime(newDate.getTime() + duration * percentage);
	return newDate;
}

export default class Histogram {
	constructor(svgId, width, height, enableTransition = true) {
		this.svgId = svgId;
		this.width = width;
		this.height = height;
		this.isEmpty = false;
		this.durationInMinutes = 0;
		this.enableTransition = enableTransition;
		this.clickCallBack = null;
	}

	get empty() {
		return this.isEmpty;
	}

	get getDurationInMinutes() {
		return this.durationInMinutes;
	}

	// Reset removes all content from the root svg
	reset() {
		this.isEmpty = false;
		this.durationInMinutes = 0;

		// Check if svg exists yet
		const documentSvg = document.getElementById(this.svgId);
		if (documentSvg === null) {
			return;
		}

		// Reset svg
		documentSvg.innerHTML = '';
	}

	draw(graphData) {
		this.drawStacked(graphData, [], null);
	}

	drawStacked(graphData, categories, colorMap) {
		let lowestDate = null;
		let highestDate = null;

		const detailArray = graphData.map(d => {
			if (d.dateTime === undefined) {
				d.dateTime = new Date(d.ts);
			}

			if (lowestDate === null || lowestDate > d.dateTime) {
				lowestDate = d.dateTime;
			}

			if (highestDate === null || highestDate < d.dateTime) {
				highestDate = d.dateTime;
			}

			return d;
		});

		const duration = highestDate - lowestDate;

		// Check if there is enough data to draw the diagram; 1000 * 60 * 60 * 3 = 10800000
		if (duration < 180000) {
			this.isEmpty = true;
			this.durationInMinutes = Math.floor(duration / 1000 / 60);
			return;
		}

		this.isEmpty = false;

		// Add a percentage of time to the date limitations,
		// so all rectangles can be displayed in their full width
		const lowestRange = addPercentageToDate(lowestDate, duration, -0.03);
		const highestRange = addPercentageToDate(highestDate, duration, 0.03);

		const svg = d3Select(`#${this.svgId}`);

		const margin = {
			top: 10, right: 10, bottom: 50, left: 45,
		};
		const width = this.width - margin.left - margin.right;
		const height = this.height - margin.top - margin.bottom;

		// Set the ranges
		const x = scaleTime().domain([lowestRange, highestRange]).rangeRound([0, width]);
		const y = scaleLinear().range([height, 0]);

		// Set the parameters for the histogram
		const histogram = bin()
			.value(d => d.dateTime)
			.domain(x.domain())
			.thresholds(x.ticks(timeTickInterval(lowestRange, highestRange, 40)));

		const svgGroup = svg
			.attr('viewBox', `0 0 ${this.width} ${this.height}`)
			.append('g')
			.attr('transform', `translate(${margin.left},${margin.top})`);

		// Group the data for the bars
		const bins = histogram(detailArray);

		// Scale the range of the data in the y domain
		y.domain([0, max(bins, d => d.length)]);

		let bars;
		if (categories.length === 0) {
			// Append bar rectangles to the svg element
			bars = svgGroup.selectAll('rect')
				.data(bins)
				.join('rect')
				.attr('class', 'bar')
				.attr('x', 1)
				.attr('width', d => x(d.x1) - x(d.x0) - 1)
				.attr('height', d => height - y(d.length));
		} else {
			// Append stacked bar rectangles to the svg element
			bars = svgGroup.selectAll('.stackedBar')
				.data(bins)
				.join('g')
				.attr('class', 'stackedBar');

			bars.selectAll('.subBar')
				.data(d => {
					// D: data of one svg group which will later contain the stacked rects of one time slot

					const elements = [];
					let parentSize = 0;

					// All data of d grouped by privacy type
					const privacyGroups = group(d, e => e.privacytype);

					if (privacyGroups.size === 0) {
						return elements;
					}

					colorMap.forEach((v, privacyType) => {
						const g = privacyGroups.get(privacyType);
						if (g === undefined) {
							// This privacy type does not exist
							return;
						}

						elements.push({
							parentSize,
							width: x(d.x1) - x(d.x0) - 1,
							height: g.length,
							color: colorMap.get(privacyType),
							// Transactions,
						});
						parentSize += g.length;
					});

					return elements;
				})
				.join('rect')
				.attr('class', 'subBar')
				.attr('x', 1)
				.attr('fill', d => d.color)
				.attr('width', d => d.width)
				.attr('y', d => height - y(d.parentSize))
				.attr('height', d => height - y(d.height));

			// Const self = this;
			// set overlay which animates the bars and has event handler attached
			bars.append('rect')
				.attr('class', 'overlay')
				.attr('x', 1)
				.attr('opacity', 0)
				.attr('width', d => x(d.x1) - x(d.x0) - 1)
				.attr('height', d => height - y(d.length))
				.on('click', (e, d) => {
					this.clickCallBack(d);
				})
			// eslint-disable-next-line func-names
				.on('mouseout', function mouseOut() {
					d3Select(this).attr('opacity', 0);
				})
			// eslint-disable-next-line func-names
				.on('mouseover', function mouseOver() {
					d3Select(this).attr('opacity', 1);
				});
		}

		if (this.enableTransition) {
			bars
				.attr('transform', d => `translate(${0},${y(d.length)})`)
				.transition(transition().duration(300).ease(easeLinear))
				.attr('transform', d => `translate(${x(d.x0)},${y(d.length)})`);
		} else {
			bars.attr('transform', d => `translate(${x(d.x0)},${y(d.length)})`);
		}

		// Add the x Axis
		svgGroup.append('g')
			.attr('transform', `translate(0,${height})`)
			.call(axisBottom(x));

		// Add x title description
		svgGroup.append('text')
			.attr('fill', 'currentColor')
			.attr('font-family', 'sans-serif')
			.attr('font-size', '1em')
			.attr(
				'transform',
				`translate(${(width / 2)} ,${
					height + margin.top + 20})`,
			)
			.style('text-anchor', 'middle')
			.text(`${lowestDate.toLocaleString()} - ${highestDate.toLocaleString()}`);

		// Only allow integer on scale
		const yAxisTicks = y.ticks().filter(tick => Number.isInteger(tick));

		// Add the y Axis
		svgGroup.append('g')
			.call(axisLeft(y).tickValues(yAxisTicks)
				.tickFormat(format('d')));

		// Add y title
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
	}

	setClickHandler(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.clickCallBack = callback;
		return true;
	}
}
