import {isFunction} from '@/utilities';
import {drag as d3Drag} from 'd3-drag';
import {select as d3Select} from 'd3-selection';
import {zoom} from 'd3-zoom';
import {forceSimulation, forceLink, forceManyBody, forceX, forceY, forceRadial, forceCenter} from 'd3-force';

function drag(simulation) {
	function dragStarted(event) {
		if (!event.active) {
			simulation.alphaTarget(0.3).restart();
		}

		event.subject.fx = event.subject.x;
		event.subject.fy = event.subject.y;
	}

	function dragged(event) {
		event.subject.fx = event.x;
		event.subject.fy = event.y;
	}

	function dragEnded(event) {
		if (!event.active) {
			simulation.alphaTarget(0);
		}

		event.subject.fx = null;
		event.subject.fy = null;
	}

	return d3Drag()
		.on('start', dragStarted)
		.on('drag', dragged)
		.on('end', dragEnded);
}

export default class ForceGraph {
	constructor(width, height, svgId, colorMap) {
		this.svgId = svgId;
		this.width = width;
		this.height = height;
		this.colorMap = colorMap;

		this.initSvg();

		this.clickCallBack = null;
		this.simulation = null;
	}

	initSvg() {
		// Add attributes to root svg
		this.rootSvg = d3Select(`#${this.svgId}`).attr('viewBox', `0 0 ${this.width} ${this.height}`);
		this.rootGroup = this.rootSvg.append('g').attr('class', 'root-group');

		this.lineGroup = this.rootGroup.append('g')
			.attr('stroke', '#999')
			.attr('stroke-opacity', 0.6);

		this.nodeGroup = this.rootGroup
			.append('g')
			.attr('stroke', '#fff')
			.attr('stroke-width', '#C2C2C2');

		// Add zoom and drag
		this.zoom = zoom()
			.on('zoom', event => {
				this.rootGroup.attr('transform', event.transform);
			})
			.scaleExtent([0.25, 8]);
		this.rootSvg.call(this.zoom);

		// Add arrow definition
		this.rootSvg
			.append('svg:defs')
			.append('svg:marker')
			.attr('id', 'arrowhead')
			.attr('viewBox', '0 -5 10 10')
			.attr('refX', 22)
			.attr('refY', 0)
			.attr('markerWidth', 6)
			.attr('markerHeight', 6)
			.attr('orient', 'auto')
			.append('svg:path')
			.attr('d', 'M0,-5L10,0L0,5')
			.attr('fill', '#999');
	}

	draw(nodes, links) {
		if (this.clickCallBack === null) {
			throw new Error('click call back is not set');
		}

		// If there is a simulation ongoing from a previous call, stop it
		if (this.simulation) {
			this.simulation.stop();
		}

		// Check if the svg is still initialized
		if (d3Select(`#${this.svgId}`).nodes()[0].childElementCount === 0) {
			this.initSvg();
		}

		this.simulation = forceSimulation(nodes)
			.force('link', forceLink(links).id(d => d.txhash))
			.force('charge', forceManyBody().strength(-80))
			.force('x', forceX(this.width / 2).strength(0.04))
			.force('y', forceY(this.height / 2).strength(0.04))
			.force('radial', forceRadial(240, this.width / 2, this.height / 2))
			.force('center', forceCenter(this.width / 2, this.height / 2));

		const link = this.lineGroup
			.selectAll('.arrow')
			.data(links, d => `${d.source}${d.target}`)
			.join('line')
			.attr('class', 'arrow')
			.attr('marker-end', 'url(#arrowhead)')
			.attr('stroke-width', 1);

		const node = this.nodeGroup
			.selectAll('circle')
			.data(nodes, d => d.txhash)
			.join('circle')
			.attr('r', 7)
			.attr('fill', d => this.colorMap.get(d.privacytype))
			.on('click', (e, d) => {
				this.clickCallBack(d);
			})
		// eslint-disable-next-line func-names
			.on('mouseover', function mouseOver() {
				d3Select(this).attr('r', 10).classed('nodeMouseOver', true);
			})
		// eslint-disable-next-line func-names
			.on('mouseout', function mouseOut() {
				d3Select(this).attr('r', 7).classed('nodeMouseOver', false);
			}).call(drag(this.simulation));

		node.append('title')
			.text(d => `${d.txhash}\n${d.privacytype}`);

		this.simulation.on('tick', () => {
			link
				.attr('x1', d => d.source.x)
				.attr('y1', d => d.source.y)
				.attr('x2', d => d.target.x)
				.attr('y2', d => d.target.y);

			node
				.attr('cx', d => d.x)
				.attr('cy', d => d.y);
		});
	}

	setClickHandler(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.clickCallBack = callback;
		return true;
	}
}
