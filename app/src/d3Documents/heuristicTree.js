
import * as d3 from 'd3';
import {
	mdiMerge, mdiPlaylistRemove, mdiTune, mdiClockAlertOutline,
} from '@mdi/js';
import Tree from './tree';
import {abbreviateNumber} from './util';
import {isFunction} from '@/utilities';

// Phantom node id
export const rootIdentifier = 'root';

export class HeuristicTree extends Tree {
	constructor(width, context) {
		super(width);

		// Dragging
		this.dragActive = false;
		this.dragNode = null;
		this.dragLayoutData = null;
		this.setPointer = false;
		this.dragLayoutHiddenNodes = null;

		// MouseOver
		this.activeMouseOverNode = null;
		this.lastMouseOverNode = null;

		// Context menu
		this.activeContextMenuNode = null;
		this.activeContextMenuSelection = null;

		// Heuristic type map
		this.heuristicTypeMap = new Map();

		// Callbacks
		this.contextMenuCallback = null;

		this.drawClicked = node => {
			node.selectAll('.rect').classed('clicked', true);
		};

		this.drawResetClick = () => {
			this.rootSvg.selectAll('.rect').classed('clicked', false);
		};

		this.stratify = d3.stratify()
			.id(d => d.uid)
			.parentId(d => {
				if (d.uid === rootIdentifier) {
					return null;
				}

				if (!d.parent) {
					return rootIdentifier;
				}

				return d.parent[0].uid;
			});

		this.drawTree = data => {
			this.drawLinks(this.rootGroup, data);
			this.drawNodes(this.rootGroup, data, context);
		};
	}

	static dragStart(d, classContext, d3This) {
		classContext.dragNode = d3.select(d3This);
		classContext.dragActive = true;
		classContext.dragLayoutData = d;
	}

	static dragEvent(event, classContext, d3This) {
		// Originally done in dragStart, but that caused the click event to not propagate
		if (!classContext.setPointer) {
			classContext.setPointer = true;
			classContext.dragNode.attr('pointer-events', 'none');

			// Filter out the dragged node, so it does not get removed
			const linksToHide = classContext.dragLayoutData.descendants();
			classContext.dragLayoutHiddenNodes = linksToHide.filter(
				d => d.data.data.uid !== classContext.dragLayoutData.data.data.uid,
			);

			// Hide the nodes
			classContext.rootSvg.selectAll('.node')
				.data(classContext.dragLayoutHiddenNodes, d => d.data.data.uid)
				.attr('pointer-events', 'none')
				.attr('opacity', 0);
			classContext.rootSvg.selectAll('.link')
				.data(linksToHide, d => d.data.data.uid)
				.attr('stroke-opacity', 0);

			// Color all nodes which are valid targets
			classContext.rootSvg.selectAll('.rect')
				.classed('valid-target', d => HeuristicTree.isValidMoveTarget(classContext.dragNode, d));
		}

		const transformationMatrix = d3This.transform.baseVal.getItem(0).matrix;

		d3.select(d3This)
		// Raise() causes bug in chrome: click is only
		// recognized on second time. moved here from dragStart
			.raise()
			.attr(
				'transform',
				`translate(${transformationMatrix.e + event.dx},${transformationMatrix.f + event.dy})`,
			);

		// Move hidden nodes to the same position as the parent node,
		// so when they get displayed they have a nice transition animation
		if (classContext.dragLayoutHiddenNodes !== null) {
			classContext.rootSvg.selectAll('.node')
				.data(classContext.dragLayoutHiddenNodes, d => d.data.data.uid)
				.attr(
					'transform',
					`translate(${transformationMatrix.e + event.dx},${transformationMatrix.f + event.dy})`,
				);
		}
	}

