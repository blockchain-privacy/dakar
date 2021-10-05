import * as d3 from 'd3';

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// isFunction returns true if the provided argument is a function
// credits: https://stackoverflow.com/questions/5999998/check-if-a-variable-is-of-function-type
export function isFunction(functionToCheck) {
  return functionToCheck && {}.toString.call(functionToCheck) === '[object Function]';
}

export class Tree {
  constructor(width) {
    this.rectWidth = width;

    // click switch
    this.isClicked = false;
    this.clickCallBack = null;
    this.drawClicked = null;
    this.drawResetClick = null;

    // root svg elements
    this.rootSvg = null;
    this.rootGroup = null;

    // zoom
    this.zoom = null;

    // must be implemented in child classes
    this.stratify = null;

    // draw tree function
    this.drawTree = null;
  }

  // centerGraph centers the graph in the center of the svg
  async centerGraph() {
    const svgRect = this.rootSvg.node().getBoundingClientRect();
    let bbRect = null;

    // wait a bit so svg elements can be properly added to the DOM
    await sleep(100);

    while (bbRect === null || bbRect.height === 0) {
      bbRect = this.rootGroup.node().getBoundingClientRect();

      // wait until group has a size
      if (bbRect.height !== 0) {
        break;
      } else {
        // we have to use await here
        // eslint-disable-next-line no-await-in-loop
        await sleep(200);
      }
    }

    const scaleHeight = svgRect.height / (bbRect.height * 1.2);
    const scaleWidth = svgRect.width / (bbRect.width * 1.2);

    // scale between 0.1 and 5
    const scaleBy = Math.max(Math.min(scaleHeight, scaleWidth, 5), 0.1);
    const transform = d3.zoomIdentity.translate(-10, 0).scale(scaleBy);
    this.rootSvg.transition().duration(250).call(this.zoom.transform, transform);
  }

  // getDescendants returns all descendants of the node with the given uid
  getDescendants(uid) {
    // find the node
    let changedNode = null;

    // use some() instead of a proper for loop to comply with eslint
    this.rootSvg.selectAll('g').data().some((d) => {
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

    //  assigns the data to a hierarchy using parent-child relationships
    const nodes = d3.hierarchy(treeData, (d) => d.children);
    const levelWidth = [1];
    let levelDepth = 0;
    const childCount = (level, n) => {
      if (levelDepth < level) {
        levelDepth = level;
      }
      if (n.children && n.children.length > 0) {
        if (levelWidth.length <= level + 1) levelWidth.push(0);

        levelWidth[level + 1] += n.children.length;
        n.children.forEach((d) => {
          childCount(level + 1, d);
        });
      }
    };

    childCount(0, treeData);

    // declares a tree layout and assigns the size
    const treeLayout = d3.tree().size([d3.max(levelWidth) * 150, levelDepth * 200]);
    // maps the node data to the tree layout

    this.drawTree(treeLayout(nodes));
  }

  resetClick() {
    if (this.drawResetClick === null) throw new Error('drawResetClick is not implemented');
    // only do work if needed
    if (!this.isClicked) return;
    // reset click representation
    this.drawResetClick();
    // set not clicked
    this.isClicked = false;
  }

  // setupSvg sets up the root svg, adds the zoom and drag handler and sets the heuristic titles
  setupSvg(context, canvasId) {
    // add attributes to root svg
    this.rootSvg = d3.select(`#${canvasId}`).on('click', () => this.resetClick());
    this.rootGroup = this.rootSvg.append('g').attr('class', 'root-group');

    // add zoom and drag
    this.zoom = d3.zoom()
      .on('zoom', (event) => {
        context.displayContextMenu = false;
        this.rootGroup.attr('transform', event.transform);
      })
      .scaleExtent([0.5, 8]);
    this.rootSvg.call(this.zoom);
  }

  // drawClickedState sets the correct class so it looks clicked
  drawClickedState(node) {
    if (this.drawClicked === null) {
      throw new Error('drawClicked is not implemented');
    }

    this.resetClick();
    // set click representation
    this.drawClicked(node);
    // set clicked
    this.isClicked = true;

    // execute clickHandler
    this.clickCallBack(node.data()[0].data.data);
    return true;
  }

  // setNodeClickHandler receives a function as an argument.
  // The function is going to be called each time a node is clicked.
  setNodeClickHandler(callback) {
    if (!isFunction(callback)) {
      return false;
    }

    this.clickCallBack = callback;
    return true;
  }

  static nodeClicked(e, classContext, d3This) {
    const thisElement = d3.select(d3This);
    if (!classContext.drawClickedState(thisElement)) return;

    e.stopPropagation();
  }
}
