/* eslint-disable no-constant-condition */
import * as d3 from "d3";


const rectWidth = 150;

// phantom node id
const rootIdentifier = 'root';

// click switch
let isClicked = false;

// root svg elements
let rootSvg, rootGroup;

// tree layout function
let treeLayout;

// dragging
let dragActive = false, dragNode, dragLayoutData = null, setPointer = false,
    dragLayoutHiddenNodes = null;

// mouseOver
let activeMouseOverNode = null, lastMouseOverNode;


// context menu
let activeContextMenuNode = null;

// heuristic type map
let heuristicTypeMap = new Map();

// zoom
let zoom = null;

function dragStart(_, d) {
    dragNode = d3.select(this);
    dragActive = true;
    dragLayoutData = d;
}

function dragEvent(event) {
    // originally done in dragStart, but that caused the click event to not propagate
    if (!setPointer) {
        setPointer = true;
        dragNode.attr("pointer-events", "none");

        // filter out the dragged node, so it does not get removed
        const linksToHide = dragLayoutData.descendants();
        dragLayoutHiddenNodes = linksToHide.filter(d => d.data.data.uid !== dragLayoutData.data.data.uid);

        // hide the nodes
        rootSvg.selectAll(".node")
            .data(dragLayoutHiddenNodes, d => d.data.data.uid)
            .attr("pointer-events", "none")
            .attr("opacity", 0);
        rootSvg.selectAll(".link")
            .data(linksToHide, d => d.data.data.uid)
            .attr("stroke-opacity", 0);

        // color all nodes which are valid targets
        rootSvg.selectAll(".rect")
            .classed("valid-target", d => {
                return isValidMoveTarget(dragNode, d);
            });
    }

    const transformationMatrix = this.transform.baseVal.getItem(0).matrix;

    d3.select(this)
        .raise()  // causes bug in chrome: click is only recognized on second time. moved here from dragStart
        .attr("transform",
            "translate(" + (transformationMatrix.e + event.dx) + "," + (transformationMatrix.f + event.dy) + ")");

    // Move hidden nodes to the same position as the parent node,
    // so when they get displayed the have a nice transition animation
    if (dragLayoutHiddenNodes !== null) {
        rootSvg.selectAll(".node")
            .data(dragLayoutHiddenNodes, d => d.data.data.uid)
            .attr("transform",
                "translate(" + (transformationMatrix.e + event.dx) + "," + (transformationMatrix.f + event.dy) + ")");
    }
}

// checkType checks the passed heuristic can be a child based on the type
function checkType(node) {
    return node.type !== 'one_source';
}

// isValidMoveTarget returns true if the potential new parent is a valid target
function isValidMoveTarget(node, newParentNode) {
    const nodeData = node.data();
    if (nodeData.length === 0)
        return false;

    if (!checkType(nodeData[0].data.data)) {
        return false
    }

    const thisUid = nodeData[0].data.data.uid;

    if (nodeData[0].data.data.parent_heuristic === undefined) {
        // check if newParentNode is not a selection
        if (newParentNode.data !== undefined && typeof newParentNode.data !== 'function') {
            return thisUid !== newParentNode.data.data.uid;
        }

        return thisUid !== newParentNode.data()[0].data.data.uid;
    }


    const thisParentUid = nodeData[0].data.data.parent_heuristic[0].uid;

    // check if newParentNode is not a selection
    if (newParentNode.data !== undefined && typeof newParentNode.data !== 'function') {
        return thisParentUid !== newParentNode.data.data.uid && thisUid !== newParentNode.data.data.uid;
    }

    return thisParentUid !== newParentNode.data()[0].data.data.uid && thisUid !== newParentNode.data()[0].data.data.uid;
}