	// MoveNode sets parent as the parent node of the child subgraph
	static moveNode(context, parent, child) {
		if (!context.data || !context.data.heuristics) {
			return;
		}

		const parentData = parent.data()[0].data.data;
		const childData = child.data()[0].data.data;
		let formerParentUid = null;

		if (childData.parent !== undefined) {
			formerParentUid = childData.parent[0].uid;
		}

		const newData = context.data.heuristics;

		for (let i = 0; i < newData.length; i += 1) {
			const dataElement = newData[i];
			if (dataElement.uid === parentData.uid) {
				if (dataElement.children === undefined) {
					dataElement.children = [];
				}

				let alreadyExists = false;
				dataElement.children.forEach(c => {
					if (c.uid === childData.uid) {
						alreadyExists = true;
					}
				});

				if (!alreadyExists) {
					dataElement.children.push({uid: childData.uid});
				}
			} else if (dataElement.uid === childData.uid) {
				if (dataElement.parent === undefined) {
					dataElement.parent = [];
				}

				dataElement.parent = [];
				dataElement.parent.push({uid: parentData.uid});
			} else if (dataElement.uid === formerParentUid) {
				dataElement.children = dataElement.children.filter(c => c.uid !== childData.uid);
			}
		}

		// Set new state
		context.data.heuristics = newData;
	}

	// DragEnd gets called when the drag event ends.
	// If applicable it moves a dragged subtree to its new parent
	static dragEnd(context, classContext) {
		classContext.dragNode = classContext.dragNode.attr('pointer-events', null);
		classContext.rootGroup.selectAll('.selected').classed('selected', false);
		classContext.rootSvg.selectAll('.rect').classed('valid-target', false);

		// Reset pointer events
		if (classContext.dragLayoutHiddenNodes !== null) {
			classContext.rootSvg.selectAll('.node')
				.data(classContext.dragLayoutHiddenNodes, d => d.data.data.uid)
				.attr('pointer-events', null);
		}

		// Only move node if drag was active before --> not clicked and activeMouseOverNode is set
		if (classContext.activeMouseOverNode !== null && classContext.setPointer
            && classContext.activeMouseOverNode.attr('opacity') > 0
            && HeuristicTree.isValidMoveTarget(classContext.dragNode, classContext.activeMouseOverNode,
            )) {
			HeuristicTree.moveNode(context, classContext.activeMouseOverNode, classContext.dragNode);
		}

		// Housekeeping
		classContext.activeMouseOverNode = null;
		classContext.lastMouseOverNode = null;
		classContext.dragLayoutData = null;
		classContext.setPointer = false;
		classContext.dragActive = false;
		classContext.dragLayoutHiddenNodes = null;

		classContext.dragNode = null;
		context.updateGraph();
	}

	setNodesChanged(changedNodes) {
		// Reset
		this.rootSvg.selectAll('rect').classed('modified', false);
		// Set changes
		this.rootSvg.selectAll('rect').filter(d => changedNodes.has(d.data.data.uid)).classed('modified', true);
	}

	// CheckType checks the passed heuristic can be a child based on the type
	static checkType(node) {
		return node.type !== 'one_source';
	}

	// IsValidMoveTarget returns true if the potential new parent is a valid target
	static isValidMoveTarget(node, newParentNode) {
		const nodeData = node.data();
		if (nodeData.length === 0) {
			return false;
		}

		if (!HeuristicTree.checkType(nodeData[0].data.data)) {
			return false;
		}

		const thisUid = nodeData[0].data.data.uid;

		if (nodeData[0].data.data.parent === undefined) {
			// Check if newParentNode is not a selection

			if (newParentNode.data !== undefined && !isFunction(newParentNode.data)) {
				return thisUid !== newParentNode.data.data.uid;
			}

			return thisUid !== newParentNode.data()[0].data.data.uid;
		}

		const thisParentUid = nodeData[0].data.data.parent[0].uid;

		// Check if newParentNode is not a selection
		if (newParentNode.data !== undefined && !isFunction(newParentNode.data)) {
			return thisParentUid !== newParentNode.data.data.uid
                && thisUid !== newParentNode.data.data.uid;
		}

		return thisParentUid !== newParentNode.data()[0].data.data.uid
            && thisUid !== newParentNode.data()[0].data.data.uid;
	}

