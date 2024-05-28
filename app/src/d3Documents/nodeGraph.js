import {isFunction} from '@/utilities';
import {drag} from 'd3-drag';
import {select as d3Select} from 'd3-selection';
import {zoom} from 'd3-zoom';
import {forceCollide, forceLink, forceSimulation} from 'd3-force';
import {
	abbreviateNumber, reduceX, reduceY, reduceXR, reduceYR,
} from '@/d3Documents/util';
import {
	mdiClockAlertOutline, mdiMerge, mdiPlaylistRemove, mdiTune,
} from '@mdi/js';
import forceLimit from '@/d3Documents/forceLimit';
import {
	WORKSPACE_NODE_TYPE_CLUSTER,
	WORKSPACE_NODE_TYPE_HEURISTIC, WORKSPACE_NODE_TYPE_NOTE,
	WORKSPACE_NODE_TYPE_TRANSACTION,
} from '@/constants/index.js';
import d3lasso from './d3Lasso.js';

// In ms
const animationDuration = 175;
const longAnimationDuration = 500;
const animationDelay = 2000;

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

	if (context.lassoSelectedNodes) {
		const lassoData = context.lassoSelectedNodes.data();
		// Reset the lasso selection if it has only one element or if the root node is not in the lasso selection
		if (lassoData.length === 1 || !lassoData.some(d => d.uid === event.subject.uid)) {
			context.resetLasso();
		}
	}

	event.subject.fx = event.subject.x;
	event.subject.fy = event.subject.y;
	context.dragStartX = event.subject.x;
	context.dragStartY = event.subject.y;
}

