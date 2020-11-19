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
          <v-list-item-title>{{ "Delete sub tree" }}</v-list-item-title>
        </v-list-item>
        <!--        <v-list-item-->
        <!--            v-for="(item, index) in items"-->
        <!--            :key="index"-->
        <!--        >-->
        <!--          <v-list-item-title @click="item.action">{{ item.title }}</v-list-item-title>-->
        <!--        </v-list-item>-->
      </v-list>
    </v-menu>

    <v-navigation-drawer
        absolute
        permanent
        expand-on-hover
    >
      <v-list-item>
        <v-list-item-icon>
          <v-icon>mdi-shape-outline</v-icon>
        </v-list-item-icon>
        <v-list-item-title class="title">
          Heuristic Types
        </v-list-item-title>
      </v-list-item>

      <v-divider></v-divider>

      <v-list
          nav
          dense
          v-for="(item, index) in heuristicTypes"
          :key="index"
      >
        <v-list-item link>
          <v-list-item-icon>
            <v-icon>mdi-shape-square-rounded-plus</v-icon>
          </v-list-item-icon>
          <v-list-item-title>{{ item.title }}</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>


    <!--    <v-btn @click="refreshData">Refresh</v-btn>-->
    <!--    <v-btn @click="changeData">Change</v-btn>-->
    <!--    <v-row align="center" justify="center">-->
    <!--      <v-col align="center" cols="12" sm="12" md="10" lg="9" xl="8">-->
    <svg id="test_canvas" viewBox="0 0 2000 2000"></svg>
    <!--      </v-col>-->
    <!--    </v-row>-->


  </v-container>
</template>

<script>
import * as d3 from "d3";

const rectWidth = 150;

// set the dimensions and margins of the diagram
const svgMargin = {top: 20, right: 200, bottom: 20, left: 200},
    svgWidth = 1000 - svgMargin.left - svgMargin.right,
    svgHeight = 750 - svgMargin.top - svgMargin.bottom;

// phantom node id
const rootIdentifier = 'root';

// click switch
let isClicked = false;

// root svg elements
let rootSvg, rootGroup;

// tree layout function
let treeLayout;

// dragging
let dragActive = false, dragNode, dragLayoutData = null, setPointer = false;

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
    const linksToRemove = dragLayoutData.descendants(),
        nodesToRemove = linksToRemove.filter(d => d.data.data.uid !== dragLayoutData.data.data.uid);

    // remove the nodes
    rootSvg.selectAll(".node")
        .data(nodesToRemove, d => d.data.data.uid)
        .remove();
    rootSvg.selectAll(".link")
        .data(linksToRemove, d => d.data.data.uid)
        .remove();
  }

  const transformationMatrix = this.transform.baseVal.getItem(0).matrix;

  d3.select(this)
      .raise()  // causes bug in chrome: click is only recognized on second time. move here from dragStart
      .attr("transform",
          "translate(" + (transformationMatrix.e + event.dx) + "," + (transformationMatrix.f + event.dy) + ")");
}

function dragEnd(event, context) {
  dragNode = dragNode.attr("pointer-events", null);
  rootGroup.selectAll(".selected").classed("selected", false);

  // only move node if drag was active before --> not clicked and activeMouseOverNode is set
  if (activeMouseOverNode !== null && setPointer) {
    moveNode(context, activeMouseOverNode, dragNode);
  }

  lastMouseOverNode = null;
  dragLayoutData = null;
  setPointer = false;
  dragActive = false;


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
  if (dragActive) {
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
    lastMouseOverNode = null;
    activeMouseOverNode = null;
  }
}

function contextMenuHandler(context, event, d) {
  context.showContextMenu(event);
  activeContextMenuNode = d;
}

function drawNodes(group, nodeData, context) {
  // adds each node as a group
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
          }
      ).attr("transform", function (d) {
        return "translate(" + d.y + "," + d.x + ")";
      })
      .attr("class", function (d) {
        if (d.data.data.uid === rootIdentifier)
          return null;

        return "node" +
            (d.children ? " node--internal" : " node--leaf");
      });
}

function drawLinks(group, nodeData) {
  // adds the links between the nodes
  group.selectAll(".link")
      .data(nodeData.descendants().slice(1), d => d.data.data.uid)
      .join("path")
      .attr("class", "link")
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

function createInitialGraph(context, height, width, margin) {
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
      .attr("class", "root-group")
      .attr("transform", "translate(" + margin.left + "," + margin.top + ")")

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

export default {
  name: "HeuristicEditor",
  data: () => ({
    heuristicTypes: [
      {
        id: "one_source",
        title: "one source",
      },
      {
        id: "global_amount",
        title: "global amount",
      },
      {
        id: "perfect_match",
        title: "perfect match",
      },
      {
        id: "denomination_type",
        title: "denomination type",
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
  mounted() {
    document.title = `Heuristic - ${this.$route.params.id}`;
    this.heuristicTypes.forEach(e => heuristicTypeMap.set(e.id, e.title));
    createInitialGraph(this, svgHeight, svgWidth, svgMargin);
    this.refreshData();
  },
  methods: {
    // called by context menu handler
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

      // remove the nodes
      rootSvg.selectAll(".node")
          .data(nodesToRemove, d => d.data.data.uid)
          .remove();
      rootSvg.selectAll(".link")
          .data(nodesToRemove, d => d.data.data.uid)
          .remove();
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
      await this.$store.dispatch('updateHeuristicData', this.$route.params.id);
      addRootElement(this.data);
      this.updateGraph();
    },
    changeData: function () {
      this.updateGraph();
    }
  }
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
  width: 100%; /* thx, http://www.sarasoueidan.com/blog/svg-coordinate-systems/ !!! */
  /*height: 100%;*/
}

</style>