	drawRect(rootElement) {
		const textAreaHeight = 50;
		const textPadding = 0;
		const rectHeight = textAreaHeight + 2 * textPadding;
		const borderRadius = 5;
		const strokeWidth = 2;
		const textHeight = 10;
		const iconScale = 'scale(0.7,0.7)';
		const iconWidth = 16;
		const iconY = rectHeight / 2 - 20;

		rootElement
			.append('rect')
			.attr('x', -this.rectWidth / 2)
			.attr('y', -rectHeight / 2)
			.attr('class', 'rect')
			.attr('width', d => {
				if (d.data.data.uid === rootIdentifier) {
					return 0;
				}

				return this.rectWidth;
			})
			.attr('height', rectHeight)
			.attr('rx', borderRadius)
			.attr('ry', borderRadius)
			.attr('stroke-width', strokeWidth)
			.attr('stroke-opacity', 1);

		// Exclusion list icon
		rootElement
			.append('g')
			.attr('transform', `translate(${this.rectWidth / 2 - iconWidth - 4},${iconY}) ${iconScale}`)
			.append('path')
			.attr('fill', 'currentColor')
			.attr('d', d => {
				if (d.data.data.uid === rootIdentifier) {
					return '';
				}

				if (d.data.data.excludeAddresses) {
					return mdiPlaylistRemove;
				}

				return '';
			});

		// Cluster icon
		rootElement
			.append('g')
			.attr('transform', `translate(${this.rectWidth / 2 - 2 * iconWidth - 2 * 4},${iconY}) ${iconScale}`)
			.append('path')
			.attr('fill', 'currentColor')
			.attr('d', d => {
				if (d.data.data.uid === rootIdentifier) {
					return '';
				}

				if (d.data.data.clusterTypes
            && d.data.data.clusterTypes.length > 0) {
					return mdiMerge;
				}

				return '';
			});

		// Spending gap icon
		rootElement
			.append('g')
			.attr('transform', `translate(${this.rectWidth / 2 - 3 * iconWidth - 3 * 4},${iconY}) ${iconScale}`)
			.append('path')
			.attr('fill', 'currentColor')
			.attr('d', d => {
				if (d.data.data.uid === rootIdentifier) {
					return '';
				}

				if (d.data.data.excludeSpendingGaps) {
					return mdiClockAlertOutline;
				}

				return '';
			});

		// Parameter icon
		rootElement
			.append('g')
			.attr('transform', `translate(${-this.rectWidth / 2 + 4},${iconY}) ${iconScale}`)
			.append('path')
			.attr('fill', 'currentColor')
			.attr('d', d => {
				if (d.data.data.uid === rootIdentifier) {
					return '';
				}

				if (d.data.data.parameter !== undefined) {
					return mdiTune;
				}

				return '';
			});

		rootElement
			.append('text')
			.style('text-anchor', 'middle')
			.attr('fill', 'currentColor')
			.attr('y', -textAreaHeight / 2 + textHeight * 2)
			.text(d => {
				let outText;
				// Only draw text if it is not the root node
				if (d.data.data.uid === rootIdentifier) {
					outText = null;
				} else {
					const title = this.heuristicTypeMap.get(d.data.data.type);
					if (title !== undefined) {
						outText = title;
					}
				}

				return outText;
			});

		rootElement.append('text')
			.attr('fill', 'currentColor')
			.attr('x', -this.rectWidth / 2 + 25)
			.attr('y', 18)
			.text(d => {
				if (d.data.data.uid === rootIdentifier) {
					return null;
				}

				return d.data.data.parameter === undefined ? '' : d.data.data.parameter;
			});

		const resultHeight = 18;
		const resultWidth = 38;
		const offset = 1.3;

		// Result box rect
		rootElement
			.append('rect')
			.attr('x', this.rectWidth / 2 - resultWidth / offset)
			.attr('y', -rectHeight / 2 - resultHeight / 2)
			.attr('class', 'rect')
			.attr('width', d => {
				if (d.data.data.uid === rootIdentifier
            || d.data.data.clusterCount === undefined) {
					return 0;
				}

				return resultWidth;
			})
			.attr('height', resultHeight)
			.attr('rx', borderRadius)
			.attr('ry', borderRadius)
			.attr('stroke-width', strokeWidth)
			.attr('stroke-opacity', 1);

		// Result box text
		rootElement.append('text')
			.attr('fill', 'currentColor')
			.attr(
				'transform',
				`translate(${(this.rectWidth / 2 - resultWidth / offset + resultWidth / 2)} ,
        ${-rectHeight / 2 - resultHeight / 2 + textHeight + 3})`,
			)
			.style('text-anchor', 'middle')
			.text(d => {
				if (d.data.data.uid === rootIdentifier
              || d.data.data.clusterCount === undefined) {
					return null;
				}

				return `${abbreviateNumber(d.data.data.clusterCount)}`;
			});
	}

