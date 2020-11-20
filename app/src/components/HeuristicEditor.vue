<template>
  <v-container class="fill-height" fluid>
    <v-menu v-model="contextMenu.display"
            origin="center center"
            transition="scale-transition"
            :position-x="contextMenu.x"
            :position-y="contextMenu.y"
            absolute
            offset-y
            :close-on-click="true"
            style="max-width: 600px">
      <v-list>
        <v-list-item @click="this.deleteSubTree" link>
          <v-list-item-icon>
            <v-icon>mdi-delete</v-icon>
          </v-list-item-icon>
          <v-list-item-title>Delete sub tree</v-list-item-title>
        </v-list-item>
        <!--        <v-list-item-->
        <!--            v-for="(item, index) in items"-->
        <!--            :key="index"-->
        <!--        >-->
        <!--          <v-list-item-title @click="item.action">{{ item.title }}</v-list-item-title>-->
        <!--        </v-list-item>-->
      </v-list>
    </v-menu>
    <v-toolbar
        style="width: 100%; left:0;"
        absolute
        dense
    >
      <v-toolbar-title>
        <v-icon class="hidden-sm-and-down">mdi-transfer</v-icon>
        Transaction {{ this.shortTransactionHash }}
      </v-toolbar-title>
      <v-btn icon @click="goToTransactionPage">
        <v-icon>mdi-open-in-new</v-icon>
      </v-btn>
      <v-spacer></v-spacer>
      <v-bottom-sheet scrollable>
        <template v-slot:activator="{ on, attrs }">
          <v-btn
              outlined
              v-bind="attrs"
              v-on="on"
          >
            <v-icon>mdi-shape-square-rounded-plus</v-icon>
            <div class="hidden-sm-and-down"> Add Heuristic</div>

          </v-btn>

        </template>
        <v-card>
          <v-subheader>Add heuristic</v-subheader>
          <v-card-text style="height: 80%">
            <div class="d-flex flex-wrap">
              <v-card
                  class="mx-auto my-12"
                  v-for="(item, index) in heuristicTypes"
                  :key="index"
                  max-width="300"
              >
                <template slot="progress">
                  <v-progress-linear
                      color="deep-purple"
                      height="10"
                      indeterminate
                  ></v-progress-linear>
                </template>
                <v-img
                    height="200"
                    src="https://images.idgesg.net/images/article/2017/09/networking-100735059-large.jpg"
                ></v-img>
                <v-card-title>
                  {{ item.title }}
                </v-card-title>
                <v-card-subtitle>
                  {{ item.description }}
                </v-card-subtitle>
              </v-card>
            </div>
          </v-card-text>
        </v-card>
      </v-bottom-sheet>
    </v-toolbar>
    <!--    <v-btn @click="refreshData">Refresh</v-btn>-->
    <!--    <v-btn @click="changeData">Change</v-btn>-->
    <svg id="test_canvas" viewBox="0 0 2000 2000"></svg>
  </v-container>
</template>

<script>
import * as d3 from "d3";
import {shortenHash} from "@/utilities";
import {ROUTE_NAME_SEARCH_PAGE} from "@/constants";

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

