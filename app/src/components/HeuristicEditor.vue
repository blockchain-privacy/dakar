<template>
  <v-container class="fill-height">
    <nested-menu
        v-model="contextMenu.display"
        origin="center center"
        :positionX="contextMenu.x"
        :positionY="contextMenu.y"
        absolute
        offset-y
        :close-on-click="true"
        style="max-width: 600px"
        name='File' :menu-items='contextMenu.items' @nested-menu-click='onMenuItemClick'/>
    <v-toolbar
        style="width: 100%; left:0;position:fixed; z-index: 99"
        dense>
      <v-toolbar-title class="hidden-md-and-up">
        {{ this.shortTransactionHash }}
      </v-toolbar-title>
      <v-toolbar-title class="hidden-sm-and-down">
        <v-icon>mdi-transfer</v-icon>
        Transaction {{ this.shortTransactionHash }}
      </v-toolbar-title>
      <v-btn icon @click="goToTransactionPage">
        <v-icon>mdi-open-in-new</v-icon>
      </v-btn>
      <v-spacer></v-spacer>
      <v-btn
          class="ml-1"
          outlined
          @click="isAddHeuristicSheetOpen = !isAddHeuristicSheetOpen"
          :disabled="this.executionStatus.value.executing">
        <v-icon>mdi-shape-square-rounded-plus</v-icon>
        <div class="hidden-sm-and-down">Add Heuristic</div>
      </v-btn>
      <v-btn
          class="ml-1"
          outlined
          @click="downloadHeuristicSummary"
          :disabled="this.executionStatus.value.executing || !doesDataExist()">
        <v-icon>mdi-file-download-outline</v-icon>
        <div class="hidden-sm-and-down">Download Summary</div>
      </v-btn>
      <v-btn
          class="ml-1"
          outlined
          @click="executeHeuristics"
          :disabled="this.executionStatus.value.executing || !isExecutable()">
        <v-icon>mdi-source-branch-check</v-icon>
        <div class="hidden-sm-and-down">Execute Heuristics</div>
      </v-btn>
      <v-bottom-sheet scrollable v-model="isAddHeuristicSheetOpen">
        <v-card>
          <div>
            <v-subheader class="float-left">Add heuristic</v-subheader>
            <v-switch
                class="float-right mr-2"
                v-model="isHeuristicSheetFixed"
                label="Fixed"
            ></v-switch>
          </div>
          <v-card-text style="height: 80%">
            <div class="d-flex flex-wrap" style="align-items: flex-start;">
              <v-card
                  class="mx-auto my-12"
                  v-for="(item, index) in heuristicTypes"
                  :key="index"
                  max-width="300">
                <v-card-title>
                  {{ item.title }}
                </v-card-title>
                <v-card-subtitle>
                  {{ item.description }}
                </v-card-subtitle>
                <v-card-subtitle>
                  <v-form v-model="item.parameter.valid" v-if="item.parameter !== undefined">
                    <v-text-field
                        v-model="item.parameter.value"
                        :rules="item.parameter.rule"
                        :label="item.parameter.description"
                        required>
                    </v-text-field>
                  </v-form>
                </v-card-subtitle>
                <v-card-actions class="pt-0">
                  <v-btn color="primary" @click="() => {
                    if (item.parameter !== undefined && !item.parameter.valid) {
                      return;
                    }
                    if (!isHeuristicSheetFixed) isAddHeuristicSheetOpen = false;
                    item.action(item);
                  }">
                    Add Heuristic
                  </v-btn>
                </v-card-actions>
              </v-card>
            </div>
          </v-card-text>
        </v-card>
      </v-bottom-sheet>
      <HeuristicDetails v-model="heuristicSheet.isOpen" :heuristic-data="heuristicSheet"
                        :address-map="heuristicDetailsMap.get(heuristicSheet.heuristicUid)"/>
    </v-toolbar>
    <v-overlay
        opacity="0.75"
        absolute
        :value="executionStatus.value.executing">
      <v-row justify="center">
        <v-col class="ma-5" v-if="!executionStatus.value.processing">
          Heuristics are waiting for processing.
          This may take several minutes depending on the chosen parameters and
          number of heuristics. You can wait or close this page and come back later.
        </v-col>
        <v-col class="ma-5" v-if="executionStatus.value.processing">
          Heuristics are executing now.
          This may take several minutes depending on the chosen parameters and
          number of heuristics. You can wait or close this page and come back later.
        </v-col>
      </v-row>
      <v-row justify="center">
        <v-progress-linear
            style="max-width: 350px;"
            indeterminate
            rounded
            height="6"
            :color="executionStatus.value.processing?'green darken-3':'primary'"
        ></v-progress-linear>
      </v-row>
    </v-overlay>
    <svg id="svg_canvas" viewBox="0 0 2000 2000"></svg>
  </v-container>
