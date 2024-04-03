import {getPrivacyTypeLabel, isFunction} from '@/utilities';
import {drag} from 'd3-drag';
import {select as d3Select} from 'd3-selection';
import {zoom} from 'd3-zoom';
import {
	forceCollide, forceLink, forceManyBody, forceSimulation,
} from 'd3-force';
import {abbreviateNumber} from '@/d3Documents/util';
import {
	mdiClockAlertOutline, mdiMerge, mdiPlaylistRemove, mdiTune,
} from '@mdi/js';
import forceLimit from '@/d3Documents/forceLimit';
import {
	WORKSPACE_NODE_TYPE_CLUSTER,
	WORKSPACE_NODE_TYPE_HEURISTIC,
	WORKSPACE_NODE_TYPE_TRANSACTION,
} from '@/constants/index.js';

// Sets a node with a valid x attribute to be excluded from force simulations
function setFxFy(node) {
	if (node.x !== undefined) {
		node.fx = node.x;
		node.fy = node.y;
	}

	return node;
}

function dragStarted(event, context) {
	if (!context.enableInteractions) {
		return;
	}

	if (!event.active) {
		context.simulation.alphaTarget(0.3).restart();
	}

	event.subject.fx = event.subject.x;
	event.subject.fy = event.subject.y;
	context.dragStartX = event.subject.x;
	context.dragStartY = event.subject.y;
}

function dragged(event, context, d3This) {
	if (!context.enableInteractions) {
		return;
	}

	event.subject.fx = event.x;
	event.subject.fy = event.y;
	// Raise() causes bug in chrome: click is only
	// recognized on second time. moved here from dragStart
	d3Select(d3This).raise();
}

function dragEnded(event, context) {
	if (!context.enableInteractions) {
		return;
	}

	if (!event.active) {
		context.simulation.alphaTarget(0);
	}

	// Call callback only if dragged at least the minimum distance
	if (context.dragEndCallback !== null
    && (Math.abs(context.dragStartX - event.x) > 3 || Math.abs(context.dragStartY - event.y) > 3)) {
		context.dragEndCallback();
	}
}

export default class NodeGraph {
	constructor(nodeTypeColorMap) {
		this.nodeClickCallBack = null;
		this.svgZoomCallback = null;
		this.svgClickCallback = null;
		this.contextMenuCallback = null;

		// Drag
		this.dragEndCallback = null;
		this.dragStartX = 0;
		this.dragStartY = 0;

		// Context node, set when a node is clicked or the contextmenu is shown
		this.contextNodeData = null;
		this.contextNodeSelection = null;

		// Svg
		this.simulation = null;
		this.nodeRadius = 14;
		this.rootSvg = null;
		this.rootGroup = null;
		this.lineGroup = null;
		this.nodeGroup = null;
		this.zoom = null;
		this.newNodes = null;

		// Data
		this.nodeMap = new Map();
		this.changedData = new Map();

		// Heuristic type map
		this.heuristicTypeMap = new Map();

		// Node type
		this.nodeTypeColorMap = nodeTypeColorMap;

		this.enableInteractions = true;
	}

	setEnableInteractions(flag) {
		this.enableInteractions = flag;
	}

	getEnableInteractions() {
		return this.enableInteractions;
	}

	svgClick() {
		this.resetClick();
		if (this.svgClickCallback !== null) {
			this.svgClickCallback();
		}
	}

	resetClick() {
		this.nodeGroup.selectAll('.clicked').classed('clicked', false);
	}

	setContextNodeClicked() {
		this.resetClick();
		d3Select(this.contextNodeSelection).select('.node').classed('clicked', true);
	}

