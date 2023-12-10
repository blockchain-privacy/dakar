import {isFunction} from '@/utilities';
import {drag} from 'd3-drag';
import {select as d3Select} from 'd3-selection';
import {zoom} from 'd3-zoom';
import {forceSimulation, forceLink, forceManyBody, forceCollide} from 'd3-force';
import {reduceX, reduceY} from '@/d3Documents/util';

// Sets a node with a valid x attribute to be excluded from force simulations
function setFxFy(node) {
	if (node.x !== undefined) {
		node.fx = node.x;
		node.fy = node.y;
	}

	return node;
}

function dragStarted(event, simulation) {
	if (!event.active) {
		simulation.alphaTarget(0.3).restart();
	}

	event.subject.fx = event.subject.x;
	event.subject.fy = event.subject.y;
}

function dragged(event, d3This) {
	event.subject.fx = event.x;
	event.subject.fy = event.y;
	// Raise() causes bug in chrome: click is only
	// recognized on second time. moved here from dragStart
	d3Select(d3This).raise();
}

function dragEnded(event, simulation) {
	if (!event.active) {
		simulation.alphaTarget(0);
	}
}

export default class HeuristicGraph {
	constructor() {
		this.nodeClickCallBack = null;
		this.svgZoomCallback = null;
		this.svgClickCallback = null;
		this.contextMenuCallback = null;

		// Context menu
		this.activeContextMenuNode = null;
		this.activeContextMenuSelection = null;

		// Svg
		this.simulation = null;
		this.nodeRadius = 14;
		this.rootSvg = null;
		this.rootGroup = null;
		this.lineGroup = null;
		this.nodeGroup = null;
		this.zoom = null;

		// Data
		this.nodeMap = new Map();
	}

	svgClick() {
		this.resetClick();
		if (this.svgClickCallback !== null) {
			this.svgClickCallback();
		}
	}

	resetClick() {
		this.nodeGroup.selectAll('.clicked')
			.classed('clicked', false)
			.attr('stroke', 'white')
			.attr('stroke-width', 1);
	}

	nodeClick(e, d, d3This) {
		e.stopPropagation();
		this.resetClick();
		d3Select(d3This)
			.classed('clicked', true)
			.attr('stroke', '#B71C1C')
			.attr('stroke-width', 3);

		if (this.nodeClickCallBack !== null) {
			this.nodeClickCallBack(d);
		}
	}

	initSvg(svgID) {
		// Add attributes to root svg
		this.rootSvg = d3Select(`#${svgID}`).on('click', () => this.svgClick());
		this.rootGroup = this.rootSvg.append('g').attr('class', 'root-group');
		this.lineGroup = this.rootGroup.append('g');
		this.nodeGroup = this.rootGroup.append('g');

		// Add zoom and drag
		this.zoom = zoom()
			.on('zoom', event => {
				if (this.svgZoomCallback !== null) {
					this.svgZoomCallback();
				}

				this.rootGroup.attr('transform', event.transform);
			})
			.scaleExtent([0.25, 8]);
		this.rootSvg.call(this.zoom);

		// Add arrow definition
		const defs = this.rootSvg.append('svg:defs');

		// Set pattern and marker
		defs.node().innerHTML
      = `<pattern id="striped" viewBox="0,0,4,4" width="40%" height="40%">
          <rect width="4" height="4" fill="rgb(var(--v-theme-primary))" />
          <path d="M-1,1 l2,-2 M0,4 l4,-4 M3,5 l2,-2" style="stroke:black; stroke-width:1.5 "/>
        </pattern>
        <marker id="arrowhead" viewBox="0 -5 10 10" refX="9" refY="0" markerWidth="10" markerHeight="10" orient="auto">
            <path d="M0,-5L10,0L0,5" fill="#999"/>
        </marker>`;
	}

	// Creates links based on the given nodes
	getLinks(nodes) {
		const links = [];
		nodes.forEach(d => {
			if (!d.children) {
				return;
			}

			d.children.forEach(child => {
				links.push({source: child, target: d.uid});
			});
		});

		return links;
	}

	checkNode(node) {
		if (!node.uid || !node.type) {
			throw new Error('node does not have required attributes: uid, type');
		}
	}