	static setNodeSelected(classContext, dataNode, selectionNode) {
		if (HeuristicTree.isValidMoveTarget(classContext.dragNode, selectionNode)) {
			classContext.lastMouseOverNode = dataNode;
			classContext.activeMouseOverNode = selectionNode;
			classContext.activeMouseOverNode.select('.rect').classed('selected', true);
		}
	}

	static setNodeNotSelected(classContext) {
		if (classContext.dragActive && classContext.activeMouseOverNode !== null) {
			classContext.activeMouseOverNode.select('.rect').classed('selected', false);
		}

		classContext.lastMouseOverNode = null;
		classContext.activeMouseOverNode = null;
	}

	static mouseOverNode(d, classContext, d3This) {
		if (classContext.dragActive && d !== classContext.dragLayoutData) {
			if (d !== classContext.lastMouseOverNode) {
				HeuristicTree.setNodeSelected(classContext, d, d3.select(d3This));
			}
		}
	}

	// TouchMove the selection state of a heuristic rect.
	// It is a combination of the mouseover and mouseout events but for touch controls.
	static touchMove(classContext, event) {
		if (!classContext.dragActive) {
			return;
		}

		const startUID = classContext.dragLayoutData.data.data.uid;
		if (startUID === '') {
			return;
		}

		const touchEvent = event.targetTouches[0];
		const elements = document.elementsFromPoint(touchEvent.pageX, touchEvent.pageY);
		if (elements === null) {
			return;
		}

		let heuristicWasHovered = false;
		elements.forEach(element => {
			const elementData = element.__data__;

			if (elementData !== undefined
            && element.tagName === 'rect'
            && elementData.data.data.uid !== startUID) {
				// Detect if a heuristic is currently hovered
				heuristicWasHovered = true;
			}

			if (elementData !== undefined
            && element.tagName === 'rect'
            && classContext.lastMouseOverNode !== elementData
            && elementData.data.data.uid !== startUID) {
				HeuristicTree.setNodeSelected(classContext, elementData, d3.select(element.parentElement));
			}
		});

		// Nothing hovered -> remove 'selected' class from last rect
		if (!heuristicWasHovered) {
			HeuristicTree.setNodeNotSelected(classContext);
		}
	}

	// ContextMenuHandler is called when a context menu event for a node occurs
	static contextMenuHandler(event, d, classContext, d3This) {
		classContext.contextMenuCallback(event);
		classContext.activeContextMenuNode = d;
		classContext.activeContextMenuSelection = d3.select(d3This);
	}

	populateHeuristicMap(heuristicDescriptions) {
		// Titles to map
		heuristicDescriptions.forEach(e => this.heuristicTypeMap.set(e.type, e.title));
	}