	nodeClick(e, d, d3This) {
		if (e) {
			e.stopPropagation();
		}

		if (!this.enableInteractions) {
			return;
		}

		this.contextNodeData = d;
		this.contextNodeSelection = d3This;

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
			.scaleExtent([0.5, 3]);
		this.rootSvg.call(this.zoom);

		const defs = this.rootSvg.append('svg:defs');

		// Set pattern and arrowhead.
		// Arrow is unused for now. In case it is used later on, use reduceY and
		// reduceX to reduce the length of the links (modify d.target.x and d.target.y)
		defs.node().innerHTML
      = `<pattern id="striped" viewBox="0,0,4,4" width="40%" height="40%">
          <rect width="4" height="4" fill="rgb(var(--v-theme-primary))" />
          <path d="M-1,1 l2,-2 M0,4 l4,-4 M3,5 l2,-2" style="stroke:black; stroke-width:1.5 "/>
        </pattern>
        <pattern id="checkers" viewBox="0,0,8,8" width="60%" height="60%" patternTransform="translate(0, -4)">
          <rect width="8" height="8" fill="rgb(var(--v-theme-primary))" />
          <path id="a" data-color="fill" fill="#000" d="M4 4h4v4H4zM0 0h4v4H0z"></path>
        </pattern>
        <marker id="arrowhead" viewBox="0 -5 10 10" refX="9" refY="0" markerWidth="10" markerHeight="10" orient="auto">
            <path d="M0,-5L10,0L0,5" fill="#999"/>
        </marker>`;

		const style = this.rootSvg.append('svg:style');
		style.node().innerHTML
      = `
        .clicked {
            stroke: #B71C1C;
            stroke-width: 3;
         }
    `;
	}

	// Creates links based on the given nodes
	getLinks(nodes) {
		const linkSet = new Set();

		const links = [];
		nodes.forEach(d => {
			if (!d.children) {
				return;
			}

			d.children.forEach(child => {
				if (!this.nodeMap.has(child)) {
					return;
				}

				// Check if link already exists
				if (linkSet.has(child + d.uid) || linkSet.has(d.uid + child)) {
					return;
				}

				links.push({source: child, target: d.uid});
				linkSet.add(child + d.uid);
			});
		});
		return links;
	}

	// CheckNode returns tur if both the UID and type of the node is set
	checkNode(node) {
		return Boolean(node.uid) && Boolean(node.type);
	}

	removeContextMenuNode() {
		if (this.contextNodeData?.uid && this.enableInteractions) {
			this.removeNode(this.contextNodeData.uid);
			this.contextNodeData = null;
			this.contextNodeSelection = null;
		}
	}

	// Removes the node with the provided UID.
	// Set draw to false, if the graph should not be redrawn.
	removeNode(uid, draw) {
		this.nodeMap.delete(uid);

		if (draw === undefined || draw === true) {
			this.draw();
		}
	}

	// Removes the nodes with the provided UIDs.
	// Set draw to false, if the graph should not be redrawn.
	removeNodes(uids, draw) {
		uids.forEach(u => this.nodeMap.delete(u));

		if (draw === undefined || draw === true) {
			this.draw();
		}
	}

	reorderNodes() {
		for (const [key, value] of this.nodeMap) {
			delete value.x;
			delete value.y;
			delete value.fx;
			delete value.fy;
			this.nodeMap[key] = value;
		}

		this.draw();
		this.centerGraph();
	}

	// Adds the given node. If a node with the
	// provided node.uid already exist the existing node is instead updated.
	// Set draw to false, if the graph should not be redrawn.
	addNode(node, draw) {
		if (!this.checkNode(node)) {
			// Skip node if it has errors
			return;
		}

		// Check if properties have to be copied
		const mapNode = this.nodeMap.get(node.uid);
		if (mapNode !== undefined) {
			node.x = mapNode.x;
			node.y = mapNode.y;
		}

		const n = setFxFy(node);
		this.nodeMap.set(n.uid, n);
		this.changedData.set(n.uid, n);
		if (draw === undefined || draw === true) {
			this.draw();
		}
	}

	// Remove all nodes. Optionally redraw the graph.
	removeAllNodes(draw) {
		this.nodeMap.clear();
		if (draw === undefined || draw === true) {
			this.draw();
		}
	}

	// Adds the given nodes. Nodes which have an
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

	// Returns the node specified node. If the node does not
	// exist in the graph undefined is returned.
	getNode(uid) {
		return this.nodeMap.get(uid);
	}

	// Returns the node's parent. If the node does not
	// exist in the graph undefined is returned.
	getParent(uid) {
		const node = this.nodeMap.get(uid);
		if (!node) {
			return undefined;
		}

		const nodes = this.getNodes();
		return nodes.find(v => v.children?.includes(uid));
	}

