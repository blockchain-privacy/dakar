/* eslint-disable no-constant-condition */
import * as d3 from 'd3';
import { mdiMerge, mdiPlaylistRemove, mdiTune } from '@mdi/js';
import Tree from './tree';
import { isFunction, abbreviateNumber } from './util';

// phantom node id
export const rootIdentifier = 'root';

export class HeuristicTree extends Tree {
  constructor(width, context) {
    super(width);

    // dragging
    this.dragActive = false;
    this.dragNode = null;
    this.dragLayoutData = null;
    this.setPointer = false;
    this.dragLayoutHiddenNodes = null;

    // mouseOver
    this.activeMouseOverNode = null;
    this.lastMouseOverNode = null;

    // context menu
    this.activeContextMenuNode = null;
    this.activeContextMenuSelection = null;

    // heuristic type map
    this.heuristicTypeMap = new Map();

    // callbacks
    this.contextMenuCallback = null;

    this.drawClicked = (node) => {
      node.selectAll('.rect').classed('clicked', true);
    };

    this.drawResetClick = () => {
      this.rootSvg.selectAll('.rect').classed('clicked', false);
    };

    this.stratify = d3.stratify()
      .id((d) => d.uid)
      .parentId((d) => {
        if (d.uid === rootIdentifier) {
          return null;
        }
        if (d.parent == null) {
          return rootIdentifier;
        }
        return d.parent[0].uid;
      });

    this.drawTree = (data) => {
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
    // originally done in dragStart, but that caused the click event to not propagate
    if (!classContext.setPointer) {
      classContext.setPointer = true;
      classContext.dragNode.attr('pointer-events', 'none');

      // filter out the dragged node, so it does not get removed
      const linksToHide = classContext.dragLayoutData.descendants();
      classContext.dragLayoutHiddenNodes = linksToHide.filter(
        (d) => d.data.data.uid !== classContext.dragLayoutData.data.data.uid,
      );

      // hide the nodes
      classContext.rootSvg.selectAll('.node')
        .data(classContext.dragLayoutHiddenNodes, (d) => d.data.data.uid)
        .attr('pointer-events', 'none')
        .attr('opacity', 0);
      classContext.rootSvg.selectAll('.link')
        .data(linksToHide, (d) => d.data.data.uid)
        .attr('stroke-opacity', 0);

      // color all nodes which are valid targets
      classContext.rootSvg.selectAll('.rect')
        .classed('valid-target', (d) => HeuristicTree.isValidMoveTarget(classContext.dragNode, d));
    }

    const transformationMatrix = d3This.transform.baseVal.getItem(0).matrix;

    d3.select(d3This)
    // raise() causes bug in chrome: click is only
    // recognized on second time. moved here from dragStart
      .raise()
      .attr('transform',
        `translate(${transformationMatrix.e + event.dx},${transformationMatrix.f + event.dy})`);

    // Move hidden nodes to the same position as the parent node,
    // so when they get displayed they have a nice transition animation
    if (classContext.dragLayoutHiddenNodes !== null) {
      classContext.rootSvg.selectAll('.node')
        .data(classContext.dragLayoutHiddenNodes, (d) => d.data.data.uid)
        .attr('transform',
          `translate(${transformationMatrix.e + event.dx},${transformationMatrix.f + event.dy})`);
    }
  }

  // moveNode sets parent as the parent node of the child subgraph
  static moveNode(context, parent, child) {
    if (!context.data || !context.data.heuristics) return;
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
        dataElement.children.forEach((c) => {
          if (c.uid === childData.uid) alreadyExists = true;
        });

        if (!alreadyExists) {
          dataElement.children.push({ uid: childData.uid });
        }
      } else if (dataElement.uid === childData.uid) {
        if (dataElement.parent === undefined) {
          dataElement.parent = [];
        }
        dataElement.parent = [];
        dataElement.parent.push({ uid: parentData.uid });
      } else if (dataElement.uid === formerParentUid) {
        dataElement.children = dataElement.children.filter((c) => c.uid !== childData.uid);
      }
    }

