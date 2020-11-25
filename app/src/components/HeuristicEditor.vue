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
      <v-btn outlined @click="sheetOpen = !sheetOpen">
        <v-icon>mdi-shape-square-rounded-plus</v-icon>
        <div class="hidden-sm-and-down">Add Heuristic</div>
      </v-btn>
      <v-menu
          bottom
          left
      >
        <template v-slot:activator="{ on, attrs }">
          <v-btn
              icon
              v-bind="attrs"
              v-on="on"
          >
            <v-icon>mdi-dots-vertical</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item @click="downloadHeuristicSummary">
            <v-list-item-icon>
              <v-icon>mdi-file-download-outline</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Download Summary</v-list-item-title>
          </v-list-item>
          <v-list-item @click="executeHeuristics" :disabled="this.data && this.data.length < 2">
            <v-list-item-icon>
              <v-icon>mdi-source-branch-check</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Execute Heuristics</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
      <v-bottom-sheet scrollable v-model="sheetOpen">
        <v-card v-touch="{
      down: () => {this.sheetOpen = false;}
    }">
          <v-subheader>Add heuristic</v-subheader>
          <v-card-text style="height: 80%">
            <div class="d-flex flex-wrap" style="align-items: flex-start;">
              <v-card
                  class="mx-auto my-12"
                  v-for="(item, index) in heuristicTypes"
                  :key="index"
                  max-width="300"
              >
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
                <v-card-actions class="pt-0">
                  <v-btn color="primary" @click="sheetOpen = false">Add Heuristic</v-btn>
                </v-card-actions>
              </v-card>
            </div>
          </v-card-text>
        </v-card>
      </v-bottom-sheet>
    </v-toolbar>
    <!--    todo: remove?-->
    <!--    <v-btn @click="refreshData">Refresh</v-btn>-->
    <!--    <v-btn @click="changeData">Change</v-btn>-->
    <svg id="svg_canvas" viewBox="0 0 2000 2000"></svg>
  </v-container>
</template>

<script>
import {shortenHash} from "@/utilities";
import {ROUTE_NAME_SEARCH_PAGE, ROUTE_EXECUTE_HEURISTICS} from "@/constants";
import * as ht from "@/heuristicTree";

// prepareData prepares the heuristic data so it can be sent to be executed
function prepareData(data) {
  let filteredData = [];

  // filter properties which do not need to be sent over the wire: timestamp and result count
  data.forEach(d => {
    // filter out the dummy element
    if (d.uid === ht.rootIdentifier) {
      return;
    }

    filteredData.push({
      uid: d.uid,
      type: d.type,
      parameter: d.parameter,
      children: d.children,
      parent_heuristic: d.parent_heuristic
    });
  });

  return filteredData;
}

export default {
  name: "HeuristicEditor",
  data: () => ({
    sheetOpen: false,
    transactionHash: "",
    shortTransactionHash: "",
    heuristicTypes: [
      {
        id: "one_source",
        title: "One Source",
        description: "Filters by time, direct input transaction amount filter and omni sources",
        fun: ht.navDrag,
      },
      {
        id: "global_amount",
        title: "Global Amount",
        description: "The amount heuristic filters all origins of sources, which do not have equal or " +
            "more denominations to fund the destination transaction. " +
            "Note that this is different from the direct input transaction amount filter, as " +
            "this heuristic only checks the set of origin transactions and sources per destina- " +
            "tion transaction, not per direct input transaction.",
        fun: ht.navDrag,
      },
      {
        id: "perfect_match",
        title: "Perfect Match",
        description: "The perfect match heuristic filters all origins of sources, which have denominations " +
            "without a perfect match for the denominations of the destination transaction.",
        fun: ht.navDrag,
      },
      {
        id: "denomination_type",
        title: "Denomination Type",
        description: "The denomination type heuristic filters all origins of sources, which have denominations " +
            "of types which do not occur in the denominations of the destination transaction." +
            "For example a destination transaction spends 5 × 10.0001 and 10 × 1.00001. " +
            "Now all sources are excluded which do not have these exact two types of denominations.",
        fun: ht.navDrag,
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
    }
  },
  methods: {
    executeHeuristics() {
      // prevent execution if not data is available
      if (this.data.length < 2) {
        return
      }

      fetch(ROUTE_EXECUTE_HEURISTICS + this.transactionHash, {
        method: 'POST', // or 'PUT'
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(prepareData(this.data)),
      })
          .then(response => response.json())
          .then(data => {
            console.log('Success:', data);
          })
          .catch((error) => {
            console.error('Error:', error);
          });
    },
    downloadHeuristicSummary() {
      // fetch(ROUTE_HEURISTICS_SUMMARY + this.transactionHash)
      //     .then(response => response.json())
      //     .then(data => {
      //       console.log('Success:', data);
      //     })
      //     .catch((error) => {
      //       console.error('Error:', error);
      //     });
    },
    // called by context menu handler
    goToTransactionPage() {
      this.$router.push({name: ROUTE_NAME_SEARCH_PAGE})
    },
    deleteSubTree() {
      const toBeRemoved = ht.getRemovableNodes();
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
      const nodeData = ht.processGraphData(this.data);
      ht.drawGraph(nodeData, this);
    },
    refreshData: async function () {
      await this.$store.dispatch('updateHeuristicData', this.transactionHash);
      ht.addRootElement(this.data);
      this.updateGraph();
    },
    changeData() {
      this.updateGraph();
    },
  },
  mounted() {
    // set transaction hashes for this page view
    this.transactionHash = this.$route.params.id;
    this.shortTransactionHash = shortenHash(this.transactionHash);

    // set page title
    document.title = `Heuristic - ${this.transactionHash}`;

    ht.setupSvg(this, "svg_canvas", this.heuristicTypes);
    this.refreshData();
    ht.centerGraph();
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

#svg_canvas {
  position: fixed;
  top: 0;
  left: 0;
  height: 100%;
  width: 100%; /* thx, http://www.sarasoueidan.com/blog/svg-coordinate-systems/ !!! */
  /*height: 100%;*/
}

</style>