function dragEnd(event, context) {
  dragNode = dragNode.attr("pointer-events", null);
  rootGroup.selectAll(".selected").classed("selected", false);

  if (dragLayoutHiddenNodes !== null) {
    rootSvg.selectAll(".node")
        .data(dragLayoutHiddenNodes, d => d.data.data.uid)
        .attr("pointer-events", null);
  }

  // only move node if drag was active before --> not clicked and activeMouseOverNode is set
  if (activeMouseOverNode !== null && setPointer && activeMouseOverNode.attr("opacity") > 0) {
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

function moveNode(context, parent, child) {
  if (context.data === null)
    return;
  let parentData = parent.data()[0].data.data, childData = child.data()[0].data.data,
      formerParentUid = null;

  if (childData.parent_heuristic !== undefined) {
    formerParentUid = childData.parent_heuristic[0].uid;
  }

  for (let i = 0; i < context.data.length; i++) {
    let dataElement = context.data[i];
    if (dataElement.uid === parentData.uid) {
      if (dataElement.children === undefined) {
        dataElement.children = [];
      }
      dataElement.children.push({'uid': childData.uid});
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
      lastMouseOverNode = d;
      activeMouseOverNode = d3.select(this);
      activeMouseOverNode.select(".rect").classed("selected", true);
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

function createInitialGraph(context) {
  // append the svg object to the body of the page
  // appends a 'group' element to 'svg'
  // moves the 'group' element to the top left margin
  rootSvg = d3.select("#test_canvas")
      // .attr("width", width + margin.left + margin.right)
      // .attr("height", height + margin.top + margin.bottom)
      .attr("class", "graph-canvas")
      .on("click", resetClick);
  rootGroup = rootSvg
      .append("g")
      .attr("class", "root-group");

  // .attr("transform", "translate(" + margin.left + "," + margin.top + ")")


  // add zoom and drag
  rootSvg.call(d3.zoom()
      .on("zoom", (event) => {
        context.displayContextMenu = false;
        rootGroup.attr('transform', event.transform);
      })
      .scaleExtent([0.5, 8])
  );
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

function drawGraph(g, data, context) {
  drawLinks(g, data);
  drawNodes(g, data, context);
}

function navDrag() {
  console.log("dragged");
}

export default {
  name: "HeuristicEditor",
  data: () => ({
    transactionHash: "",
    shortTransactionHash: "",
    heuristicTypes: [
      {
        id: "one_source",
        title: "One Source",
        description: "Filters by time, direct input transaction amount filter and omni sources",
        fun: navDrag,
      },
      {
        id: "global_amount",
        title: "Global Amount",
        description: "The amount heuristic filters all origins of sources, which do not have equal or " +
            "more denominations to fund the destination transaction. " +
            "Note that this is different from the direct input transaction amount filter, as " +
            "this heuristic only checks the set of origin transactions and sources per destina- " +
            "tion transaction, not per direct input transaction.",
        fun: navDrag,
      },
      {
        id: "perfect_match",
        title: "Perfect Match",
        description: "The perfect match heuristic filters all origins of sources, which have denominations " +
            "without a perfect match for the denominations of the destination transaction.",
        fun: navDrag,
      },
      {
        id: "denomination_type",
        title: "Denomination Type",
        description: "The denomination type heuristic filters all origins of sources, which have denominations " +
            "of types which do not occur in the denominations of the destination transaction." +
            "For example a destination transaction spends 5 × 10.0001 and 10 × 1.00001. " +
            "Now all sources are excluded which do not have these exact two types of denominations.",
        fun: navDrag,
      },
    ],
    contextMenu: {
      display: false,
      x: 0,
      y: 0,
      items: [
        {title: 'Dummy', action: null},
      ],
    },
  }),
  computed: {
    data() {
      return this.$store.getters.getHeuristicData;
    },
  },
  methods: {
    // called by context menu handler
    goToTransactionPage() {
      this.$router.push({name: ROUTE_NAME_SEARCH_PAGE})
    },
    deleteSubTree() {
      const nodesToRemove = activeContextMenuNode.descendants();

      let toBeRemoved = [];

      nodesToRemove.forEach(e => {
        toBeRemoved.push(e.data.data.uid);
      });
      const updatedData = this.data.filter(e =>
          !toBeRemoved.includes(e.uid)
      );

      this.$store.dispatch('setHeuristicData', updatedData);

      // update displayed graph
      this.updateGraph();
    },
    showContextMenu(e) {
      this.contextMenu.display = false;

      e.preventDefault();
      this.contextMenu.x = e.clientX;
      this.contextMenu.y = e.clientY;

      this.$nextTick(() => {
        this.contextMenu.display = true;
      })
    },
    updateGraph() {
      // maps the node data to the tree layout
      const nodeData = processGraphData(this.data);
      drawGraph(rootGroup, nodeData, this);
      // drawLinks(rootGroup, nodeData);
      // drawNodes(rootGroup, nodeData, this);
    },
    refreshData: async function () {
      await this.$store.dispatch('updateHeuristicData', this.transactionHash);
      addRootElement(this.data);
      this.updateGraph();
    },
    changeData: function () {
      this.updateGraph();
    }
  },
  mounted() {
    this.transactionHash = this.$route.params.id;
    this.shortTransactionHash = shortenHash(this.transactionHash);
    document.title = `Heuristic - ${this.transactionHash}`;
    this.heuristicTypes.forEach(e => heuristicTypeMap.set(e.id, e.title));
    createInitialGraph(this);
    this.refreshData();
  },
}
</script>

<style>
.node text {
  font: 12px sans-serif;
}

.link {
  fill: none;
  stroke: darkslategrey;
  stroke-width: 2px;
}

rect {
  stroke: #008ee5;
  fill-opacity: 0;
}

.clicked {
  stroke: #FDD835;
  fill: antiquewhite;
  fill-opacity: 1;
  stroke-dasharray: 10;
}

.graph-canvas {
  background-color: whitesmoke;
}

.selected {
  fill: #9CCC65;
  fill-opacity: 1;
}

#test_canvas {
  position: fixed;
  top: 0;
  left: 0;
  height: 100%;
  width: 100%; /* thx, http://www.sarasoueidan.com/blog/svg-coordinate-systems/ !!! */
  /*height: 100%;*/
}

</style>