// dragEnd gets called when the drag event ends. If applicable it moves a dragged subtree to its new parent
function dragEnd(event, context) {
    dragNode = dragNode.attr("pointer-events", null);
    rootGroup.selectAll(".selected").classed("selected", false);
    rootSvg.selectAll(".rect").classed("valid-target", false);

    // reset pointer events
    if (dragLayoutHiddenNodes !== null) {
        rootSvg.selectAll(".node")
            .data(dragLayoutHiddenNodes, d => d.data.data.uid)
            .attr("pointer-events", null);
    }

    // only move node if drag was active before --> not clicked and activeMouseOverNode is set
    if (activeMouseOverNode !== null && setPointer && activeMouseOverNode.attr("opacity") > 0
        && isValidMoveTarget(dragNode, activeMouseOverNode)) {
        moveNode(context, activeMouseOverNode, dragNode);
    }

    // house keeping
    activeMouseOverNode = null;
    lastMouseOverNode = null;
    dragLayoutData = null;
    setPointer = false;
    dragActive = false;
    dragLayoutHiddenNodes = null;

    dragNode = null;
    context.updateGraph();
}

function setNodesChanged(changedNodes) {
    // reset
    rootSvg.selectAll("rect")
        .classed("modified", false);
    // set changes
    rootSvg.selectAll("rect")
        .data(changedNodes, d => d.data.data.uid)
        .classed("modified", true);
}

// moveNode sets parent as the parent node of the child subgraph
function moveNode(context, parent, child) {
    if (context.data === null)
        return;
    let parentData = parent.data()[0].data.data, childData = child.data()[0].data.data,
        formerParentUid = null;

    if (childData.parent_heuristic !== undefined) {
        formerParentUid = childData.parent_heuristic[0].uid;
    }

    // // set class for changed nodes
    // // todo remove
    // rootSvg.selectAll("rect").data(child.data()[0].descendants(), d => d.data.data.uid).classed("modified", true);

    let newData = context.data;

    for (let i = 0; i < newData.length; i++) {
        let dataElement = newData[i];
        if (dataElement.uid === parentData.uid) {
            if (dataElement.children === undefined) {
                dataElement.children = [];
            }

            let alreadyExists = false;
            dataElement.children.forEach(c => {
                if (c.uid === childData.uid)
                    alreadyExists = true;
            });

            if (!alreadyExists) {
                dataElement.children.push({'uid': childData.uid});
            }
        } else if (dataElement.uid === childData.uid) {
            if (dataElement.parent_heuristic === undefined) {
                dataElement.parent_heuristic = [];
            }
            dataElement.parent_heuristic = [];
            dataElement.parent_heuristic.push({'uid': parentData.uid});
        } else if (dataElement.uid === formerParentUid) {
            dataElement.children = dataElement.children.filter(c => c.uid !== childData.uid);
        }
    }

    // set new state
    context.data = newData;

    context.updateChangeSet()
}

// getDescendants returns all descendants of the node with the given uid
function getDescendants(uid) {
    // find the node
    let changedNode = null;
    for (const d of rootSvg.selectAll("g").data()) {
        if (d === undefined) {
            continue;
        }
        if (uid === d.data.data.uid) {
            changedNode = d;
            break;
        }
    }

    if (changedNode === null) {
        return [];
    }

    return changedNode.descendants();
}

function drawRect(rootElement) {
    const textAreaHeight = 50, textPadding = 0, rectHeight = textAreaHeight + 2 * textPadding,
        borderRadius = 5, strokeWidth = 2, textHeight = 10;

    rootElement
        .append("rect")
        .attr("x", -rectWidth / 2)
        .attr("y", -rectHeight / 2)
        .attr("class", "rect")
        .attr("width", (d) => {
            if (d.data.data.uid === rootIdentifier)
                return 0;
            return rectWidth;
        })
        .attr("height", rectHeight)
        .attr("rx", borderRadius)
        .attr("ry", borderRadius)
        .attr("stroke-width", strokeWidth)
        .attr("stroke-opacity", d => {
            // only draw rect if it is not the root node
            if (d.data.data.uid !== rootIdentifier)
                return 1;
            return 1;
        });


    rootElement.append("text")
        .attr("x", function () {
            return -rectWidth / 2 + strokeWidth + 2;
        })
        .attr("y", function (d) {
            // is parameter is not set, position text at center
            return d.data.data.parameter !== undefined ? -textAreaHeight / 2 + textHeight * 2 : textHeight / 2;
        })
        .text(function (d) {
            let outText;
            // only draw text if it is not the root node
            if (d.data.data.uid === rootIdentifier) {
                outText = null;
            } else {
                const title = heuristicTypeMap.get(d.data.data.type);
                if (title !== undefined)
                    outText = `Type: ${title}`;
            }

            return outText;
        });

    rootElement.append("text")
        .attr("x", function () {
            return -rectWidth / 2 + strokeWidth + 2;
        })
        .attr("y", textAreaHeight / 2 - textHeight)
        .text(function (d) {
            return d.data.data.parameter !== undefined ? `Parameter: ${d.data.data.parameter}` : null;
        });
}

