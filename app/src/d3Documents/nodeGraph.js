// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {drag} from 'd3-drag';
import {select as d3Select} from 'd3-selection';
import {zoom, zoomTransform} from 'd3-zoom';
import {forceCollide, forceLink, forceSimulation} from 'd3-force';
import {mdiExclamationThick} from '@mdi/js';
import d3lasso from './d3Lasso.js';
import {abbreviateNumber, reduceX, reduceY} from '@/d3Documents/util';
import forceLimit from '@/d3Documents/forceLimit';
import {
	WORKSPACE_NODE_TYPE_SELECTOR,
	WORKSPACE_NODE_TYPE_NOTE,
	WORKSPACE_NODE_TYPE_TRANSACTION,
	SELECTOR_STATUS_WAITING,
	SELECTOR_STATUS_ERROR,
} from '@/constants/index.js';
import {isFunction} from '@/utilities';

// In ms
const animationDuration = 175;
const longAnimationDuration = 500;
const animationDelay = 2000;
const markerColor = 'rgba(255, 109, 0, 0.3)';

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
			// Don't change the actual dragged node
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
	// Node type
	#nodeTypeColorMap = null;
	// If enabled, node descriptions are not rendered and event handlers are disabled
	#enableThumbnailMode = false;
	enableInteractions = true;

	constructor(nodeTypeColorMap) {
		this.#nodeTypeColorMap = nodeTypeColorMap;
	}

	setEnableInteractions(flag) {
		this.enableInteractions = flag;
	}

	setEnableThumbnailMode(flag) {
		this.#enableThumbnailMode = flag;
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

		const allowed = this.#filterNodeTypes.length > 0 ? this.#filterNodeTypes.includes(node.type) : true;

		if (!allowed) {
			return false;
		}

		if (node.type === WORKSPACE_NODE_TYPE_TRANSACTION && node.txtype) {
			return this.#filterPrivacyTypes.includes(node.txtype);
		}

		return true;
	}

	filterNodes(nodeTypes, privacyTypes) {
		this.#filterNodeTypes = nodeTypes ?? [];

		this.#filterPrivacyTypes = privacyTypes ?? [];

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
		this.#nodeGroup.selectAll('.nodeClicked').classed('nodeClicked', false);
		this.#shadowLineGroup.selectAll('.arrowClicked').classed('arrowClicked', false);
	}

	resetLasso() {
		this.#nodeGroup.selectAll('.lasso-selected').classed('lasso-selected', false);
		this.lassoSelectedNodes = null;

		if (this.#lassoResetCallback !== null) {
			this.#lassoResetCallback();
		}
	}

	setContextObjectClicked() {
		if (!this.#contextNodeSelection) {
			return;
		}

		this.resetClick();
		this.resetLasso();

		const contextNode = d3Select(this.#contextNodeSelection);

		// Try selecting the active object, can be a node or a line
		if (contextNode.classed('shadowArrow')) {
			contextNode.classed('arrowClicked', true);
		} else {
			contextNode.select('.shadowNode').classed('nodeClicked', true);
		}
	}

	selectAllNodes() {
		const m = this.getFilteredMap();

		const filteredNodeKeys = new Set(m.keys());
		const visibleNodes = this.#nodeGroup.selectAll('.node,.note').filter(d => filteredNodeKeys.has(d.uid));

		visibleNodes.classed('lasso-selected', true);
		this.lassoSelectedNodes = this.#nodeGroup.selectAll('.lasso-selected');
		if (this.#lassoSelectionCallback !== null) {
			this.#lassoSelectionCallback();
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
		this.#rootSvg = d3Select(`#${svgID}`);
		if (!this.#enableThumbnailMode) {
			this.#rootSvg.on('click', () => this.svgClick());
		}

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
			.filter(e => ((!e.ctrlKey && !this.getLassoEnabled()) || e instanceof WheelEvent) && !e.button && !this.#enableThumbnailMode)
			.scaleExtent([0.5, 3]);
		this.#rootSvg.call(this.#zoom);

		// Add lasso
		const self = this;
		if (!this.#enableThumbnailMode) {
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
		}

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

        .node {
          stroke: currentColor;
          stroke-width: 1;
          cursor: pointer;
        }

        .shadowNode {
          fill: rgb(var(--v-theme-primary));
          opacity: 0;
          transition: all 0.175s ease;
        }

        .nodeClicked {
          opacity: 0.3;
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
          stroke-width: 4;
          opacity: 0;
          marker-end: url(#${this.#svgID}_arrowhead_shadow);
          transition: all 0.175s ease;
        }

        .arrowHovered, .arrowClicked {
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

		return [...links.values()];
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

	// ReorderNodes randomizes the position of all visible nodes
	// and runs the force simulation to determine new node coordinates.
	reorderNodes() {
		if (this.lassoSelectedNodes === null) {
			for (const [key, value] of this.getFilteredMap()) {
				// Randomize position, reorderNodes creates different arrangements for each call
				value.x = Math.random();
				value.y = Math.random();
				delete value.fx;
				delete value.fy;
				this.#nodeMap[key] = value;
				if (this.#filteredNodeMap[key] !== undefined) {
					this.#filteredNodeMap[key] = value;
				}
			}

			this.draw();
			this.centerGraph();
		} else {
			this.lassoSelectedNodes.each(d => {
				delete d.fx;
				delete d.fy;
				this.#nodeMap[d.uid] = d;
				if (this.#filteredNodeMap[d.uid] !== undefined) {
					this.#filteredNodeMap[d.uid] = d;
				}
			});
			this.draw(false, false);
			this.centerOnSelection(this.lassoSelectedNodes);
		}
	}

	// Returns the coordinates of the current center of the visible SVG area
	getCenterOfView() {
		const svgNode = this.#rootSvg.node();
		const bbox = svgNode.getBoundingClientRect();
		const transform = zoomTransform(svgNode);

		return {
			x: ((bbox.width / 2) - transform.x) / transform.k,
			y: ((bbox.height / 2) - transform.y) / transform.k,
		};
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
		let enableSimulation = false;
		if (mapNode !== undefined) {
			node.x = mapNode.x;
			node.y = mapNode.y;
		} else if (node.x === undefined) {
			const centerPosition = this.getCenterOfView();
			if (centerPosition.x !== undefined) {
				node.x = centerPosition.x;
				node.y = centerPosition.y;
				enableSimulation = true;
			}
		}

		// Enable simulation for nodes added to the center of the graph
		const n = enableSimulation ? node : setFxFy(node);

		this.#nodeMap.set(n.uid, n);
		this.#changedData.set(n.uid, n);

		if (this.isAllowedByFilter(n)) {
			this.#filteredNodeMap.set(n.uid, n);
		}

		if (draw === undefined || draw === true) {
			this.draw(true, false);
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
			this.draw(true, false);
		}
	}

	// Returns the node specified node. If the node does not
	// exist in the graph undefined is returned.
	getNode(uid) {
		return this.#nodeMap.get(uid);
	}

	getNodes() {
		return [...this.#nodeMap.values()];
	}

	centerOnNewNodes() {
		this.centerOnSelection(this.#newNodes);
	}

	centerOnNode(n) {
		if (!n.uid) {
			return;
		}

		const selection = this.#nodeGroup
			.selectAll('.nodeContainer')
			.data([...this.getFilteredMap().values()], d => d.uid)
			.filter(d => d.uid === n.uid);
		if (selection.empty()) {
			return;
		}

		this.centerOnSelection(selection);
	}

	centerOnSelection(selection) {
		if (selection === null || selection.empty()) {
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
		const centerX = minX + (width / 2);
		const centerY = minY + (height / 2);

		this.#rootSvg.transition().duration(250).call(this.#zoom.translateTo, centerX, centerY);
	}

	drawIcons(groupElement, isTitleSet, icons, parameter) {
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
		let iconY = this.#nodeRadius + textHeight + (textAreaMargin * 2);

		if (!isTitleSet) {
			iconY = this.#nodeRadius + 3 + textAreaMargin;
		}

		// Remove all children
		iconGroup.selectAll('*').remove();

		icons.forEach(async (icon, i) => {
			iconGroup.append('path')
				.attr('transform', `translate(${(iconWidth * i) + (iconMargin * i)},${iconY}) scale(0.45,0.45)`)
				.attr('fill', 'currentColor')
				.attr('d', icon);
		});

		if (parameter) {
			iconGroup.append('text')
				.attr('transform', `translate(${(iconWidth * icons.length) + (iconMargin * icons.length)},${iconY + 9})`)
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

		this.drawNodes(
			groupElement.filter(d => d.type !== WORKSPACE_NODE_TYPE_NOTE),
			entityGroup.filter(d => d.type !== WORKSPACE_NODE_TYPE_NOTE),
		);
		this.drawNotes(
			groupElement.filter(d => d.type === WORKSPACE_NODE_TYPE_NOTE),
			entityGroup.filter(d => d.type === WORKSPACE_NODE_TYPE_NOTE),
		);
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
					.text(element => element ?? ' '); // Insert space for empty row so vertical spacing works

				const nodeRect = this.getBBox();
				d.bbHeight = nodeRect.height;
				d.bbWidth = nodeRect.width;
				d3Select(this).attr('y', -(nodeRect.height / 2) - 2);
			});

		const noteMarkerIncrease = 20;

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
				if (d.fx !== undefined && !d.showMarker) {
					return;
				}

				d3Select(this.parentNode).append('rect')
					.attr('width', d.width + noteMarkerIncrease)
					.attr('height', d.height + noteMarkerIncrease)
					.attr('x', -(d.width + noteMarkerIncrease) / 2)
					.attr('y', -(d.height + noteMarkerIncrease) / 2)
					.attr('rx', 3)
					.attr('ry', 3)
					.attr('fill', markerColor)
					.lower()
					.transition().delay(animationDelay).duration(longAnimationDuration)
					.attr('width', 0)
					.attr('height', 0)
					.attr('x', 0)
					.attr('y', 0)
					.remove();
			});

		if (this.#enableThumbnailMode) {
			return;
		}

		const noteHoverIncrease = 5;
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
					.attr('width', d => d.width + noteHoverIncrease)
					.attr('height', d => d.height + noteHoverIncrease)
					.attr('x', d => -(d.width + noteHoverIncrease) / 2)
					.attr('y', d => -(d.height + noteHoverIncrease) / 2);
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

					if (d.selectorType) {
						nodeColor = this.#nodeTypeColorMap.get(d.selectorType);
					} else if (d.txtype) {
						nodeColor = this.#nodeTypeColorMap.get(d.txtype);
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
				const parent = d3Select(this.parentNode);

				parent
					.append('circle')
					.classed('shadowNode', true)
					.attr('r', self.#nodeRadius * 1.5)
					.lower();

				// Add marker to new nodes
				if (d.fx !== undefined && !d.showMarker) {
					return;
				}

				parent
					.append('circle')
					.attr('r', self.#nodeRadius * 2)
					.attr('fill', markerColor)
					.lower()
					.transition().delay(animationDelay).duration(longAnimationDuration).attr('r', 0).remove();
			});

		if (this.#enableThumbnailMode) {
			return;
		}

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

		// Add node symbol
		const loadingRadius = this.#nodeRadius - 6;
		const gap = 2 * Math.PI * loadingRadius / 4;

		const gapString = `${gap} ${gap}`;

		entityGroup.each(function (d) {
			switch (d.selectorStatus) {
				case SELECTOR_STATUS_WAITING:
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
					break;
				case SELECTOR_STATUS_ERROR:
					d3Select(this)
						.append('path')
						.attr('transform', 'translate(-12,-12) scale(1,1)')
						.attr('fill', 'white')
						.attr('d', mdiExclamationThick);
					break;
				default:
			}
		});

		// Add node descriptions
		const textAreaWidth = 120;
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
			const d3Self = d3Select(this);
			let text = d3Self.text();

			// Don't elide text which is 5 characters or smaller
			if (text.length <= 5) {
				return;
			}

			const selfNode = d3Self.node();
			let textLength = selfNode.getComputedTextLength();

			// Reduce text by 10% each time and at minimum by 1
			const cutLength = Math.max(Math.floor(text.length / 10), 1);

			while (textLength > textAreaWidth && text.length > 0) {
				text = text.slice(0, -cutLength);
				// \u2026 = ...
				d3Self.text(text + '\u2026');
				textLength = selfNode.getComputedTextLength();
			}
		}

		nodeTitle
			.attr('font-size', fontSize)
			.attr('text-anchor', 'middle')
			.style('cursor', 'default')
			.attr('fill', 'currentColor')
			.attr('y', this.#nodeRadius + textHeight + textAreaMargin)
			.text(d => d.nodeDisplayTitle)
			.each(elide);

		// Symbol or text which is centered on the node
		let resultCount = entityGroup.select('.resultCount');
		if (resultCount.empty()) {
			resultCount = entityGroup.append('text').classed('resultCount', true);
		}

		resultCount.raise();

		resultCount
			.attr('text-anchor', 'middle')
			.style('cursor', 'pointer')
			.style('font-weight', 'bold')
			.attr('dominant-baseline', 'middle')
			.attr('fill', 'white')
			.attr('font-size', 12)
			.attr('y', 1)
			.text(function (d) {
				if (d.nodeDisplayResultCount.length > 3) {
					d3Select(this).attr('font-size',	9);
				}

				return d.nodeDisplayResultCount;
			});

		textContainer
			.each(function (d) {
				if (!d.nodeDisplayIconObject) {
					return;
				}

				self.drawIcons(d3Select(this), Boolean(d.nodeDisplayTitle), d.nodeDisplayIconObject.icons, d.nodeDisplayIconObject.parameter);
			});
	}

	setMouseOverAnimation(context, nodeContext, isEnter) {
		const thisNode = d3Select(nodeContext).select('.node');
		const nodeRadius = isEnter ? context.#nodeRadius * 1.2 : context.#nodeRadius;
		const opacity = isEnter ? 0.3 : 1;
		thisNode.transition().duration(animationDuration).attr('r', nodeRadius);

		const thisNodeUID = thisNode.data()[0].uid;
		context.#lineGroup.selectAll('.arrow')
			.filter(d => d.source.uid !== thisNodeUID && d.target.uid !== thisNodeUID)
			.each(function () {
				d3Select(this).transition().duration(animationDuration).attr('opacity', opacity);
			});
	}

	applyDragHandler(nodes) {
		if (!nodes || this.#enableThumbnailMode) {
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
			.clickDistance(3));
	}

	// Draws the state of the graph, returns all newly added nodes
	draw(reset = true, limitSimulation = true) {
		if (reset) {
			this.resetClick();
			this.resetLasso();
		}

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
			}));

		// The limit simulation should only be done if all nodes are reordered.
		// Otherwise, nodes can get stuck in the limit rectangle if host nodes are outside the rectangle
		if (limitSimulation) {
			this.simulation = this.simulation.force('limit', forceLimit().x0(0).x1(svgRect.width).y0(0).y1(svgRect.height).radius(this.#nodeRadius));
		}

		this.simulation.stop();

		// Do simulation
		this.simulation.tick(Math.ceil(Math.log(this.simulation.alphaMin()) / Math.log(1 - this.simulation.alphaDecay())));

		const link = this.#lineGroup
			.selectAll('.arrow')
			.data(links, d => `${d.source}${d.target}`)
			.join('line')
			.classed('arrow', true)
			.attr('marker-start', d => d.isDual ? `url(#${this.#svgID}_arrowhead_reversed)` : undefined)
			.attr('x1', d => d.isDual ? reduceX(d.target, d.source, this.#nodeRadius) : d.source.x)
			.attr('y1', d => d.isDual ? reduceY(d.target, d.source, this.#nodeRadius) : d.source.y)
			.attr('x2', d => reduceX(d.source, d.target, this.#nodeRadius))
			.attr('y2', d => reduceY(d.source, d.target, this.#nodeRadius));

		// For mouseover and click events
		const shadowLinks = this.#shadowLineGroup
			.selectAll('.shadowArrow')
			.data(links, d => `${d.source}${d.target}`)
			.join('line')
			.classed('shadowArrow', true)
			.attr('marker-start', d => d.isDual ? `url(#${this.#svgID}_arrowhead_reversed_shadow)` : undefined)
			.attr('x1', d => d.isDual ? reduceX(d.target, d.source, this.#nodeRadius) : d.source.x)
			.attr('y1', d => d.isDual ? reduceY(d.target, d.source, this.#nodeRadius) : d.source.y)
			.attr('x2', d => reduceX(d.source, d.target, this.#nodeRadius))
			.attr('y2', d => reduceY(d.source, d.target, this.#nodeRadius));

		let arrowText = null;
		if (!this.#enableThumbnailMode) {
			arrowText = this.#lineGroup.selectAll('.arrowText').data(links, d => `${d.source}${d.target}`)
				.join('text')
				.classed('arrowText', true)
				.text(d => {
					if (d.source.type === WORKSPACE_NODE_TYPE_SELECTOR && d.target.type === WORKSPACE_NODE_TYPE_SELECTOR) {
						return abbreviateNumber(d.source.selectorResultCount);
					}

					return null;
				})
				.attr('font-size', 10)
				.attr('fill', 'currentColor')
				.attr('text-anchor', 'middle')
				.attr('transform', d => `translate(${d.source.x + ((d.target.x - d.source.x) / 2)},${d.source.y + ((d.target.y - d.source.y) / 2) - 5})`);
		}

		const self = this;
		if (!this.#enableThumbnailMode) {
			shadowLinks
				.on('click', function (e, d) {
					self.lineClick(e, d, this);
				})
				.on('mouseenter', function () {
					d3Select(this).classed('arrowHovered', true);
				})
				.on('mouseleave', function () {
					d3Select(this).classed('arrowHovered', false);
				});
		}

		const node = this.#nodeGroup
			.selectAll('.nodeContainer')
			.data(nodes, d => d.uid)
			.join(
				enter => {
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
				},
			)
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
				.attr('x1', d => d.isDual ? reduceX(d.target, d.source, this.#nodeRadius) : d.source.x)
				.attr('y1', d => d.isDual ? reduceY(d.target, d.source, this.#nodeRadius) : d.source.y)
				.attr('x2', d => reduceX(d.source, d.target, this.#nodeRadius))
				.attr('y2', d => reduceY(d.source, d.target, this.#nodeRadius));
			shadowLinks
				.attr('x1', d => d.isDual ? reduceX(d.target, d.source, this.#nodeRadius) : d.source.x)
				.attr('y1', d => d.isDual ? reduceY(d.target, d.source, this.#nodeRadius) : d.source.y)
				.attr('x2', d => reduceX(d.source, d.target, this.#nodeRadius))
				.attr('y2', d => reduceY(d.source, d.target, this.#nodeRadius));

			node.attr('transform', d => `translate(${d.x},${d.y})`);
			if (arrowText) {
				arrowText.attr('transform', d => `translate(${d.source.x + ((d.target.x - d.source.x) / 2)},${d.source.y + ((d.target.y - d.source.y) / 2) - 5})`);
			}
		});

		if (!this.#enableThumbnailMode) {
			this.#lasso.items(node.selectAll('.node,.note'));
		}

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

		this.#rootSvg.call(this.#zoom.translateTo, rgBoundingBox.x + (rgBoundingBox.width / 2), rgBoundingBox.y + (rgBoundingBox.height / 2));

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
			d.x = Math.round(d.x * 10_000) / 10_000;
			d.y = Math.round(d.y * 10_000) / 10_000;
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

	resetContextNode() {
		this.#contextNodeData = null;
		this.#contextNodeSelection = null;
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
}