	// DrawClickedState sets the correct class so it looks clicked
	drawClickedState(node) {
		if (node.data()[0].data.data.uid === rootIdentifier) {
			return false;
		}

		return super.drawClickedState(node);
	}

	// SimulateClick simulates a click and executes the click handler
	simulateClick() {
		this.drawClickedState(this.activeContextMenuSelection);
	}

	drawNodes(group, nodeData, context) {
		const self = this;

		// Adds each node as a group
		return group.selectAll('.node')
			.data(nodeData.descendants(), d => d.data.data.uid)
			.join(enter => {
				const g = enter.append('g')
				// eslint-disable-next-line func-names
					.on('mouseover', function mouseOver(e, d) {
						HeuristicTree.mouseOverNode(d, self, this);
					})
					.on('touchmove', e => {
						HeuristicTree.touchMove(self, e);
					})
					.on('mouseout', () => {
						HeuristicTree.setNodeNotSelected(self);
					})
				// Set click handler
				// eslint-disable-next-line func-names
					.on('click', function click(e) {
						Tree.nodeClicked(e, self, this);
					})
				// Set context menu handler
				// eslint-disable-next-line func-names
					.on('contextmenu', function contextMenu(e, d) {
						HeuristicTree.contextMenuHandler(e, d, self, this);
					})
				// Set drag handler
					.call(d3.drag()
					// Functions can not be defined as arrow functions,
					// because then 'this' will not be set to the d3Context
					// eslint-disable-next-line func-names
						.on('start', function dragStart(e, d) {
							HeuristicTree.dragStart(d, self, this);
						})
					// eslint-disable-next-line func-names
						.on('drag', function drag(e) {
							HeuristicTree.dragEvent(e, self, this);
						})
						.on('end', () => {
							HeuristicTree.dragEnd(context, self);
						}));
				// Draw outline and text
				self.drawRect(g);
				return g;
			})
			.attr('opacity', 1)
			.attr('class', d => {
				if (d.data.data.uid === rootIdentifier) {
					return 'node';
				}

				return `node${
					d.children ? ' node--internal' : ' node--leaf'}`;
			})
			.transition(d3.transition().duration(300).ease(d3.easeLinear))
			.attr('transform', d => `translate(${d.y},${d.x})`);
	}

	drawLinks(group, nodeData) {
		// Adds the links between the nodes
		group.selectAll('.link')
			.data(nodeData.descendants().slice(1), d => d.data.data.uid)
			.join('path')
			.attr('class', 'link')
			.transition(d3.transition().duration(300).ease(d3.easeLinear))
			.attr('stroke-opacity', d => {
				// Only draw link if parent is not the root node
				if (d.parent.data.data.uid !== rootIdentifier) {
					return 1;
				}

				return 0;
			})
			.attr('d', d => `M${d.y - this.rectWidth / 2},${d.x
			}C${(d.y + d.parent.y) / 2},${d.x
			} ${(d.y + d.parent.y) / 2},${d.parent.x
			} ${d.parent.y + this.rectWidth / 2},${d.parent.x}`);
	}

	// GetRemovableNodes returns elements which can be removed based on
	// the position saved in activeContextMenuNode
	getRemovableNodes() {
		const nodesToRemove = this.activeContextMenuNode.descendants();

		const toBeRemoved = [];

		nodesToRemove.forEach(e => {
			toBeRemoved.push(e.data.data.uid);
		});

		return toBeRemoved;
	}

	// GetRemovableRelationship returns the uid and parent uid of the node to be removed
	getRemovableRelationship() {
		const ret = {};
		ret.childUid = this.activeContextMenuNode.data.data.uid;
		if (this.activeContextMenuNode.data.data.parent) {
			ret.parentUid = this.activeContextMenuNode.data.data.parent[0].uid;
		} else {
			ret.parentUid = '';
		}

		return ret;
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