	// Adds the given node to the graph. If a node with the
	// provided node.uid already exist the existing node is instead updated.
	// Set draw to false, if the graph should not be redrawn.
	addNode(node, draw) {
		this.checkNode(node);

		// Check if properties have to be copied
		let mapNode = this.nodeMap.get(node.uid);
		if (mapNode === undefined) {
			mapNode = node;
		} else {
			Object.assign(mapNode, node);
		}

		this.nodeMap.set(node.uid, setFxFy(mapNode));
		if (draw === undefined || draw === true) {
			this.draw();
		}
	}

	// Adds the given nodes to the graph. Nodes which have an
	// already existing UID are instead updated.
	// Set draw to false, if the graph should not be redrawn.
	addNodes(nodes, draw) {
		nodes.forEach(node => {
			this.addNode(node, false);
		});
		if (draw === undefined || draw === true) {
			this.draw();
		}
	}

	drawNode(groupElement) {
		const self = this;
		groupElement.append('circle')
			.attr('class', 'node')
			.attr('r', this.nodeRadius)
			.attr('stroke', 'white')
			.attr('stroke-width', 1)
			.attr('cursor', 'pointer')
			.attr('fill', d => {
				if (d.type === 'transaction') {
					return 'url(#striped)';
				}

				if (d.type === 'heuristic') {
					return 'orange';
				}

				return 'green';
			})
			.each(function (d) {
				// Add marker to new nodes
				if (d.fx !== undefined) {
					return;
				}

				const thisElement = d3Select(this.parentNode);

				const marker = thisElement.append('circle')
					.attr('r', self.nodeRadius * 2)
					.attr('fill', 'rgba(var(--v-theme-primary),0.1)')
					.lower();

				marker.transition().delay(1000).duration(500).attr('r', 0).remove();
			})
			.on('click', function (e, d) {
				self.nodeClick(e, d, this);
			})
			.on('contextmenu', function (e, d) {
				if (self.contextMenuCallback !== null) {
					self.contextMenuCallback(e);
				}

				self.activeContextMenuNode = d;
				self.activeContextMenuSelection = d3Select(this);
			})
			.on('mouseenter', function () {
				d3Select(this.parentNode).raise();
				d3Select(this).transition().duration(100).attr('r', self.nodeRadius * 1.2);
			})
			.on('mouseleave', function () {
				d3Select(this).transition().duration(100).attr('r', self.nodeRadius);
			});

		const loadingRadius = this.nodeRadius - 4;
		const gap = 2 * Math.PI * loadingRadius / 4;

		const gapString = `${gap} ${gap}`;

		groupElement.each(function (d) {
		  if (d.status !== 'loading') {
				return;
			}

			d3Select(this).append('circle')
				.attr('cx', 0)
				.attr('cy', 0)
				.attr('r', loadingRadius)
				.attr('stroke-width', 3)
				.attr('stroke', '#fff')
				.attr('stroke-dasharray', gapString)
				.attr('fill', 'none')
				.attr('stroke-linecap', 'round')
				.append('animateTransform')
				.attr('attributeName', 'transform')
				.attr('type', 'rotate')
				.attr('repeatCount', 'indefinite')
				.attr('dur', '2.941176470588235s')
				.attr('keyTimes', '0;1')
				.attr('values', '0 0 0;360 0 0');
		});

		const textAreaWidth = 150;
		const textAreaMargin = 3;
		const textHeight = 12;
		const fontSize = textHeight - 2;

		function wrap() {
			const self = d3Select(this);
			let textLength = self.node().getComputedTextLength();
			let text = self.text();
			while (textLength > (textAreaWidth) && text.length > 0) {
				text = text.slice(0, -1);
				self.text(text + '...');
				textLength = self.node().getComputedTextLength();
			}
		}

		groupElement
			.append('text')
			.attr('font-size', fontSize)
			.style('text-anchor', 'middle')
			.style('cursor', 'default')
			.attr('fill', 'currentColor')
			.attr('stroke-width', 0)
			.attr('y', this.nodeRadius + textHeight + textAreaMargin)
			.text(d => `${d.uid}`)
			.each(wrap);
		groupElement
			.append('text')
			.attr('font-size', fontSize)
			.style('text-anchor', 'middle')
			.style('cursor', 'default')
			.attr('fill', 'currentColor')
			.attr('stroke-width', 0)
			.attr('y', this.nodeRadius + textHeight * 2 + textAreaMargin)
			.text(d => `${d.type}`)
			.each(wrap);
	}