</template>

<script>
import HeuristicDetails from './HeuristicDetails.vue';
import {
  ROUTE_NAME_TRANSACTION_PAGE, ROUTE_EXECUTE_HEURISTICS,
  ROUTE_NAME_HEURISTIC_PAGE, ROUTE_HEURISTICS_SUMMARY,
  ROUTE_HEURISTIC_STATUS,
} from '../constants';
import NestedMenu from './common/NestedMenu.vue';
import * as ht from '../heuristicTree';
import { shortenHash, getCurrentDate } from '../utilities';

function getDeletedData(oldStateMap, newStateMap) {
  // search for deleted items
  const deletedUids = [];
  oldStateMap.forEach((value, key) => {
    if (!newStateMap.has(key)) {
      deletedUids.push(key);
    }
  });

  return deletedUids;
}

// prepareData prepares the heuristic data so it can be sent to be executed
function prepareData(oldStateMap, newState, changeSet, deletedData) {
  const changedItems = [];
  const newStateMap = new Map(newState.map((d) => [d.uid, d]));
  changeSet.forEach((changedUid) => {
    changedItems.push(newStateMap.get(changedUid));
  });

  const filteredData = [];
  // filter properties which do not need to be sent over the wire: timestamp and result count
  changedItems.forEach((d) => {
    // filter out the dummy element
    if (d.uid === ht.rootIdentifier) {
      return;
    }

    filteredData.push({
      uid: d.uid,
      type: d.type,
      parameter: d.parameter,
      children: d.children,
      parent_heuristic: d.parent_heuristic,
    });
  });

  return { changed: filteredData, deleted: deletedData };
}

function newRouting(context) {
  const { id } = context.$route.params;
  if (id === undefined || context.$route.name !== ROUTE_NAME_HEURISTIC_PAGE) {
    return;
  }

  context.onMounted();
}

function areDataElementsEqual(a, b) {
  if (a.uid !== b.uid || a.parameter !== b.parameter || a.type !== b.type) {
    return false;
  }

  if (a.parent_heuristic !== undefined && b.parent_heuristic !== undefined) {
    return a.parent_heuristic[0].uid === b.parent_heuristic[0].uid;
  }
  return !((a.parent_heuristic !== undefined && b.parent_heuristic === undefined)
      || (b.parent_heuristic !== undefined && a.parent_heuristic === undefined));
}