	getNodes() {
		return Array.from(this.nodeMap.values());
	}

	centerOnNewNodes() {
		this.centerOnSelection(this.newNodes);
	}

	centerOnSelection(selection) {
		if (selection === null) {
			return;
		}

		let maxX = null;
		let minX = null;
		let maxY = null;
		let minY = null;
		selection.each(d => {
			if (maxX === null || maxX < d.x) {
				maxX = d.x;
			}

			if (minX === null || minX > d.x) {
				minX = d.x;
			}

			if (maxY === null || maxY < d.y) {
				maxY = d.y;
			}

			if (minY === null || minY > d.y) {
				minY = d.y;
			}
		});

		// Check if at least one element is in selection
		if (maxX === null) {
			return;
		}

		const width = maxX - minX;
		const height = maxY - minY;
		const centerX = minX + width / 2;
		const centerY = minY + height / 2;

		this.rootSvg.transition().duration(250).call(this.zoom.translateTo, centerX, centerY);
	}

	drawIcons(groupElement, icons, parameter) {
		if (icons.length === 0) {
			return;
		}

		let iconGroup = groupElement.select('.iconGroup');
		if (iconGroup.empty()) {
			iconGroup = groupElement.append('g').attr('class', 'iconGroup');
		}

		const textAreaMargin = 3;
		const textHeight = 12;
		const iconWidth = 12;
		const iconMargin = 1;
		const iconY = this.nodeRadius + textHeight + textAreaMargin * 2;

		// Remove all children
		iconGroup.selectAll('*').remove();

		icons.forEach((icon, i) => {
			iconGroup.append('path')
				.attr('transform', `translate(${iconWidth * i + iconMargin * i},${iconY}) scale(0.45,0.45)`)
				.attr('fill', 'currentColor')
				.attr('d', icon);
		});

		if (parameter) {
			iconGroup.append('text')
				.attr('transform', `translate(${iconWidth * icons.length + iconMargin * icons.length},${iconY + 9})`)
				.attr('font-size', 10)
				.style('cursor', 'default')
				.attr('fill', 'currentColor')
				.text(parameter);
		}

		const groupWidth = iconGroup.node().getBBox().width;
		iconGroup.attr('transform', `translate(${-groupWidth / 2},0)`);
	}