function mouseOverNode(_, d) {
    if (dragActive && d !== dragLayoutData) {
        if (d !== lastMouseOverNode) {
            let tmpMouseOverNode = d3.select(this);
            if (isValidMoveTarget(dragNode, tmpMouseOverNode)) {
                lastMouseOverNode = d;
                activeMouseOverNode = tmpMouseOverNode;
                activeMouseOverNode.select(".rect").classed("selected", true);
            }
        }
    }
}

function mouseOutNode() {
    if (dragActive) {
        d3.select(this).select(".rect").classed("selected", false);
    }
    lastMouseOverNode = null;
    activeMouseOverNode = null;
}

function contextMenuHandler(context, event, d) {
    context.showContextMenu(event);
    activeContextMenuNode = d;
}

function drawNodes(group, nodeData, context) {
    // adds each node as a group
    const t = d3.transition()
        .duration(300)
        .ease(d3.easeLinear);

    return group.selectAll(".node")
        .data(nodeData.descendants(), d => d.data.data.uid)
        .join(enter => {
                const g = enter.append("g")
                    .on('mouseover', mouseOverNode)
                    .on('mouseout', mouseOutNode)
                    // set click handler
                    .on('click', nodeClicked)
                    // set context menu handler
                    .on("contextmenu", (e, d) => contextMenuHandler(context, e, d))
                    // set drag handler
                    .call(d3.drag()
                        .on("start", dragStart)
                        .on("drag", dragEvent)
                        .on("end", (e) => dragEnd(e, context)));
                // draw outline and text
                drawRect(g);
                return g;
            },
        )
        .attr("opacity", 1)
        .attr("class", function (d) {
            if (d.data.data.uid === rootIdentifier)
                return null;

            return "node" +
                (d.children ? " node--internal" : " node--leaf");
        }).transition(t)
        .attr("transform", function (d) {
            return "translate(" + d.y + "," + d.x + ")";
        });
}

function drawLinks(group, nodeData) {
    // adds the links between the nodes
    const t = d3.transition()
        .duration(300)
        .ease(d3.easeLinear);
    group.selectAll(".link")
        .data(nodeData.descendants().slice(1), d => d.data.data.uid)
        .join("path")
        .attr("class", "link")
        .transition(t)
        .attr("stroke-opacity", d => {
            // only draw link if parent is not the root node
            if (d.parent.data.data.uid !== rootIdentifier)
                return 1;
            return 0;
        })
        .attr("d", function (d) {
            return "M" + (d.y - rectWidth / 2) + "," + d.x
                + "C" + (d.y + d.parent.y) / 2 + "," + d.x
                + " " + (d.y + d.parent.y) / 2 + "," + d.parent.x
                + " " + (d.parent.y + rectWidth / 2) + "," + d.parent.x;
        });
}