function dragged(event, context, d3This, data) {
	if (!context.enableInteractions) {
		return;
	}

	event.subject.fx = event.x;
	event.subject.fy = event.y;

	if (context.lassoSelectedNodes) {
		context.lassoSelectedNodes.each(d => {
			// Don't change the actual dragged ndoe
			if (d.uid === data.uid) {
				return;
			}

			// Set selected node positions
			d.fx += event.dx;
			d.fy += event.dy;
		});
	}

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
	// Callbacks
	#nodeClickCallBack = null;
	#lineClickCallBack = null;
	#svgZoomCallback = null;
	#svgClickCallback = null;
	#contextMenuCallback = null;
	#lassoSelectionCallback = null;
	#lassoResetCallback = null;
	// Drag
	dragEndCallback = null;
	dragStartX = 0;
	dragStartY = 0;
	// Lasso
	#isLassoEnabled = false;
	#lasso = null;
	lassoSelectedNodes = null;
	// Context node, set when a node is clicked or the contextmenu is shown
	#contextNodeData = null;
	#contextNodeSelection = null;
	// Svg
	#svgID = '';
	simulation = null;
	#nodeRadius = 14;
	#rootSvg = null;
	#rootGroup = null;
	#lineGroup = null;
	#shadowLineGroup = null;
	#nodeGroup = null;
	#zoom = null;
	#newNodes = null;
	// Data
	#nodeMap = new Map();
	#filteredNodeMap = new Map();
	#filterNodeTypes = [];
	#filterPrivacyTypes = [];
	#changedData = new Map();
	// Heuristic type map
	#heuristicTypeMap = new Map();
	// Node type
	#nodeTypeColorMap = null;
	enableInteractions = true;

	constructor(nodeTypeColorMap) {
		this.#nodeTypeColorMap = nodeTypeColorMap;
	}

	setEnableInteractions(flag) {
		this.enableInteractions = flag;
	}

	getFilteredMap() {
		if (this.#filterNodeTypes.length > 0) {
			return this.#filteredNodeMap;
		}

		return this.#nodeMap;
	}

	resetNodeFilter() {
		this.#filterNodeTypes = [];
		this.#filteredNodeMap.clear();
	}

	isAllowedByFilter(node) {
		if (!node) {
			return false;
		}

		if (node.type === WORKSPACE_NODE_TYPE_NOTE) {
			if (!node.children) {
				return false;
			}

			return node.children.some(child => this.isAllowedByFilter(this.#nodeMap.get(child)));
		}

		let allowed = false;
		if (this.#filterNodeTypes.length > 0) {
			allowed = this.#filterNodeTypes.includes(node.type);
		} else {
			allowed = true;
		}

		if (!allowed) {
			return false;
		}

		if (node.type === WORKSPACE_NODE_TYPE_TRANSACTION) {
			return this.#filterPrivacyTypes.includes(node.privacyTypeLabel);
		}

		return true;
	}

	filterNodes(nodeTypes, privacyTypes) {
		if (nodeTypes) {
			this.#filterNodeTypes = nodeTypes;
		} else {
			this.#filterNodeTypes = [];
		}

		if (privacyTypes) {
			this.#filterPrivacyTypes = privacyTypes;
		} else {
			this.#filterPrivacyTypes = [];
		}

		this.#filteredNodeMap.clear();
		if (!nodeTypes) {
			return;
		}

		const entries = this.#nodeMap.entries();
		for (const entry of entries) {
			if (this.isAllowedByFilter(entry[1])) {
				this.#filteredNodeMap.set(entry[0], entry[1]);
			}
		}
	}

	getEnableInteractions() {
		return this.enableInteractions;
	}

	svgClick() {
		this.resetClick();
		this.resetLasso();
		if (this.#svgClickCallback !== null) {
			this.#svgClickCallback();
		}
	}

	resetClick() {
		this.#nodeGroup.selectAll('.clicked').classed('clicked', false);
		this.#shadowLineGroup.selectAll('.lineClicked').classed('lineClicked', false);
	}

	resetLasso() {
		this.#nodeGroup.selectAll('.lasso-selected').classed('lasso-selected', false);
		this.lassoSelectedNodes = null;

		if (this.#lassoResetCallback !== null) {
			this.#lassoResetCallback();
		}
	}

	setContextObjectClicked() {
		this.resetClick();
		this.resetLasso();

		const contextNode = d3Select(this.#contextNodeSelection);

		// Try selecting the active object, can be a node or a line
		if (contextNode.classed('shadowArrow')) {
			contextNode.classed('lineClicked', true);
		} else {
			contextNode.select('.node').classed('clicked', true);
		}
	}

	nodeClick(e, d, d3This) {
		if (e) {
			e.stopPropagation();
		}

		if (!this.enableInteractions) {
			return;
		}

		if (e.ctrlKey || e.shiftKey) {
			const n = d3Select(d3This).select('.node');
			if (n.classed('lasso-selected')) {
				n.classed('lasso-selected', false);
			} else {
				n.classed('lasso-selected', true);
			}

			this.lassoSelectedNodes = this.#nodeGroup.selectAll('.lasso-selected');
			if (this.#lassoSelectionCallback !== null) {
				this.#lassoSelectionCallback();
			}

			return;
		}

		this.#contextNodeData = d;
		this.#contextNodeSelection = d3This;

		if (this.#nodeClickCallBack !== null) {
			this.#nodeClickCallBack(d);
		}
	}

	noteClick(e, _, d3This) {
		if (e) {
			e.stopPropagation();
		}

		if (!this.enableInteractions || (!e.ctrlKey && !e.shiftKey)) {
			return;
		}

		const n = d3Select(d3This).select('.note');
		if (n.classed('lasso-selected')) {
			n.classed('lasso-selected', false);
		} else {
			n.classed('lasso-selected', true);
		}

		this.lassoSelectedNodes = this.#nodeGroup.selectAll('.lasso-selected');
		if (this.#lassoSelectionCallback !== null) {
			this.#lassoSelectionCallback();
		}
	}

	lineClick(e, d, d3This) {
		if (e) {
			e.stopPropagation();
		}

		if (!this.enableInteractions) {
			return;
		}

		this.#contextNodeData = d;
		this.#contextNodeSelection = d3This;

		if (this.#lineClickCallBack !== null) {
			this.#lineClickCallBack(d);
		}
	}

	setLassoEnabled(flag) {
		this.#isLassoEnabled = flag;
	}

	getLassoEnabled() {
		return this.#isLassoEnabled;
	}

	getLassoSelectedNodesData() {
		if (this.lassoSelectedNodes === null) {
			return [];
		}

		return this.lassoSelectedNodes.data();
	}

	initSvg(svgID, width, height) {
		// Add attributes to root svg
		this.#svgID = svgID;
		this.#rootSvg = d3Select(`#${svgID}`).on('click', () => this.svgClick());
		this.#rootGroup = this.#rootSvg.append('g').classed('root-group', true);
		this.#lineGroup = this.#rootGroup.append('g');
		this.#shadowLineGroup = this.#rootGroup.append('g');
		this.#nodeGroup = this.#rootGroup.append('g');

		this.virtualWidth = width;
		this.virutalHeight = height;

		// Add zoom and drag
		this.#zoom = zoom()
			.on('zoom', event => {
				if (this.#svgZoomCallback !== null) {
					this.#svgZoomCallback();
				}

				this.#rootGroup.attr('transform', event.transform);
			})
			.filter(e => ((!e.ctrlKey && !this.getLassoEnabled()) || e instanceof WheelEvent) && !e.button)
			.scaleExtent([0.5, 3]);
		this.#rootSvg.call(this.#zoom);

		// Add lasso
		const self = this;
		this.#lasso = d3lasso()
			.closePathDistance(2000)
			.closePathSelect(true)
			.dragFilter(e => e.ctrlKey || this.getLassoEnabled())
			.targetArea(this.#rootSvg)
			.on('draw', () => {
				self.#lasso.possibleItems().classed('lasso-selected', true);
				self.#lasso.notPossibleItems().classed('lasso-selected', false);
			})
			.on('end', () => {
				self.lassoSelectedNodes = self.#lasso.selectedItems();
				if (this.#lassoSelectionCallback !== null) {
					this.#lassoSelectionCallback();
				}
			});

		this.#rootSvg.call(this.#lasso);

		const defs = this.#rootSvg.append('svg:defs');

		// Set pattern and arrowhead.
		// Arrow is unused for now. In case it is used later on, use reduceY and
		// reduceX to reduce the length of the links (modify d.target.x and d.target.y)
		defs.node().innerHTML
      = `<marker id="${this.#svgID}_arrowhead" viewBox="0 -5 10 10" refX="0" refY="0" markerWidth="10" markerHeight="10" orient="auto">
            <path d="M0,-5L10,0L0,5" fill="currentColor"/>
        </marker>
        <marker id="${this.#svgID}_arrowhead_shadow" viewBox="0 -5 10 10" refX="1" refY="0" markerWidth="3" markerHeight="3" orient="auto">
            <path d="M0,-5L10,0L0,5" fill="rgb(var(--v-theme-primary))" />
        </marker>
        <marker id="${this.#svgID}_arrowhead_reversed" viewBox="-10 -5 10 10" refX="0" refY="0" markerWidth="10" markerHeight="10" orient="auto">
            <path d="M0,-5L-10,0L0,5" fill="currentColor" />
        </marker>
        <marker id="${this.#svgID}_arrowhead_reversed_shadow" viewBox="-10 -5 10 10" refX="-1" refY="0" markerWidth="3" markerHeight="3" orient="auto">
            <path d="M0,-5L-10,0L0,5" fill="rgb(var(--v-theme-primary))" />
        </marker>`;

		const style = this.#rootSvg.append('svg:style');
		style.node().innerHTML
      = `
        .node {
          stroke: currentColor;
          stroke-width: 1;
          cursor: pointer;
        }

        .note {
          stroke: currentColor;
          stroke-width: 1;
          fill: rgb(var(--v-theme-surface));
          cursor: pointer;
        }

        .note-text {
          fill: currentColor;
          font-size: 10px;
          text-anchor: middle;
          cursor: pointer;
        }

        .clicked {
          stroke: #B71C1C;
          stroke-width: 3;
         }

        .arrow {
          stroke: currentColor;
          stroke-opacity: 1;
          stroke-width: 1;
          marker-end: url(#${this.#svgID}_arrowhead);
        }

        .shadowArrow {
          cursor: pointer;
          stroke: rgb(var(--v-theme-primary));
          stroke-opacity: 1;
          stroke-width: 4;
          opacity: 0;
          marker-end: url(#${this.#svgID}_arrowhead_shadow);
          transition: all 0.175s ease;
        }

        .lineHovered {
          stroke-width: 5;
          opacity: 0.3;
        }

        .lineClicked {
          stroke-width: 5;
          opacity: 0.3;
        }

        .lasso-selected {
          stroke: rgb(var(--v-theme-primary));
          stroke-width: 3;
        }

        .lasso path {
            # stroke: rgb(80,80,80);
            stroke: rgb(var(--v-theme-primary));
            stroke-width: 2px;
        }

        .lasso .drawn {
            fill: rgb(var(--v-theme-primary));
            fill-opacity: 0.1 ;
        }

        .lasso .loop_close {
            fill: none;
            stroke-dasharray: 4, 4;
        }

        .lasso .origin {
            fill: #3399FF;
            fill-opacity: 0.5;
        }
    `;
	}

	// Creates links based on the given nodes
	getLinks(nodes) {
		const links = new Map();

		nodes.forEach(d => {
			if (!d.children) {
				return;
			}

			d.children.forEach(child => {
				if (!this.getFilteredMap().has(child)) {
					return;
				}

				// Check if link already exists
				if (links.has(d.uid + child)) {
					return;
				}

				// If reverse link exist, mark it as having both directions
				const reversedLink = links.get(child + d.uid);
				if (reversedLink !== undefined) {
					reversedLink.isDual = true;
					return;
				}

				links.set(d.uid + child, {source: d.uid, target: child});
			});
		});

		return Array.from(links.values());
	}

	// CheckNode returns tur if both the UID and type of the node is set
	checkNode(node) {
		return Boolean(node.uid) && Boolean(node.type);
	}

	removeContextMenuNode() {
		if (this.#contextNodeData?.uid && this.enableInteractions) {
			this.removeNode(this.#contextNodeData.uid);
			this.#contextNodeData = null;
			this.#contextNodeSelection = null;
		}
	}

	// Removes the node with the provided UID.
	// Set draw to false, if the graph should not be redrawn.
	removeNode(uid, draw) {
		this.#nodeMap.delete(uid);
		this.#filteredNodeMap.delete(uid);

		if (draw === undefined || draw === true) {
			this.draw();
		}
	}

	// Removes the nodes with the provided UIDs.
	// Set draw to false, if the graph should not be redrawn.
	removeNodes(uids, draw) {
		uids.forEach(u => this.removeNode(u, false));

		if (draw === undefined || draw === true) {
			this.draw();
		}
	}

	reorderNodes() {
		for (const [key, value] of this.getFilteredMap()) {
			// Randomize position, reorderNodes creates different arrangements for each call
			const random = Math.random() * 30;
			value.x = random;
			value.y = random;
			delete value.fx;
			delete value.fy;
			this.#nodeMap[key] = value;
			if (this.#filteredNodeMap[key] !== undefined) {
				this.#filteredNodeMap[key] = value;
			}
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
		const mapNode = this.#nodeMap.get(node.uid);
		if (mapNode !== undefined) {
			node.x = mapNode.x;
			node.y = mapNode.y;
		}

		const n = setFxFy(node);
		this.#nodeMap.set(n.uid, n);
		this.#changedData.set(n.uid, n);

		if (this.isAllowedByFilter(n)) {
			this.#filteredNodeMap.set(n.uid, n);
		}

		if (draw === undefined || draw === true) {
			this.draw();
		}
	}

	// Remove all nodes. Optionally redraw the graph.
	removeAllNodes(draw) {
		this.#nodeMap.clear();
		this.#filteredNodeMap.clear();
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
		return this.#nodeMap.get(uid);
	}

	getNodes() {
		return Array.from(this.#nodeMap.values());
	}

	centerOnNewNodes() {
		this.centerOnSelection(this.#newNodes);
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

		this.#rootSvg.transition().duration(250).call(this.#zoom.translateTo, centerX, centerY);
	}

	drawIcons(groupElement, icons, parameter) {
		if (icons.length === 0) {
			return;
		}

		let iconGroup = groupElement.select('.iconGroup');
		if (iconGroup.empty()) {
			iconGroup = groupElement.append('g').classed('iconGroup', true);
		}

		const textAreaMargin = 3;
		const textHeight = 12;
		const iconWidth = 12;
		const iconMargin = 1;
		const iconY = this.#nodeRadius + textHeight + textAreaMargin * 2;

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

	// Draws nodes and notes
	drawEntities(groupElement) {
		// CircleGroup contains the node circle and loading circle
		let entityGroup = groupElement.select('g');
		if (entityGroup.empty()) {
			entityGroup = groupElement.append('g');
		}

		this.drawNodes(groupElement.filter(d => d.type !== WORKSPACE_NODE_TYPE_NOTE),
			entityGroup.filter(d => d.type !== WORKSPACE_NODE_TYPE_NOTE));
		this.drawNotes(groupElement.filter(d => d.type === WORKSPACE_NODE_TYPE_NOTE),
			entityGroup.filter(d => d.type === WORKSPACE_NODE_TYPE_NOTE));
	}

	drawNotes(groupElement, entityGroup) {
		entityGroup.selectAll('.note,.note-text').remove();

		entityGroup.append('text')
			.classed('note-text', true)
			.each(function (d) {
				const textLines = d.text.split('\n');
				d3Select(this)
					.selectAll('tspan')
					.data(textLines)
					.enter()
					.append('tspan')
					.attr('x', 0)
					.attr('dy', '1.2em') // Line spacing
					.text(d => d ? d : ' '); // Insert space for empty row so vertical spacing works

				const nodeRect = this.getBBox();
				d.bbHeight = nodeRect.height;
				d.bbWidth = nodeRect.width;
				d3Select(this).attr('y', -nodeRect.height / 2 - 2);
			});

		entityGroup.append('rect')
			.classed('note', true)
			.attr('rx', 3)
			.attr('ry', 3)
			.lower()
			.each(function (d) {
				const rectMargin = 10;
				d.width = d.bbWidth + rectMargin;
				d.height = d.bbHeight + rectMargin;
				d3Select(this)
					.attr('width', d.width)
					.attr('height', d.height)
					.attr('x', -d.width / 2)
					.attr('y', -d.height / 2);

				// Add marker to new nodes
				if (d.fx !== undefined) {
					return;
				}

				const thisElement = d3Select(this.parentNode);

				const marker = thisElement.append('rect')
					.attr('width', d.width * 2)
					.attr('height', d.height * 2)
					.attr('x', -d.width)
					.attr('y', -d.height)
					.attr('rx', 3)
					.attr('ry', 3)
					.attr('fill', 'rgba(255, 109, 0, 0.3)')
					.lower();

				marker.transition().delay(animationDelay).duration(longAnimationDuration)
					.attr('width', 0)
					.attr('height', 0)
					.attr('x', 0)
					.attr('y', 0)
					.remove();
			});

		const self = this;
		// Set event handlers
		entityGroup
			.on('click', function (e, d) {
				self.noteClick(e, d, this);
			})
			.on('contextmenu', function (e, d) {
				if (!self.enableInteractions) {
					return;
				}

				self.#contextNodeData = d;
				self.#contextNodeSelection = this;

				if (self.#contextMenuCallback !== null) {
					self.#contextMenuCallback(e);
				}
			})
			.on('mouseenter', function () {
				if (!self.enableInteractions) {
					return;
				}

				d3Select(this.parentNode).raise();
				d3Select(this).select('.note').transition().duration(animationDuration)
					.attr('width', d => d.width * 1.2)
					.attr('height', d => d.height * 1.2)
					.attr('x', d => -d.width * 1.2 / 2)
					.attr('y', d => -d.height * 1.2 / 2);
			})
			.on('mouseleave', function () {
				if (!self.enableInteractions) {
					return;
				}

				d3Select(this).select('.note').transition().duration(animationDuration)
					.attr('width', d => d.width)
					.attr('height', d => d.height)
					.attr('x', d => -d.width / 2)
					.attr('y', d => -d.height / 2);
			});
	}

	drawNodes(groupElement, entityGroup) {
		const self = this;
		entityGroup.selectAll('circle').remove();

		// Node circle
		entityGroup.append('circle')
			.classed('node', true)
			.attr('r', this.#nodeRadius)
			.attr('fill', d => {
				if (this.#nodeTypeColorMap) {
					let nodeColor;

					if (d.privacyTypeLabel) {
						nodeColor = this.#nodeTypeColorMap.get(d.privacyTypeLabel);
					} else {
						nodeColor = this.#nodeTypeColorMap.get(d.type);
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
					.attr('r', self.#nodeRadius * 2)
					.attr('fill', 'rgba(255, 109, 0, 0.3)')
					.lower();

				marker.transition().delay(animationDelay).duration(longAnimationDuration).attr('r', 0).remove();
			});

		// Set event handlers
		entityGroup
			.on('click', function (e, d) {
				self.nodeClick(e, d, this);
			})
			.on('contextmenu', function (e, d) {
				if (!self.enableInteractions) {
					return;
				}

				self.#contextNodeData = d;
				self.#contextNodeSelection = this;

				if (self.#contextMenuCallback !== null) {
					self.#contextMenuCallback(e);
				}
			})
			.on('mouseenter', function () {
				if (!self.enableInteractions) {
					return;
				}

				d3Select(this.parentNode).raise();
				self.setMouseOverAnimation(self, this, true);
			})
			.on('mouseleave', function () {
				if (!self.enableInteractions) {
					return;
				}

				self.setMouseOverAnimation(self, this, false);
			});

		// Add loading circle
		const loadingRadius = this.#nodeRadius - 6;
		const gap = 2 * Math.PI * loadingRadius / 4;

		const gapString = `${gap} ${gap}`;

		entityGroup.each(function (d) {
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

		// Text container
		let textContainer = groupElement.select('.textContainer');
		if (textContainer.empty()) {
			textContainer = groupElement.append('g').classed('textContainer', true);
		}

		// Node title
		let nodeTitle = textContainer.select('.nodeTitle');
		if (nodeTitle.empty()) {
			nodeTitle = textContainer.append('text').classed('nodeTitle', true);
		}

		function elide() {
			const self = d3Select(this);
			let text = self.text();

			// Don't elide text which is 5 characters or smaller
			if (text.length <= 5) {
				return;
			}

			const selfNode = self.node();
			let textLength = selfNode.getComputedTextLength();

			// Reduce text by 10% each time and at minimum by 1
			const cutLength = Math.max(Math.floor(text.length / 10), 1);

			while (textLength > textAreaWidth && text.length > 0) {
				text = text.slice(0, -cutLength);
				// \u2026 = ...
				self.text(text + '\u2026');
				textLength = selfNode.getComputedTextLength();
			}
		}

		nodeTitle
			.attr('font-size', fontSize)
			.attr('text-anchor', 'middle')
			.style('cursor', 'default')
			.attr('fill', 'currentColor')
			.attr('y', this.#nodeRadius + textHeight + textAreaMargin)
			.text(d => {
				if (d.type === WORKSPACE_NODE_TYPE_CLUSTER) {
					return d.addressHash;
				}

				if (d.type === WORKSPACE_NODE_TYPE_TRANSACTION) {
					return d.transactionHash;
				}

				if (d.type === WORKSPACE_NODE_TYPE_HEURISTIC) {
					const title = this.#heuristicTypeMap.get(d.heuristicType);
					if (title !== undefined) {
						return title;
					}
				}

				return d.uid;
			})
			.each(elide);

		let nodeSubtitle = textContainer.select('.nodeSubtitle');
		if (nodeSubtitle.empty()) {
			nodeSubtitle = textContainer.append('text').classed('nodeSubtitle', true);
		}

		nodeSubtitle
			.attr('font-size', fontSize)
			.attr('text-anchor', 'middle')
			.style('cursor', 'default')
			.attr('fill', 'currentColor')
			.attr('y', this.#nodeRadius + textHeight * 2 + textAreaMargin)
			.text(d => {
				if (d.type === WORKSPACE_NODE_TYPE_TRANSACTION && d.privacyTypeLabel) {
					return d.privacyTypeLabel;
				}

				return '';
			})
			.each(elide);

		// Heuristic properties
		// Cluster count
		let nodeClusterCount = entityGroup.select('.clusterCount');
		if (nodeClusterCount.empty()) {
			nodeClusterCount = entityGroup.append('text').classed('clusterCount', true);
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

	setMouseOverAnimation(context, nodeContext, isEnter) {
		const thisNode = d3Select(nodeContext).select('.node');
		const nodeRadius = isEnter ? context.#nodeRadius * 1.2 : context.#nodeRadius;
		const opacity = isEnter ? 0.3 : 1.0;
		thisNode.transition().duration(animationDuration).attr('r', nodeRadius);

		const thisNodeUID = thisNode.data()[0].uid;
		context.#lineGroup.selectAll('.arrow')
			.filter(d => d.source.uid !== thisNodeUID && d.target.uid !== thisNodeUID)
			.each(function () {
				d3Select(this).transition().duration(animationDuration).attr('opacity', opacity);
			});
	}

	applyDragHandler(nodes) {
		if (!nodes) {
			return;
		}

		const self = this;
		nodes.call(drag()
			.on('start', e => {
				dragStarted(e, self);
			})
			.on('drag', function (e, d) {
				dragged(e, self, this, d);
			})
			.on('end', e => {
				dragEnded(e, self);
			})
			.filter(e => !e.ctrlKey && !e.shiftKey && !e.button)
			.clickDistance(3),
		);
	}

	// Draws the state of the graph, returns all newly added nodes
	draw() {
		this.resetClick();
		this.resetLasso();

		// If there is a simulation ongoing from a previous call, stop it
		if (this.simulation) {
			this.simulation.stop();
		}

		const nodes = [...this.getFilteredMap().values()];
		const links = this.getLinks(nodes);

		const svgRect = this.#rootSvg.node().getBoundingClientRect();
		if (this.virtualWidth && this.virutalHeight) {
			svgRect.width = this.virtualWidth;
			svgRect.height = this.virutalHeight;
		}

		this.simulation = forceSimulation(nodes)
			.force('link', forceLink(links).id(d => d.uid))
			.force('collide', forceCollide(d => {
				if (d.type === WORKSPACE_NODE_TYPE_NOTE) {
					if (d.width && d.height) {
						// With max() the distance to other nodes in non-quadratic rects is too high, therefore use min()
						return Math.min(d.width, d.height);
					}

					return 50;
				}

				return this.#nodeRadius * 4;
			}))
			.force('limit', forceLimit().x0(0).x1(svgRect.width).y0(0).y1(svgRect.height).radius(this.#nodeRadius)).stop();

		// Do simulation
		this.simulation.tick(Math.ceil(Math.log(this.simulation.alphaMin()) / Math.log(1 - this.simulation.alphaDecay())));

		const link = this.#lineGroup
			.selectAll('.arrow')
			.data(links, d => `${d.source}${d.target}`)
			.join('line')
			.classed('arrow', true)
			.attr('marker-start', d => d.isDual ? `url(#${this.#svgID}_arrowhead_reversed)` : undefined)
			.attr('x1', d => d.isDual ? reduceXR(d, this.#nodeRadius) : d.source.x)
			.attr('y1', d => d.isDual ? reduceYR(d, this.#nodeRadius) : d.source.y)
			.attr('x2', d => reduceX(d, this.#nodeRadius))
			.attr('y2', d => reduceY(d, this.#nodeRadius));

		// For mouseover events
		const shadowLinks = this.#shadowLineGroup
			.selectAll('.shadowArrow')
			.data(links, d => `${d.source}${d.target}`)
			.join('line')
			.classed('shadowArrow', true)
			.attr('marker-start', d => d.isDual ? `url(#${this.#svgID}_arrowhead_reversed_shadow)` : undefined)
			.attr('x1', d => d.isDual ? reduceXR(d, this.#nodeRadius) : d.source.x)
			.attr('y1', d => d.isDual ? reduceYR(d, this.#nodeRadius) : d.source.y)
			.attr('x2', d => reduceX(d, this.#nodeRadius))
			.attr('y2', d => reduceY(d, this.#nodeRadius));

		const self = this;
		shadowLinks
			.on('click', function (e, d) {
				self.lineClick(e, d, this);
			})
			.on('mouseenter', function () {
				d3Select(this).classed('lineHovered', true);
			})
			.on('mouseleave', function () {
				d3Select(this).classed('lineHovered', false);
			});

		const node = this.#nodeGroup
			.selectAll('.nodeContainer')
			.data(nodes, d => d.uid)
			.join(enter => {
				const g = enter.append('g');
				this.drawEntities(g);
				this.#newNodes = g;
				return g;
			},
			update => {
				if (this.#changedData.size > 0) {
					// Do drawing only for actually updated nodes
					this.drawEntities(update.filter(d => this.#changedData.has(d.uid)));
				}

				return update;
			})
			.classed('nodeContainer', true)
			.each(d => {
				// Exclude every node from force simulation
				d.fx = d.x;
				d.fy = d.y;
			})
			.attr('transform', d => `translate(${d.x},${d.y})`);

		this.applyDragHandler(node);

		this.simulation.on('tick', () => {
			link
				.attr('x1', d => d.isDual ? reduceXR(d, this.#nodeRadius) : d.source.x)
				.attr('y1', d => d.isDual ? reduceYR(d, this.#nodeRadius) : d.source.y)
				.attr('x2', d => reduceX(d, this.#nodeRadius))
				.attr('y2', d => reduceY(d, this.#nodeRadius));
			shadowLinks
				.attr('x1', d => d.isDual ? reduceXR(d, this.#nodeRadius) : d.source.x)
				.attr('y1', d => d.isDual ? reduceYR(d, this.#nodeRadius) : d.source.y)
				.attr('x2', d => reduceX(d, this.#nodeRadius))
				.attr('y2', d => reduceY(d, this.#nodeRadius));

			node.attr('transform', d => `translate(${d.x},${d.y})`);
		});

		this.#lasso.items(node.selectAll('.node,.note'));

		this.#changedData.clear();
	}

	// CenterGraph centers the graph in the center of the svg
	centerGraph() {
		const svgBoundingRect = this.#rootSvg.node().getBoundingClientRect();
		const rgBoundingBox = this.#rootGroup.node().getBBox();
		const rgBoundingRect = this.#rootGroup.node().getBoundingClientRect();

		// Calculate scaling, reduce the svg size so the root group is scaled slightly smaller than the svg size
		const scaleHeight = (svgBoundingRect.height - 120) / rgBoundingRect.height;
		const scaleWidth = (svgBoundingRect.width - 100) / rgBoundingRect.width;

		this.#rootSvg.call(this.#zoom.translateTo, rgBoundingBox.x + rgBoundingBox.width / 2, rgBoundingBox.y + rgBoundingBox.height / 2);

		const scaleBy = Math.min(scaleHeight, scaleWidth);

		// Return if scaling is negligible
		if (Math.abs(1 - scaleBy) < 0.1) {
			return;
		}

		this.#rootSvg.call(this.#zoom.scaleBy, scaleBy);
	}

	// Returns all nodes with their attached attributes from the force simulation performed in draw()
	exportNodes() {
		const nodes = structuredClone([...this.#nodeMap.values()]);
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

	isEmpty() {
		return this.#nodeMap.size === 0;
	}

	setNodeClickCallback(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.#nodeClickCallBack = callback;
		return true;
	}

	setLineClickCallback(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.#lineClickCallBack = callback;
		return true;
	}

	// SetZoomCallback receives a function as an argument.
	// The function is going to be called each time the root SVG zoomed upon
	setSvgZoomCallback(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.#svgZoomCallback = callback;
		return true;
	}

	// SetSvgClickCallback receives a function as an argument.
	// The function is going to be called each time the root SVG is clicked
	setSvgClickCallback(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.#svgClickCallback = callback;
		return true;
	}

	// SetContextMenuCallback receives a function as an argument.
	// The function is going to be called each time the context menu is activated.
	setContextMenuCallback(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.#contextMenuCallback = callback;
		return true;
	}

	// SetLassoSelectionCallback receives a function as an argument.
	// The function is going to be called each time nodes are selected via the lasso
	setLassoSelectionCallback(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.#lassoSelectionCallback = callback;
		return true;
	}

	// Returns the node which triggered the context menu event or click event
	getContextNode() {
		return this.#contextNodeData;
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

	// SetLassoResetCallback receives a function as an argument.
	// The function is going to be called when the lasso is reset.
	setLassoResetCallback(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.#lassoResetCallback = callback;
		return true;
	}

	populateHeuristicMap(heuristicDescriptions) {
		// Titles to map
		heuristicDescriptions.forEach(e => this.#heuristicTypeMap.set(e.type, e.title));
	}
}