	drawNode(groupElement) {
		const self = this;
		// CircleGroup contains the node circle and loading circle

		let circleGroup = groupElement.select('g');
		if (circleGroup.empty()) {
			circleGroup = groupElement.append('g');
		}

		circleGroup.selectAll('circle').remove();

		// Node circle
		circleGroup.append('circle')
			.attr('class', 'node')
			.attr('r', this.nodeRadius)
			.attr('stroke', 'currentColor')
			.attr('stroke-width', 1)
			.attr('cursor', 'pointer')
			.attr('fill', d => {
				if (this.nodeTypeColorMap) {
					let nodeColor;

					if (d.privacyType) {
						nodeColor = this.nodeTypeColorMap.get(getPrivacyTypeLabel(d.privacyType));
					} else {
						nodeColor = this.nodeTypeColorMap.get(d.type);
					}

					if (nodeColor) {
						if (nodeColor === 'striped' || nodeColor === 'checkers') {
							return `url(#${nodeColor})`;
						}

						return nodeColor;
					}
				}

				return 'rgb(var(--v-theme-primary))';
			})
			.each(function (d) {
				// Add marker to new nodes
				if (d.fx !== undefined) {
					return;
				}

				const thisElement = d3Select(this.parentNode);

				const marker = thisElement.append('circle')
					.attr('r', self.nodeRadius * 2)
					.attr('fill', 'rgba(255, 109, 0, 0.3)')
					.lower();

				marker.transition().delay(1000).duration(500).attr('r', 0).remove();
			});

		// Set event handlers
		circleGroup
			.on('click', function (e, d) {
				self.nodeClick(e, d, this);
			})
			.on('contextmenu', function (e, d) {
				if (!self.enableInteractions) {
					return;
				}

				self.contextNodeData = d;
				self.contextNodeSelection = this;

				if (self.contextMenuCallback !== null) {
					self.contextMenuCallback(e, d);
				}
			})
			.on('mouseenter', function () {
				if (!self.enableInteractions) {
					return;
				}

				d3Select(this.parentNode).raise();
				d3Select(this).select('.node').transition().duration(100).attr('r', self.nodeRadius * 1.2);
			})
			.on('mouseleave', function () {
				if (!self.enableInteractions) {
					return;
				}

				d3Select(this).select('.node').transition().duration(100).attr('r', self.nodeRadius);
			});

		// Add loading circle
		const loadingRadius = this.nodeRadius - 6;
		const gap = 2 * Math.PI * loadingRadius / 4;

		const gapString = `${gap} ${gap}`;

		circleGroup.each(function (d) {
			if (!d.loading) {
				return;
			}

			d3Select(this).append('circle')
				.attr('r', loadingRadius)
				.attr('cursor', 'pointer')
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

		// Add node descriptions
		const textAreaWidth = 100;
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

		// Text container
		let textContainer = groupElement.select('.textContainer');
		if (textContainer.empty()) {
			textContainer = groupElement.append('g').attr('class', 'textContainer');
		}

		// Node title
		let nodeTitle = textContainer.select('.nodeTitle');
		if (nodeTitle.empty()) {
			nodeTitle = textContainer.append('text').attr('class', 'nodeTitle');
		}

		nodeTitle
			.attr('font-size', fontSize)
			.attr('text-anchor', 'middle')
			.style('cursor', 'default')
			.attr('fill', 'currentColor')
			.attr('y', this.nodeRadius + textHeight + textAreaMargin)
			.text(d => {
				if (d.type === WORKSPACE_NODE_TYPE_CLUSTER) {
					return d.addressHash;
				}

				if (d.type === WORKSPACE_NODE_TYPE_TRANSACTION) {
					return d.transactionHash;
				}

				if (d.type === WORKSPACE_NODE_TYPE_HEURISTIC) {
					const title = this.heuristicTypeMap.get(d.heuristicType);
					if (title !== undefined) {
						return title;
					}
				}

				return d.uid;
			})
			.each(wrap);

		let nodeSubtitle = textContainer.select('.nodeSubtitle');
		if (nodeSubtitle.empty()) {
			nodeSubtitle = textContainer.append('text').attr('class', 'nodeSubtitle');
		}

		nodeSubtitle
			.attr('font-size', fontSize)
			.attr('text-anchor', 'middle')
			.style('cursor', 'default')
			.attr('fill', 'currentColor')
			.attr('y', this.nodeRadius + textHeight * 2 + textAreaMargin)
			.text(d => {
				if (d.type === WORKSPACE_NODE_TYPE_TRANSACTION && d.privacyType) {
					return getPrivacyTypeLabel(d.privacyType);
				}

				return '';
			})
			.each(wrap);

		// Heuristic properties
		// Cluster count
		let nodeClusterCount = circleGroup.select('.clusterCount');
		if (nodeClusterCount.empty()) {
			nodeClusterCount = circleGroup.append('text').attr('class', 'clusterCount');
		}

		nodeClusterCount.raise();

		nodeClusterCount
			.attr('text-anchor', 'middle')
			.style('cursor', 'pointer')
			.style('font-weight', 'bold')
			.attr('dominant-baseline', 'middle')
			.attr('fill', 'white')
			.attr('font-size', 12)
			.attr('y', 1)
			.text(d => {
				if (d.type !== WORKSPACE_NODE_TYPE_HEURISTIC) {
					return '';
				}

				return abbreviateNumber(d.heuristicClusterCount);
			});

		textContainer
			.each(function (d) {
				if (d.type !== WORKSPACE_NODE_TYPE_HEURISTIC) {
					return;
				}

				const icons = [];
				if (d.heuristicExcludeAddresses) {
					icons.push(mdiPlaylistRemove);
				}

				if (d.heuristicClusterTypes?.length > 0) {
					icons.push(mdiMerge);
				}

				if (d.heuristicExcludeSpendingGaps) {
					icons.push(mdiClockAlertOutline);
				}

				if (d.heuristicParameter) {
					icons.push(mdiTune);
				}

				self.drawIcons(d3Select(this), icons, d.heuristicParameter);
			});
	}

	// Draws the state of the graph, returns all newly added nodes
	draw() {
		// If there is a simulation ongoing from a previous call, stop it
		if (this.simulation) {
			this.simulation.stop();
		}

		const nodes = [...this.nodeMap.values()];
		const links = this.getLinks(nodes);

		const svgRect = this.rootSvg.node().getBoundingClientRect();

		this.simulation = forceSimulation(nodes)
			.force('link', forceLink(links).id(d => d.uid))
			.force('charge', forceManyBody().strength(-150))
			.force('collide', forceCollide(this.nodeRadius * 4))
			.force('limit', forceLimit().x0(0).x1(svgRect.width).y0(0).y1(svgRect.height).radius(this.nodeRadius)).stop();

		// Do simulation
		this.simulation.tick(Math.ceil(Math.log(this.simulation.alphaMin()) / Math.log(1 - this.simulation.alphaDecay())));

		const link = this.lineGroup
			.selectAll('.arrow')
			.data(links, d => `${d.source}${d.target}`)
			.join('line')
			.attr('class', 'arrow')
			.attr('stroke', 'currentColor')
			.attr('stroke-opacity', 1)
			.attr('stroke-width', 1)
			.attr('x1', d => d.source.x)
			.attr('y1', d => d.source.y)
			.attr('x2', d => d.target.x)
			.attr('y2', d => d.target.y);

		const self = this;
		const node = this.nodeGroup
			.selectAll('.nodeContainer')
			.data(nodes, d => d.uid)
			.join(enter => {
				const g = enter.append('g');

				this.drawNode(g);
				this.newNodes = g;
				return g;
			},
			update => {
				if (this.changedData.size > 0) {
					// Do drawing only for actually updated nodes
					this.drawNode(update.filter(d => this.changedData.has(d.uid)));
				}

				return update;
			})
			.call(drag()
				.on('start', e => {
					dragStarted(e, self);
				})
				.on('drag', function (e) {
					dragged(e, self, this);
				})
				.on('end', e => {
					dragEnded(e, self);
				})
				.clickDistance(3),
			)
			.attr('class', 'nodeContainer')
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
				.attr('x2', d => d.target.x)
				.attr('y2', d => d.target.y);

			node.attr('transform', d => `translate(${d.x},${d.y})`);
		});

		this.changedData.clear();
	}