    // set new state
    context.data.heuristics = newData;
  }

  // dragEnd gets called when the drag event ends.
  // If applicable it moves a dragged subtree to its new parent
  static dragEnd(context, classContext) {
    classContext.dragNode = classContext.dragNode.attr('pointer-events', null);
    classContext.rootGroup.selectAll('.selected').classed('selected', false);
    classContext.rootSvg.selectAll('.rect').classed('valid-target', false);

    // reset pointer events
    if (classContext.dragLayoutHiddenNodes !== null) {
      classContext.rootSvg.selectAll('.node')
        .data(classContext.dragLayoutHiddenNodes, (d) => d.data.data.uid)
        .attr('pointer-events', null);
    }

    // only move node if drag was active before --> not clicked and activeMouseOverNode is set
    if (classContext.activeMouseOverNode !== null && classContext.setPointer
            && classContext.activeMouseOverNode.attr('opacity') > 0
            && HeuristicTree.isValidMoveTarget(classContext.dragNode,
              classContext.activeMouseOverNode)) {
      HeuristicTree.moveNode(context, classContext.activeMouseOverNode, classContext.dragNode);
    }

    // housekeeping
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
    // reset
    this.rootSvg.selectAll('rect').classed('modified', false);
    // set changes
    this.rootSvg.selectAll('rect').filter((d) => changedNodes.has(d.data.data.uid)).classed('modified', true);
  }

  // checkType checks the passed heuristic can be a child based on the type
  static checkType(node) {
    return node.type !== 'one_source';
  }

  // isValidMoveTarget returns true if the potential new parent is a valid target
  static isValidMoveTarget(node, newParentNode) {
    const nodeData = node.data();
    if (nodeData.length === 0) return false;

    if (!HeuristicTree.checkType(nodeData[0].data.data)) {
      return false;
    }

    const thisUid = nodeData[0].data.data.uid;

    if (nodeData[0].data.data.parent === undefined) {
      // check if newParentNode is not a selection
      if (newParentNode.data !== undefined && typeof newParentNode.data !== 'function') {
        return thisUid !== newParentNode.data.data.uid;
      }

      return thisUid !== newParentNode.data()[0].data.data.uid;
    }

    const thisParentUid = nodeData[0].data.data.parent[0].uid;

    // check if newParentNode is not a selection
    if (newParentNode.data !== undefined && typeof newParentNode.data !== 'function') {
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
      .attr('width', (d) => {
        if (d.data.data.uid === rootIdentifier) return 0;
        return this.rectWidth;
      })
      .attr('height', rectHeight)
      .attr('rx', borderRadius)
      .attr('ry', borderRadius)
      .attr('stroke-width', strokeWidth)
      .attr('stroke-opacity', 1);

    // exclusion list icon
    rootElement
      .append('g')
      .attr('transform', `translate(${this.rectWidth / 2 - iconWidth - 4},${iconY}) ${iconScale}`)
      .append('path')
      .attr('fill', 'currentColor')
      .attr('d', (d) => {
        if (d.data.data.uid === rootIdentifier) {
          return '';
        }
        if (d.data.data.excludeAddresses) return mdiPlaylistRemove;
        return '';
      });

    // cluster icon
    rootElement
      .append('g')
      .attr('transform', `translate(${this.rectWidth / 2 - 2 * iconWidth - 2 * 4},${iconY}) ${iconScale}`)
      .append('path')
      .attr('fill', 'currentColor')
      .attr('d', (d) => {
        if (d.data.data.uid === rootIdentifier) {
          return '';
        }
        if (d.data.data.clusterTypes
            && d.data.data.clusterTypes.length > 0) return mdiMerge;
        return '';
      });

    // parameter icon
    rootElement
      .append('g')
      .attr('transform', `translate(${-this.rectWidth / 2 + 4},${iconY}) ${iconScale}`)
      .append('path')
      .attr('fill', 'currentColor')
      .attr('d', (d) => {
        if (d.data.data.uid === rootIdentifier) {
          return '';
        }
        if (d.data.data.parameter !== undefined) return mdiTune;
        return '';
      });

    rootElement
      .append('text')
      .style('text-anchor', 'middle')
      .attr('fill', 'currentColor')
      .attr('y', -textAreaHeight / 2 + textHeight * 2)
      .text((d) => {
        let outText;
        // only draw text if it is not the root node
        if (d.data.data.uid === rootIdentifier) {
          outText = null;
        } else {
          const title = this.heuristicTypeMap.get(d.data.data.type);
          if (title !== undefined) outText = title;
        }

        return outText;
      });

    rootElement.append('text')
      .attr('fill', 'currentColor')
      .attr('x', -this.rectWidth / 2 + 25)
      .attr('y', 18)
      .text((d) => {
        if (d.data.data.uid === rootIdentifier) return null;
        return d.data.data.parameter !== undefined ? d.data.data.parameter : '';
      });

    const resultHeight = 18; const
      resultWidth = 38; const
      offset = 1.3;

    // result box rect
    rootElement
      .append('rect')
      .attr('x', this.rectWidth / 2 - resultWidth / offset)
      .attr('y', -rectHeight / 2 - resultHeight / 2)
      .attr('class', 'rect')
      .attr('width', (d) => {
        if (d.data.data.uid === rootIdentifier
            || d.data.data.clusterCount === undefined) return 0;
        return resultWidth;
      })
      .attr('height', resultHeight)
      .attr('rx', borderRadius)
      .attr('ry', borderRadius)
      .attr('stroke-width', strokeWidth)
      .attr('stroke-opacity', 1);

    // result box text
    rootElement.append('text')
      .attr('fill', 'currentColor')
      .attr('transform',
        `translate(${(this.rectWidth / 2 - resultWidth / offset + resultWidth / 2)} ,
        ${-rectHeight / 2 - resultHeight / 2 + textHeight + 3})`)
      .style('text-anchor', 'middle')
      .text((d) => {
        if (d.data.data.uid === rootIdentifier
              || d.data.data.clusterCount === undefined) return null;
        return `${abbreviateNumber(d.data.data.clusterCount)}`;
      });
  }

  static mouseOverNode(d, classContext, d3This) {
    if (classContext.dragActive && d !== classContext.dragLayoutData) {
      if (d !== classContext.lastMouseOverNode) {
        const tmpMouseOverNode = d3.select(d3This);
        if (HeuristicTree.isValidMoveTarget(classContext.dragNode, tmpMouseOverNode)) {
          classContext.lastMouseOverNode = d;
          classContext.activeMouseOverNode = tmpMouseOverNode;
          classContext.activeMouseOverNode.select('.rect').classed('selected', true);
        }
      }
    }
  }

  static mouseOutNode(classContext, d3This) {
    if (classContext.dragActive) {
      d3.select(d3This).select('.rect').classed('selected', false);
    }
    classContext.lastMouseOverNode = null;
    classContext.activeMouseOverNode = null;
  }

  // contextMenuHandler is called when a context menu event for a node occurs
  static contextMenuHandler(event, d, classContext, d3This) {
    classContext.contextMenuCallback(event);
    classContext.activeContextMenuNode = d;
    classContext.activeContextMenuSelection = d3.select(d3This);
  }

  populateHeuristicMap(heuristicDescriptions) {
    // titles to map
    heuristicDescriptions.forEach((e) => this.heuristicTypeMap.set(e.type, e.title));
  }

  // drawClickedState sets the correct class so it looks clicked
  drawClickedState(node) {
    if (node.data()[0].data.data.uid === rootIdentifier) return false;
    return super.drawClickedState(node);
  }

  // simulateClick simulates a click and executes the click handler
  simulateClick() {
    this.drawClickedState(this.activeContextMenuSelection);
  }

  drawNodes(group, nodeData, context) {
    const self = this;

    // adds each node as a group
    const t = d3.transition().duration(300).ease(d3.easeLinear);

    return group.selectAll('.node')
      .data(nodeData.descendants(), (d) => d.data.data.uid)
      .join((enter) => {
        const g = enter.append('g')
          .on('mouseover', function mouseOver(e, d) { HeuristicTree.mouseOverNode(d, self, this); })
          .on('mouseout', function mouseOut() { HeuristicTree.mouseOutNode(self, this); })
        // set click handler
          .on('click', function click(e) { Tree.nodeClicked(e, self, this); })
        // set context menu handler
          .on('contextmenu', function contextMenu(e, d) { HeuristicTree.contextMenuHandler(e, d, self, this); })
        // set drag handler
          .call(d3.drag()
          // functions can not be defined as arrow functions,
          // because then 'this' will not be set to the d3Context
            .on('start', function dragStart(e, d) { HeuristicTree.dragStart(d, self, this); })
            .on('drag', function drag(e) { HeuristicTree.dragEvent(e, self, this); })
            .on('end', () => { HeuristicTree.dragEnd(context, self); }));
        // draw outline and text
        self.drawRect(g);
        return g;
      })
      .attr('opacity', 1)
      .attr('class', (d) => {
        if (d.data.data.uid === rootIdentifier) return 'node';

        return `node${
          d.children ? ' node--internal' : ' node--leaf'}`;
      })
      .transition(t)
      .attr('transform', (d) => `translate(${d.y},${d.x})`);
  }

  drawLinks(group, nodeData) {
    // adds the links between the nodes
    const t = d3.transition()
      .duration(300)
      .ease(d3.easeLinear);
    group.selectAll('.link')
      .data(nodeData.descendants().slice(1), (d) => d.data.data.uid)
      .join('path')
      .attr('class', 'link')
      .transition(t)
      .attr('stroke-opacity', (d) => {
        // only draw link if parent is not the root node
        if (d.parent.data.data.uid !== rootIdentifier) return 1;
        return 0;
      })
      .attr('d', (d) => `M${d.y - this.rectWidth / 2},${d.x
      }C${(d.y + d.parent.y) / 2},${d.x
      } ${(d.y + d.parent.y) / 2},${d.parent.x
      } ${d.parent.y + this.rectWidth / 2},${d.parent.x}`);
  }

  // getRemovableNodes returns elements which can be removed based on
  // the position saved in activeContextMenuNode
  getRemovableNodes() {
    const nodesToRemove = this.activeContextMenuNode.descendants();

    const toBeRemoved = [];

    nodesToRemove.forEach((e) => {
      toBeRemoved.push(e.data.data.uid);
    });

    return toBeRemoved;
  }

  // getRemovableRelationship returns the uid and parent uid of the node to be removed
  getRemovableRelationship() {
    const ret = {};
    ret.childUid = this.activeContextMenuNode.data.data.uid;
    if (this.activeContextMenuNode.data.data.parent) {
      ret.parentUid = this.activeContextMenuNode.data.data.parent[0].uid;
    } else ret.parentUid = '';

    return ret;
  }

  // setContextMenuCallback receives a function as an argument.
  // The function is going to be called each time the context menu is activated.
  setContextMenuCallback(callback) {
    if (!isFunction(callback)) {
      return false;
    }

    this.contextMenuCallback = callback;
    return true;
  }
}
