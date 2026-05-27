// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {select as d3Select} from 'd3-selection';
import {scaleTime, scaleLinear} from 'd3-scale';
import {timeTickInterval} from 'd3-time';
import {bin, max, group} from 'd3-array';
import {axisBottom, axisLeft} from 'd3-axis';
import {format} from 'd3-format';
import {isFunction} from '@/utilities';

// AddPercentageToDate returns a new date which has a percentage of duration added
function addPercentageToDate(date, duration, percentage) {
	const newDate = new Date(date);
	newDate.setTime(newDate.getTime() + (duration * percentage));
	return newDate;
}

export default class BarChart {
	constructor(svgId, width, height) {
		this.svgId = svgId;
		this.width = width;
		this.height = height;
		this.isEmpty = false;
		this.durationInMinutes = 0;
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

	// Draws the graph.
	// Set the click handler before calling this function.
	draw(graphData) {
		this.drawStacked(graphData, null);
	}

	// Draws the graph. If colorMap is set, draws a stacked graph.
	// Set the click handler before calling this function.
	drawStacked(graphData, colorMap) {
		let lowestDate = null;
		let highestDate = null;

		const detailArray = [];
		for (const d of graphData) {
			if (d.dateTime === undefined && d.ts === undefined) {
				continue;
			}

			if (d.dateTime === undefined) {
				d.dateTime = new Date(d.ts);
			} else if (!(d.dateTime instanceof Date)) {
				d.dateTime = new Date(d.ts);
			}

			if (lowestDate === null || lowestDate > d.dateTime) {
				lowestDate = d.dateTime;
			}

			if (highestDate === null || highestDate < d.dateTime) {
				highestDate = d.dateTime;
			}

			detailArray.push(d);
		}

		let duration = 0;
		if (highestDate !== null && lowestDate !== null) {
			duration = highestDate - lowestDate;
		}

		// Check if there is enough data to draw the diagram; 1000 * 60 * 60 * 3 = 10800000
		if (duration < 180_000) {
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
		svg.selectAll('*').remove();
		const margin = {
			top: 10, right: 10, bottom: 50, left: 45,
		};
		const width = this.width - margin.left - margin.right;
		const height = this.height - margin.top - margin.bottom;

		// Set the ranges
		const x = scaleTime().domain([lowestRange, highestRange]).rangeRound([0, width]);
		const y = scaleLinear().range([height, 0]);

		// Set the parameters for the bar chart
		const barChart = bin()
			.value(d => d.dateTime)
			.domain(x.domain())
			.thresholds(x.ticks(timeTickInterval(lowestRange, highestRange, 40)));

		const svgGroup = svg
			.attr('viewBox', `0 0 ${this.width} ${this.height}`)
			.append('g')
			.attr('transform', `translate(${margin.left},${margin.top})`);

		// Group the data for the bars
		const bins = barChart(detailArray);

		// Scale the range of the data in the y domain
		y.domain([0, max(bins, d => d.length)]);

		let bars;
		if (colorMap === null) {
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

					// All data of d grouped by transaction type
					const txTypeGroups = group(d, e => e.txtype);

					if (txTypeGroups.size === 0) {
						return elements;
					}

					colorMap.forEach((v, txtype) => {
						const g = txTypeGroups.get(txtype);
						if (g === undefined) {
							// This transaction type does not exist
							return;
						}

						elements.push({
							parentSize,
							width: x(d.x1) - x(d.x0) - 1,
							height: g.length,
							color: colorMap.get(txtype),
							txType: txtype,
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
				.attr('height', d => height - y(d.height))
				.on('mousemove', function (e, d) {
					d3Select(`.${this.svgId}_tooltip`)
						.data([{txType: d.txType}])
						.join('span')
						.classed(`${this.svgId}_tooltip`, true)
						.style('left', () => (window.innerWidth - e.clientX) >= 200 ? `${e.clientX + 15}px` : undefined)
						.style('right', () => (window.innerWidth - e.clientX) < 200 ? `${window.innerWidth - e.clientX + 5}px` : undefined)
						.style('top', `${e.pageY + 10}px`)
						.style('background-color', 'grey')
						.style('color', 'white')
						.style('border-radius', '6px')
						.style('padding', '0 5px 0 5px')
						.style('border', '1px black')
						.text(element => element.txType)
						.style('visibility', 'visible')
						.style('z-index', 1500)
						.style('position', 'absolute');
				})
				.on('mouseleave', function () {
					d3Select(`.${this.svgId}_tooltip`).style('visibility', 'hidden');
				});

			// Set overlay which animates the bars and has event handler attached
			if (this.clickCallBack !== null) {
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
						d3Select(this).attr('opacity', 0.4);
					});
			}
		}

		bars.attr('transform', d => {
			if (x(d.x0) === undefined) {
				return null;
			}

			return `translate(${x(d.x0)},${y(d.length)})`;
		});

		// Add the x Axis
		svgGroup.append('g')
			.attr('transform', `translate(0,${height})`)
			.call(axisBottom(x).ticks(6));

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
		const yAxisTicks = y.ticks(5).filter(tick => Number.isInteger(tick));

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
			.attr('x', 0 - (height / 2))
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