	// CenterGraph centers the graph in the center of the svg
	centerGraph() {
		const svgBoundingRect = this.rootSvg.node().getBoundingClientRect();
		const rgBoundingBox = this.rootGroup.node().getBBox();
		const rgBoundingRect = this.rootGroup.node().getBoundingClientRect();

		// Calculate scaling, reduce the svg size so the root group is scaled slightly smaller than the svg size
		const scaleHeight = (svgBoundingRect.height - 120) / rgBoundingRect.height;
		const scaleWidth = (svgBoundingRect.width - 100) / rgBoundingRect.width;

		this.rootSvg.call(this.zoom.translateTo, rgBoundingBox.x + rgBoundingBox.width / 2, rgBoundingBox.y + rgBoundingBox.height / 2);

		const scaleBy = Math.min(scaleHeight, scaleWidth);

		// Return if scaling is negligible
		if (Math.abs(1 - scaleBy) < 0.1) {
			return;
		}

		this.rootSvg.call(this.zoom.scaleBy, scaleBy);
	}

	// Returns all nodes with their attached attributes from the force simulation performed in draw()
	exportNodes() {
		const nodes = structuredClone([...this.nodeMap.values()]);
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

	// Returns the node which triggered the context menu event or click event
	getContextNode() {
		return this.contextNodeData;
	}

	// SetDragCallback receives a function as an argument.
	// The function is going to be called after each drag event
	setDragEndCallback(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.dragEndCallback = callback;
		return true;
	}

	populateHeuristicMap(heuristicDescriptions) {
		// Titles to map
		heuristicDescriptions.forEach(e => this.heuristicTypeMap.set(e.type, e.title));
	}
}
