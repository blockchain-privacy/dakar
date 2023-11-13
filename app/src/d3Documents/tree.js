import * as d3 from 'd3';
import {sleep} from './util';
import {isFunction} from '@/utilities';
export default class Tree {
	constructor(width) {
		this.rectWidth = width;

		// Click switch
		this.isClicked = false;
		this.clickCallBack = null;
		this.drawClicked = null;
		this.drawResetClick = null;
		this.svgClickCallback = null;

		// Root svg elements
		this.rootSvg = null;
		this.rootGroup = null;

		// Zoom
		this.zoom = null;

		// Must be implemented in child classes
		this.stratify = null;

		// Draw tree function
		this.drawTree = null;
	}

	// CenterGraph centers the graph in the center of the svg
	async centerGraph() {
		const svgRect = this.rootSvg.node().getBoundingClientRect();
		let bbRect = null;

		// Wait a bit so svg elements can be properly added to the DOM
		await sleep(100);

		while (bbRect === null || bbRect.height === 0) {
			bbRect = this.rootGroup.node().getBoundingClientRect();

			// Wait until group has a size
			if (bbRect.height === 0) {
				// We have to use await here
				// eslint-disable-next-line no-await-in-loop
				await sleep(200);
			} else {
				break;
			}
		}

		const scaleHeight = svgRect.height / (bbRect.height * 1.2);
		const scaleWidth = svgRect.width / (bbRect.width * 1.2);

		// Scale between 0.1 and 5
		const scaleBy = Math.max(Math.min(scaleHeight, scaleWidth, 5), 0.1);
		const transform = d3.zoomIdentity.translate(-10, 0).scale(scaleBy);
		this.rootSvg.transition().duration(250).call(this.zoom.transform, transform);
	}

	// GetDescendants returns all descendants of the node with the given uid
	getDescendants(uid) {
		// Find the node
		let changedNode = null;

		// Use some() instead of a proper for loop to comply with eslint
		this.rootSvg.selectAll('g').data().some(d => {
			if (d !== undefined && uid === d.data.data.uid) {
				changedNode = d;
				return true;
			}

			return false;
		});

		if (changedNode === null) {
			return [];
		}

		return changedNode.descendants();
	}

	processGraphData(graphData) {
		if (this.stratify === null) {
			throw new Error('stratify is not implemented');
		}

		if (this.drawTree === null) {
			throw new Error('drawTree is not implemented');
		}

		const treeData = this.stratify(graphData);

		//  Assigns the data to a hierarchy using parent-child relationships
		const nodes = d3.hierarchy(treeData, d => d.children);
		const levelWidth = [1];
		let levelDepth = 0;
		const childCount = (level, n) => {
			if (levelDepth < level) {
				levelDepth = level;
			}

			if (n.children && n.children.length > 0) {
				if (levelWidth.length <= level + 1) {
					levelWidth.push(0);
				}

				levelWidth[level + 1] += n.children.length;
				n.children.forEach(d => {
					childCount(level + 1, d);
				});
			}
		};

		childCount(0, treeData);

		// Declares a tree layout and assigns the size
		const treeLayout = d3.tree().size([d3.max(levelWidth) * 150, levelDepth * 200]);
		// Maps the node data to the tree layout

		this.drawTree(treeLayout(nodes));
	}

	svgClick() {
		if (this.drawResetClick === null) {
			throw new Error('drawResetClick is not implemented');
		}

		if (this.svgClickCallback !== null) {
			this.svgClickCallback();
		}

		// Only do work if needed
		if (!this.isClicked) {
			return;
		}

		// Reset click representation
		this.drawResetClick();
		// Set not clicked
		this.isClicked = false;
	}

	// SetupSvg sets up the root svg, adds the zoom and drag handler and sets the heuristic titles
	setupSvg(context, canvasId) {
		// Add attributes to root svg
		this.rootSvg = d3.select(`#${canvasId}`).on('click', () => this.svgClick());
		this.rootGroup = this.rootSvg.append('g').attr('class', 'root-group');

		// Add zoom and drag
		this.zoom = d3.zoom()
			.on('zoom', event => {
				context.displayContextMenu = false;
				this.rootGroup.attr('transform', event.transform);
			})
			.scaleExtent([0.5, 8]);
		this.rootSvg.call(this.zoom);
	}

	// DrawClickedState sets the correct class so it looks clicked
	drawClickedState(node) {
		if (this.drawClicked === null) {
			throw new Error('drawClicked is not implemented');
		}

		this.svgClick();
		// Set click representation
		this.drawClicked(node);
		// Set clicked
		this.isClicked = true;

		// Execute clickHandler
		this.clickCallBack(node.data()[0].data.data);
		return true;
	}

	// SetNodeClickHandler receives a function as an argument.
	// The function is going to be called each time a node is clicked.
	setNodeClickHandler(callback) {
		if (!isFunction(callback)) {
			return false;
		}

		this.clickCallBack = callback;
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

	static nodeClicked(e, classContext, d3This) {
		const thisElement = d3.select(d3This);
		if (!classContext.drawClickedState(thisElement)) {
			return;
		}

		e.stopPropagation();
	}
}
