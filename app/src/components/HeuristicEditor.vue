<template>
  <v-container class="fill-height" fluid>
    <v-btn @click="refreshData">Refresh</v-btn>
    <v-row align="center" justify="center">
      <v-col align="center" cols="12" sm="12" md="10" lg="9" xl="8">
        <svg id="test_canvas" viewBox="0 0 2000 2000"></svg>
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

// graph dimensions
const graphWidth = 600, graphHeight = 800;

// phantom node id
const rootIdentifier = 'root';

// click switch
let isClicked = false;

// root svg elements
let rootSvg, rootGroup;

let treeMap;

function dragstarted() {
  d3.select(this).raise();
}

function dragged(event) {
  const transformationMatrix = this.transform.baseVal.getItem(0).matrix;
  d3.select(this)
      .attr("transform",
          "translate(" + (transformationMatrix.e + event.dx) + "," + (transformationMatrix.f + event.dy) + ")");
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
          outText = `Type: ${d.data.data.type}`;
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


function drawNodes(group, nodeData) {
  // adds each node as a group
  return group.selectAll(".node")
      .data(nodeData.descendants(), d => d.data.data.uid)
      .join(enter => {
            const g = enter.append("g")
                .attr("transform", function (d) {
                  return "translate(" + d.y + "," + d.x + ")";
                })
                // set click handler
                .on('click', nodeClicked)
                // set drag handler
                .call(d3.drag()
                    .on("start", dragstarted)
                    .on("drag", dragged));
            // draw outline and text
            drawRect(g);
            return g;
          }
      )
      .attr("class", function (d) {
        if (d.data.data.uid === rootIdentifier)
          return null;

        return "node" +
            (d.children ? " node--internal" : " node--leaf");
      })
      ;
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
  let levelWidth = [1];
  const childCount = function (level, n) {

    if (n.children && n.children.length > 0) {
      if (levelWidth.length <= level + 1) levelWidth.push(0);

      levelWidth[level + 1] += n.children.length;
      n.children.forEach(function (d) {
        childCount(level + 1, d);
      });
    }
  };

  childCount(0, treeData);

  // declares a tree layout and assigns the size
  treeMap = d3.tree().size([d3.max(levelWidth) * 100, treeWidth]);
  // maps the node data to the tree layout
  nodes = treeMap(nodes);
  return nodes;
}

function createInitialGraph(height, width, margin) {
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
  console.log(height, width, margin);
  // add zoom and drag
  rootSvg.call(d3.zoom()
      .on("zoom", (event) => {
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
      // maps the node data to the tree layout
      const nodeData = processGraphData(this.data, graphHeight, graphWidth);
      drawLinks(rootGroup, nodeData);
      drawNodes(rootGroup, nodeData);
    },
    refreshData: async function () {
      await this.$store.dispatch('updateHeuristicData', this.$route.params.id);
      this.createGraph();
    }
  }
}
</script>

<style>
.node text {
  font: 12px sans-serif;
}

.node--internal text {
  text-shadow: 0 1px 0 #fff, 0 -1px 0 #fff, 1px 0 0 #fff, -1px 0 0 #fff;
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

#test_canvas {
  width: 100%; /* thx, http://www.sarasoueidan.com/blog/svg-coordinate-systems/ !!! */
}

</style>