	draw() {
		// If there is a simulation ongoing from a previous call, stop it
		if (this.simulation) {
			this.simulation.stop();
		}

		const nodes = [...this.nodeMap.values()];
		const links = this.getLinks(nodes);

		this.simulation = forceSimulation(nodes)
			.force('link', forceLink(links).id(d => d.uid))
			.force('charge', forceManyBody().strength(-50))
			.force('collide', forceCollide(this.nodeRadius * 5)).stop();

		// Do simulation
		this.simulation.tick(Math.ceil(Math.log(this.simulation.alphaMin()) / Math.log(1 - this.simulation.alphaDecay())));

		const link = this.lineGroup
			.selectAll('.arrow')
			.data(links, d => `${d.source}${d.target}`)
			.join('line')
			.attr('x1', d => d.source.x)
			.attr('y1', d => d.source.y)
			.attr('x2', d => reduceX(d, this.nodeRadius))
			.attr('y2', d => reduceY(d, this.nodeRadius))
			.attr('class', 'arrow')
			.attr('stroke', '#999')
			.attr('stroke-opacity', 1)
			.attr('marker-end', 'url(#arrowhead)')
			.attr('stroke-width', 1);

		const self = this;
		const node = this.nodeGroup
			.selectAll('.nodeContainer')
			.data(nodes, d => d.uid)
			.join(enter => {
				const g = enter.append('g');

				this.drawNode(g);
				return g;
			})
			.attr('class', 'nodeContainer')
			.call(drag()
				.on('start', e => {
					dragStarted(e, self.simulation);
				})
				.on('drag', function (e) {
					dragged(e, this);
				})
				.on('end', e => {
					dragEnded(e, self.simulation);
				}),
			)
			.each(d => {
				// Exclude every node from force simulation
				d.fx = d.x;
				d.fy = d.y;
			})
			.attr('transform', d => `translate(${d.x},${d.y})`);

		this.simulation.on('tick', () => {
			link
				.attr('x1', d => d.source.x)
				.attr('y1', d => d.source.y)
				.attr('x2', d => reduceX(d, this.nodeRadius))
				.attr('y2', d => reduceY(d, this.nodeRadius));

			node.attr('transform', d => `translate(${d.x},${d.y})`);
		});
	}

	// CenterGraph centers the graph in the center of the svg
	async centerGraph() {
		const svgBoundingRect = this.rootSvg.node().getBoundingClientRect();
		const rgBoundingBox = this.rootGroup.node().getBBox();
		const rgBoundingRect = this.rootGroup.node().getBoundingClientRect();

		const scaleHeight = svgBoundingRect.height / (rgBoundingRect.height * 1.5);
		const scaleWidth = svgBoundingRect.width / (rgBoundingRect.width * 1.5);

		this.rootSvg.call(this.zoom.translateTo, rgBoundingBox.x + rgBoundingBox.width / 2, rgBoundingBox.y + rgBoundingBox.height / 2);
		// Scale between 0.1 and 5
		this.rootSvg.call(this.zoom.scaleBy, Math.max(Math.min(scaleHeight, scaleWidth, 5), 0.1));
	}

	// Returns all nodes with their attached attributes from the force simulation performed in draw()
	exportNodes() {
		const nodes = [...this.nodeMap.values()];
		return nodes.map(d => {
			// Remove redundant attributes
			delete d.vx;
			delete d.vy;
			delete d.index;
			delete d.fx;
			delete d.fy;

			// Reduce precision to reduce space requirements
			d.x = Math.round(d.x * 10000) / 10000;
			d.y = Math.round(d.y * 10000) / 10000;
			return d;
		});
	}

	setNodeClickHandler(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.nodeClickCallBack = callback;
		return true;
	}

	// SetZoomCallback receives a function as an argument.
	// The function is going to be called each time the root SVG zoomed upon
	setSvgZoomCallback(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.svgZoomCallback = callback;
		return true;
	}

	// SetSvgClickCallback receives a function as an argument.
	// The function is going to be called each time the root SVG is clicked
	setSvgClickCallback(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.svgClickCallback = callback;
		return true;
	}

	// SetContextMenuCallback receives a function as an argument.
	// The function is going to be called each time the context menu is activated.
	setContextMenuCallback(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.contextMenuCallback = callback;
		return true;
	}
}