export default {
  name: 'HeuristicEditor',
  components: { HeuristicDetails, NestedMenu },
  data() {
    return {
      routeTransaction: ROUTE_NAME_TRANSACTION_PAGE,
      isHeuristicExecuting: false,
      executionStatus: {
        dormantTimer: {
          timer: null,
          refreshRate: 20000,
        },
        activeTimer: {
          timer: null,
          refreshRate: 2000,
        },
        value: {
          executing: false,
          processing: false,
        },
        enum: {
          added: 0,
          duplicate: 1,
          notInQueue: 2,
          inQueue: 3,
          processing: 4,
        },
      },
      newUidPrefix: 'newUid_',
      uidCounter: 1,
      transactionHash: '',
      shortTransactionHash: '',
      // dbState holds the state of the database.
      // It is used to detect changes in this.data.heuristics (computed)
      dbState: null,
      // changeSet holds all changes based on dbState and this.data.heuristics (computed)
      changeSet: [],
      // deletedData holds all uids of the heuristic which are deleted
      deletedData: [],
      isHeuristicSheetFixed: false,
      isAddHeuristicSheetOpen: false,
      // heuristicDetailsMap: map[heuristicUid]map[addressHash]array[originHash]
      heuristicDetailsMap: new Map(),
      heuristicSheet: {
        isOpen: false,
        heuristicUid: '',
        heuristicType: '',
        heuristicParameter: '',
        resultCount: null,
      },
      heuristicTypes: [
        {
          id: 'one_source',
          parameter: {
            value: 48,
            description: 'Look back time in hours',
            rule: [(v) => {
              if (!/^\d+$/.test(v)) return false;
              const num = parseInt(v, 10);
              return Number.isInteger(num) && num > 0;
            }],
            valid: false,
          },
          title: 'One Source',
          description: 'Filters by time, direct input transaction amount filter and omni sources',
          action: this.addNewHeuristic,
        },
        {
          id: 'global_amount',
          title: 'Global Amount',
          description: 'The amount heuristic filters all origins of sources, which do not have equal or '
              + 'more denominations to fund the destination transaction. '
              + 'Note that this is different from the direct input transaction amount filter, as '
              + 'this heuristic only checks the set of origin transactions and sources per destina- '
              + 'tion transaction, not per direct input transaction.',
          action: this.addNewHeuristic,
        },
        {
          id: 'perfect_match',
          title: 'Perfect Match',
          description: 'The perfect match heuristic filters all origins of sources, which have denominations '
              + 'without a perfect match for the denominations of the destination transaction.',
          action: this.addNewHeuristic,
        },
        {
          id: 'denomination_type',
          title: 'Denomination Type',
          description: 'The denomination type heuristic filters all origins of sources, which have denominations '
              + 'of types which do not occur in the denominations of the destination transaction.'
              + 'For example a destination transaction spends 5 × 10.0001 and 10 × 1.00001. '
              + 'Now all sources are excluded which do not have these exact two types of denominations.',
          action: this.addNewHeuristic,
        },
      ],
      contextMenu: {
        display: false,
        x: 0,
        y: 0,
        items: [
          { title: 'Delete Heuristic', icon: 'mdi-delete', action: this.deleteSubTree },
          { title: 'Show Properties', icon: 'mdi-chart-bar', action: ht.simulateClick },
          { isDivider: true },
          {
            title: 'Add Heuristic',
            icon: 'mdi-shape-square-rounded-plus',
            action: () => {
              this.isAddHeuristicSheetOpen = true;
            },
          },

          {
            title: 'Actions',
            menu: [
              {
                title: 'Download Summary',
                icon: 'mdi-file-download-outline',
                action: this.downloadHeuristicSummary,
                disabled: this.doesDataExist,
              },
              {
                title: 'Execute Heuristics',
                icon: 'mdi-source-branch-check',
                action: this.executeHeuristics,
                disabled: this.isExecutable,
              },
            ],
          },
        ],
      },
    };
  },
  computed: {
    errMsg: {
      get() {
        return this.$store.getters.getErrorMsg;
      },
      set(value) {
        this.$store.dispatch('setErrorMsg', value);
      },
    },
    data: {
      get() {
        return this.$store.getters.getHeuristicData;
      },
      set(value) {
        this.$store.dispatch('setHeuristicData', value);
      },
    },
    heuristicDetails: {
      get() {
        return this.$store.getters.getHeuristicDetails;
      },
      set(value) {
        if (value === null) {
          this.$store.dispatch('resetHeuristicDetails');
          return;
        }
        this.$store.dispatch('setHeuristicDetails', value);
      },
    },
    successMsg: {
      get() {
        return this.$store.getters.getSuccessMsg;
      },
      set(value) {
        this.$store.dispatch('setSuccessMsg', value);
      },
    },
  },
  methods: {
    addNewHeuristic(heuristic) {
      const newHeuristic = { type: heuristic.id, uid: `${this.newUidPrefix}${this.uidCounter}` };
      if (heuristic.parameter) {
        newHeuristic.parameter = `${heuristic.parameter.value}`;
      }
      this.uidCounter = this.uidCounter + 1;

      this.data.heuristics.push(newHeuristic);
      this.updateGraph();
    },
    openPropertySheet(heuristic) {
      const sheet = this.heuristicSheet;

      sheet.heuristicParameter = heuristic.parameter;
      sheet.heuristicType = heuristic.type;
      sheet.resultCount = heuristic.num_results;
      sheet.heuristicUid = heuristic.uid;

      // check if data must be loaded from backend
      if (heuristic.num_results === undefined || heuristic.num_results === 0
          || this.heuristicDetails.has(heuristic.uid)) {
        sheet.isOpen = true;
        return;
      }

      // request data from backend
      this.$store.dispatch('updateHeuristicDetails', {
        parameter: this.transactionHash,
        body: { uid: heuristic.uid },
      }).then(() => {
        if (this.heuristicDetails === null || this.heuristicDetails.length === 0
            || !this.heuristicDetails.has(heuristic.uid)) return;

        // results format: [{ts, addresshash, txhash}, ...]
        const { results } = this.heuristicDetails.get(heuristic.uid);

        if (results.length === 0) return;

        const addressMap = new Map();
        results.forEach((d) => {
          // if key already exists
          if (addressMap.has(d.addresshash)) {
            const origins = addressMap.get(d.addresshash);
            origins.push({ txhash: d.txhash, ts: d.ts });
            addressMap.set(d.addresshash, origins);
            return;
          }
          // new entry
          addressMap.set(d.addresshash, [{ txhash: d.txhash, ts: d.ts }]);
        });
        // append values to context variable
        this.heuristicDetailsMap.set(heuristic.uid, addressMap);

        sheet.isOpen = true;
      });
    },
    isExecutable() {
      if (this.dbState !== null && ((this.changeSet !== null && this.changeSet.length > 0))) {
        return true;
      }

      return this.deletedData.length > 0;
    },
    doesDataExist() {
      return !(!this.data || !this.data.heuristics || this.data.heuristics.length < 2);
    },
    executeHeuristics() {
      // prevent execution if not data is available
      if (!this.isExecutable()) {
        return;
      }

      fetch(ROUTE_EXECUTE_HEURISTICS + this.transactionHash, {
        method: 'POST', // or 'PUT'
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(prepareData(this.dbState, this.data.heuristics,
          this.changeSet, this.deletedData)),
      })
        .then((response) => response.json())
        .then((data) => {
          if (data.status === undefined) throw Error('execution status is not defined');
          this.setExecutionStatus(data.status);
          this.startActiveTimer();
        })
        .catch((error) => {
          this.errMsg = error;
        });
    },
    setExecutionStatus(status) {
      switch (status) {
        case this.executionStatus.enum.added:
          this.executionStatus.value.processing = false;
          this.executionStatus.value.executing = true;
          break;
        case this.executionStatus.enum.inQueue:
          this.executionStatus.value.processing = false;
          this.executionStatus.value.executing = true;
          break;
        case this.executionStatus.enum.processing:
          this.executionStatus.value.processing = true;
          this.executionStatus.value.executing = true;
          break;
        default:
          this.executionStatus.value.processing = false;
          this.executionStatus.value.executing = false;
      }
    },
    downloadHeuristicSummary() {
      if (!this.doesDataExist()) return;
      fetch(ROUTE_HEURISTICS_SUMMARY + this.transactionHash)
        .then((res) => res.blob())
        .then((blob) => {
          // looks hacky, but it is the only way with good UX
          const a = document.createElement('a');
          a.href = URL.createObjectURL(blob);

          a.setAttribute('download',
            `heuristic_summary_${getCurrentDate()}_${this.transactionHash}.csv`);
          a.click();
          a.remove();
        })
        .catch((error) => {
          this.errorMsg = error;
        });
    },
    // updateChangeSet updates the change set <this.changeSet> based
    // on the differences of this.data.heuristics and this.dbState
    updateChangeSet() {
      this.changeSet = [];
      const originChangeSet = [];
      this.data.heuristics.forEach((d) => {
        if (this.dbState.has(d.uid)) {
          const thisElement = this.dbState.get(d.uid);
          if (!areDataElementsEqual(thisElement, d)) {
            // changed element
            originChangeSet.push(d);
          }
        } else {
          // new element
          originChangeSet.push(d);
        }
      });

      // this set will have some duplicates, if changes are nested we get overlapping descendants
      const descendantSet = [];
      // find descendants for each changed root element
      originChangeSet.forEach((d) => {
        // get subtree
        const descendants = ht.getDescendants(d.uid);
        descendantSet.push(...descendants);
      });

      // remove duplicates
      const descendantMap = new Map(descendantSet.map(
        (tempObject) => [tempObject.data.data.uid, tempObject],
      ));

      // save in global changeSet
      descendantMap.forEach((e) => this.changeSet.push(e.data.data.uid));

      ht.setNodesChanged(descendantSet);
    },
    // called by context menu handler
    goToTransactionPage() {
      this.$router.push({ name: this.routeTransaction });
    },
    deleteSubTree() {
      const toBeRemoved = ht.getRemovableNodes();
      const rel = ht.getRemovableRelationship();

      const updatedData = [];

      this.data.heuristics.forEach((e) => {
        // update children set of parent
        if (rel.parentUid !== '' && e.uid === rel.parentUid) {
          e.children = e.children.filter((c) => c.uid !== rel.childUid);
        }

        // remove removable nodes
        if (!toBeRemoved.includes(e.uid)) {
          updatedData.push(e);
        }
      });

      this.data = { heuristics: updatedData, status: 0 };

      // this.$store.dispatch('setHeuristicData', );

      const newStateMap = new Map(this.data.heuristics.map((d) => [d.uid, d]));
      this.deletedData = getDeletedData(this.dbState, newStateMap);

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
      });
    },
    updateGraph() {
      // maps the node data to the tree layout
      const nodeData = ht.processGraphData(this.data.heuristics);
      ht.drawGraph(nodeData, this);
      // updateChangeSet is called after a graph update,
      // because otherwise it gets a not up to date descendant state
      this.updateChangeSet();
    },
    async refreshData() {
      await this.$store.dispatch('updateHeuristicData', this.transactionHash);
      this.setExecutionStatus(this.data.status);

      if (this.executionStatus.value.executing) this.startActiveTimer();

      // if the transaction has not yet any heuristics associated
      if (!this.data || !this.data.heuristics) {
        this.data.heuristics = [];
      }
      this.data.heuristics.push({ uid: 'root' });

      // reset deleted data
      this.deletedData = [];

      // deep copy of the array; Yes, this is how to do a deep copy in vanilla Javascript.
      // It's mind-boggling. As this.data.heuristics is effectively a JSON,
      // we can safely use the JSON functions (complex types like function are not allowed):
      this.dbState = new Map(JSON.parse(JSON.stringify(this.data.heuristics))
        .map((d) => [d.uid, d]));
      this.updateGraph();
    },
    onMounted() {
      const svgCanvasId = 'svg_canvas';
      // remove previous svg children
      document.getElementById(svgCanvasId).innerHTML = '';

      // set transaction hashes for this page view
      this.transactionHash = this.$route.params.id;
      this.shortTransactionHash = shortenHash(this.transactionHash);

      // set page title
      document.title = `Heuristic - ${this.transactionHash}`;

      if (!ht.setHeuristicClickHandler(this.openPropertySheet)) {
        this.errMsg = 'error setting heuristic click handler';
      }
      if (!ht.setContextMenuCallback(this.showContextMenu)) {
        this.errMsg = 'error setting context menu handler';
      }

      ht.setupSvg(this, svgCanvasId, this.heuristicTypes);
      this.refreshData();
      ht.centerGraph();
    },
    onMenuItemClick(item) {
      if (item.action) {
        item.action();
      }
      this.contextMenu.display = false;
    },
    async updateExecutionStatus() {
      await fetch(ROUTE_HEURISTIC_STATUS + this.transactionHash)
        .then((response) => response.json())
        .then((data) => {
          if (data.status === undefined) throw Error('execution status is not defined');
          const oldExecutionStatus = this.executionStatus.value.executing;
          this.setExecutionStatus(data.status);
          // if it was previously executing refresh data
          if (oldExecutionStatus && !this.executionStatus.value.executing) {
            this.refreshData();
            this.stopActiveTimer();
          }
        })
        .catch((error) => {
          this.errMsg = error;
        });
    },
    startDormantTimer() {
      this.executionStatus.dormantTimer.timer = setInterval(async () => {
        await this.updateExecutionStatus();
      }, this.executionStatus.dormantTimer.refreshRate);
    },
    startActiveTimer() {
      this.executionStatus.activeTimer.timer = setInterval(async () => {
        await this.updateExecutionStatus();
      }, this.executionStatus.activeTimer.refreshRate);
    },
    stopDormantTimer() {
      clearInterval(this.executionStatus.dormantTimer.timer);
    },
    stopActiveTimer() {
      clearInterval(this.executionStatus.activeTimer.timer);
    },
    resetExecutionStatus() {
      this.isHeuristicExecuting = false;
      this.stopDormantTimer();
      this.stopActiveTimer();
    },
  },
  beforeDestroy() {
    // reset memory
    this.heuristicDetails = null;
    this.resetExecutionStatus();
  },
  mounted() {
    this.onMounted();
    this.startDormantTimer();
  },
  watch: {
    $route() {
      newRouting(this);
    },
  },
};
</script>

<style>
.node text {
  font: 12px sans-serif;
  cursor: pointer;
}

.link {
  fill: none;
  stroke: darkslategrey;
  stroke-width: 2px;
}

.rect {
  stroke: #008ee5;
  fill-opacity: 0;
  cursor: pointer;
}

.clicked {
  stroke: #FDD835;
  fill: antiquewhite;
  fill-opacity: 1;
}

.graph-canvas {
  background-color: whitesmoke;
}

.modified {
  stroke-dasharray: 5;
}

.selected {
  fill: #9CCC65;
  fill-opacity: 1;
}

.valid-target {
  stroke: #2E7D32;
  stroke-width: 4px;
}

#svg_canvas {
  position: fixed;
  top: 0;
  left: 0;
  height: 100%;
  width: 100%; /* thx, http://www.sarasoueidan.com/blog/svg-coordinate-systems/ !!! */
}

</style>
