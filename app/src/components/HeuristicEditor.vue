<template>
  <v-container class="fill-height" fluid>
    <v-btn @click="refreshData">Refresh</v-btn>
    <v-row align="center" justify="center">
      <v-col align="center" cols="12" sm="12" md="10" lg="9" xl="8">
        <svg id="test_canvas"></svg>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import * as d3 from "d3";

const rectWidth = 150;

// set the dimensions and margins of the diagram
const svgMargin = {top: 20, right: 200, bottom: 20, left: 200},
    svgWidth = 1000 - svgMargin.left - svgMargin.right,
    svgHeight = 750 - svgMargin.top - svgMargin.bottom;

const graphWidth = 600, graphHeight = 800;

const rootIdentifier = 'root';

function drawRect(rootElement) {
  const textAreaHeight = 50, textPadding = 10, rectHeight = textAreaHeight + 2 * textPadding,
      borderRadius = 5, strokeWidth = 2, textHeight = 10;
  rootElement
      .append("rect")
      .attr("x", -rectWidth / 2)
      .attr("y", -rectHeight / 2)
      .attr("width", rectWidth)
      .attr("height", rectHeight)
      .attr("rx", borderRadius)
      .attr("ry", borderRadius)
      .attr("fill-opacity", 0)
      .attr("stroke-width", strokeWidth)
      .attr("stroke-opacity", d => {
        // only draw rect if it is not the root node
        if (d.data.data.uid !== rootIdentifier)
          return 1;
        return 0;
      });


  rootElement.append("svg:text")
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
          outText = `Type: ${d.data.data.type}`;
        }

        return outText;
      });

  rootElement.append("svg:text")
      .attr("x", function () {
        return -rectWidth / 2 + strokeWidth + 2;
      })
      .attr("y", textAreaHeight / 2 - textHeight)
      .text(function (d) {
        return d.data.data.parameter !== undefined ? `Parameter: ${d.data.data.parameter}` : null;
      });
}


function drawNodes(group, nodeData) {
  // adds each node as a group
  const n = group.selectAll(".node")
      .data(nodeData.descendants(), d => d.data.data.uid)
      .join(enter => {
        const g = enter.append("g");
        drawRect(g);
        return g;
      })
      .attr("class", function (d) {
        return "node" +
            (d.children ? " node--internal" : " node--leaf");
      })
      .attr("transform", function (d) {
        return "translate(" + d.y + "," + d.x + ")";
      });

  // console.log(perElement);

  return n;
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
          return 0.5;
        return 0;
      })
      .attr("d", function (d) {
        return "M" + (d.y - rectWidth / 2) + "," + d.x
            + "C" + (d.y + d.parent.y) / 2 + "," + d.x
            + " " + (d.y + d.parent.y) / 2 + "," + d.parent.x
            + " " + (d.parent.y + rectWidth / 2) + "," + d.parent.x;
      });
}

function processGraphData(graphData, treeHeight, treeWidth) {
  let dataWithRoot = graphData;

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

  dataWithRoot.push({'uid': 'root'});

  const treeData = stratifyData(dataWithRoot);

  //  assigns the data to a hierarchy using parent-child relationships
  let nodes = d3.hierarchy(treeData, function (d) {
    return d.children;
  });

  // declares a tree layout and assigns the size
  const treemap = d3.tree().size([treeHeight, treeWidth]);
  console.log(document.body.clientHeight, document.body.clientWidth);
  // maps the node data to the tree layout
  nodes = treemap(nodes);
  return nodes;
}

function createInitialGraph(height, width, margin) {
  // append the svg object to the body of the page
  // appends a 'group' element to 'svg'
  // moves the 'group' element to the top left margin
  const svg = d3.select("#test_canvas")
          .attr("width", width + margin.left + margin.right)
          .attr("height", height + margin.top + margin.bottom)
          .attr("class", "graph-canvas"),
      g = svg
          .append("g")
          .attr("class", "root-group")
          .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  // add zoom and drag
  svg.call(d3.zoom()
      .on("zoom", (event) => {
        g.attr('transform', event.transform);
      })
      .scaleExtent([1, 3])
  );
}

export default {
  name: "HeuristicEditor",
  computed: {
    data() {
      return this.$store.getters.getHeuristicData;
    },
  },
  mounted() {
    document.title = `Heuristic - ${this.$route.params.id}`;
    createInitialGraph(svgHeight, svgWidth, svgMargin);
    this.refreshData();
  },
  methods: {
    createGraph() {
      const g = d3.select(".root-group");

      // maps the node data to the tree layout
      const nodeData = processGraphData(this.data, graphHeight, graphWidth);

      drawLinks(g, nodeData);

      drawNodes(g, nodeData);

      // drawRectangle(nodes);
    },
    refreshData: async function () {
      await this.$store.dispatch('updateHeuristicData', this.$route.params.id);
      this.createGraph();
    }
  },
  created() {


  },
}
</script>

<style>
.node circle {
  fill: #fff;
  stroke: steelblue;
  stroke-width: 3px;
}

.node text {
  font: 12px sans-serif;
}

.node--internal text {
  text-shadow: 0 1px 0 #fff, 0 -1px 0 #fff, 1px 0 0 #fff, -1px 0 0 #fff;
}

.link {
  fill: none;
  stroke: black;
  stroke-width: 2px;
}

rect {
  stroke: #008ee5;
}

.graph-canvas {
  background-color: whitesmoke;
}

</style>