function processGraphData(graphData) {
    const stratifyData = d3.stratify()
        .id(function (d) {
            return d.uid;
        })
        .parentId(function (d) {
            if (d.uid === rootIdentifier) {
                return null;
            } else if (d.parent_heuristic == null) {
                return rootIdentifier;
            }
            return d.parent_heuristic[0].uid;
        });

    const treeData = stratifyData(graphData);

    //  assigns the data to a hierarchy using parent-child relationships
    let nodes = d3.hierarchy(treeData, function (d) {
        return d.children;
    });
    let levelWidth = [1], levelDepth = 0;
    const childCount = function (level, n) {
        if (levelDepth < level) {
            levelDepth = level;
        }
        if (n.children && n.children.length > 0) {
            if (levelWidth.length <= level + 1)
                levelWidth.push(0);


            levelWidth[level + 1] += n.children.length;
            n.children.forEach(function (d) {
                childCount(level + 1, d);
            });
        }
    };

    childCount(0, treeData);

    // declares a tree layout and assigns the size
    treeLayout = d3.tree().size([d3.max(levelWidth) * 150, levelDepth * 200]);
    // maps the node data to the tree layout
    nodes = treeLayout(nodes);
    return nodes;
}

// setupSvg sets up the root svg, adds the zoom and drag handler and sets the heuristic titles
function setupSvg(context, canvasId, heuristicDescriptions) {
    // titles to map
    heuristicDescriptions.forEach(e => heuristicTypeMap.set(e.id, e.title));

    // add attributes to root svg
    rootSvg = d3.select("#" + canvasId)
        .attr("class", "graph-canvas")
        .on("click", resetClick);
    rootGroup = rootSvg
        .append("g")
        .attr("class", "root-group");

    // add zoom and drag
    zoom = d3.zoom()
        .on("zoom", (event) => {
            context.displayContextMenu = false;
            rootGroup.attr('transform', event.transform);
        })
        .scaleExtent([0.5, 8]);
    rootSvg.call(zoom);
}

function resetClick() {
    // only do work if needed
    if (!isClicked)
        return;

    // reset click representation
    d3.selectAll(".rect").classed("clicked", false);
    // set not clicked
    isClicked = false;
}

function nodeClicked(e) {
    const thisElement = d3.select(this);
    if (thisElement.data()[0].data.data.uid === rootIdentifier)
        return;

    resetClick();
    e.stopPropagation();
    // set click representation
    thisElement.select(".rect").classed("clicked", true);
    // set clicked
    isClicked = true;
}

function addRootElement(data) {
    data.push({'uid': 'root'});
}

function drawGraph(data, context) {
    drawLinks(rootGroup, data);
    drawNodes(rootGroup, data, context);
}

// adds a heuristic with the given id and parameter
function addHeuristic(id, parameter) {
    if (parameter)
        console.log("adding heuristic " + id + " with paramter " + parameter);
    else
        console.log("adding heuristic " + id);
}

// getRemovableNodes returns elements which can be removed based on the position saved in activeContextMenuNode
function getRemovableNodes() {
    const nodesToRemove = activeContextMenuNode.descendants();

    let toBeRemoved = [];

    nodesToRemove.forEach(e => {
        toBeRemoved.push(e.data.data.uid);
    });

    return toBeRemoved;
}

// getRemovableRelationship returns the uid and parent uid of the node to be removed
function getRemovableRelationship() {
    let ret = {};
    ret.childUid = activeContextMenuNode.data.data.uid;
    if (activeContextMenuNode.data.data.parent_heuristic)
        ret.parentUid = activeContextMenuNode.data.data.parent_heuristic[0].uid;
    else
        ret.parentUid = '';


    return ret;
}

// centerGraph centers the graph in the center of the svg
async function centerGraph() {
    const svgRect = rootSvg.node().getBoundingClientRect();
    let bbRect = null;
    while (true) {
        bbRect = rootGroup.node().getBoundingClientRect();

        // wait until group has a size
        if (bbRect.height !== 0) {
            break
        } else {
            await new Promise(r => setTimeout(r, 200));
        }
    }

    const transform = d3.zoomIdentity
        .translate(bbRect.width * 2, 200)
        .scale((svgRect.height - 200) / bbRect.height);

    rootSvg.transition().duration(750).call(zoom.transform, transform);
}

export {
    drawGraph, processGraphData, addHeuristic, setupSvg,
    addRootElement, getRemovableNodes, centerGraph, getRemovableRelationship,
    getDescendants, setNodesChanged, rootIdentifier
};