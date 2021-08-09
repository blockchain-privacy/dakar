<template>
  <div class="flex-column d-flex" style="height: 100%;">
    <v-banner v-if="banner.show"
              v-model="banner.display"
              transition="slide-y-transition"
              color="warning">
      <v-avatar slot="icon" size="40">
        <v-icon icon="mdi-lock">{{ this.icon.mdiAlertOctagon }}</v-icon>
      </v-avatar>
      Server is not ready to accept request for new heuristics.
      Please try again later. Existing heuristic results can be viewed.
      <template v-slot:actions="{ dismiss }">
        <v-btn outlined @click="dismiss">Dismiss</v-btn>
      </template>
    </v-banner>
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
        dense dark
        color="primary"
        style="z-index: 10; box-shadow: 0 2px 4px -1px rgba(0, 0, 0, 0.2);">
      <v-toolbar-title class="hidden-md-and-up">
        {{ this.transactionHash }}
      </v-toolbar-title>
      <v-toolbar-title class="hidden-sm-and-down">
        <v-icon>{{ icon.mdiTransfer }}</v-icon>
        Transaction {{ this.transactionHash }}
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
          style="min-width: 32px !important;"
          class="ml-1 pa-2"
          outlined
          @click="isAddHeuristicSheetOpen = !isAddHeuristicSheetOpen"
          :disabled="this.banner.show || this.executionStatus.value.executing">
        <v-icon>{{ icon.mdiShapeSquareRoundedPlus }}</v-icon>
        <div class="hidden-sm-and-down">Add Heuristic</div>
      </v-btn>
      <v-btn
          style="min-width: 32px !important;"
          class="ml-1 pa-2"
          outlined
          @click="downloadHeuristicSummary"
          :disabled="this.executionStatus.value.executing || !doesDataExist()">
        <v-icon>{{ icon.mdiFileDownloadOutline }}</v-icon>
        <div class="hidden-sm-and-down">Summary</div>
      </v-btn>
      <v-btn
          style="min-width: 32px !important;"
          class="ml-1 pa-2"
          outlined
          @click="executeHeuristics"
          :disabled="this.banner.show || this.executionStatus.value.executing || !isExecutable()">
        <v-icon>{{ icon.mdiSourceBranchCheck }}</v-icon>
        <div class="hidden-sm-and-down">Execute</div>
      </v-btn>
      <v-menu bottom>
        <template v-slot:activator="{ on, attrs }">
          <v-btn icon v-bind="attrs" v-on="on" style="outline: 0">
            <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item @click="goToTransactionPage">
            <v-list-item-icon>
              <v-icon>{{ icon.mdiOpenInNew }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Transaction Page</v-list-item-title>
          </v-list-item>
          <v-list-item @click="goToHeuristicOverviewPage">
            <v-list-item-icon>
              <v-icon>{{ icon.mdiOpenInNew }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Heuristic Overview</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
      <TypeSelection
          v-model="isAddHeuristicSheetOpen"
          :tab-items="heuristicTabItems"
          :descriptors="heuristicDescriptors"
          v-on:add-heuristic="addNewHeuristic"
      />
      <Details
          v-model="heuristicSheet.isOpen"
          :heuristic-data="heuristicSheet"
          :new-heuristic-prefix="this.newUidPrefix"
          :cluster-to-transactions="heuristicDetailsMap.get(heuristicSheet.heuristicUid)"/>
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
    <svg id="svg_canvas"/>
  </div>
</template>

<script>
import {
  mdiTransfer, mdiOpenInNew, mdiShapeSquareRoundedPlus, mdiFileDownloadOutline,
  mdiSourceBranchCheck, mdiDelete, mdiChartBar, mdiShapeSquarePlus, mdiDotsVertical,
  mdiAlertOctagon,
} from '@mdi/js';
import TypeSelection from './TypeSelection.vue';
import Details from './Details.vue';
import {
  ROUTE_NAME_TRANSACTION_PAGE, ROUTE_EXECUTE_HEURISTICS,
  ROUTE_NAME_HEURISTIC_PAGE, ROUTE_HEURISTICS_SUMMARY,
  ROUTE_HEURISTIC_STATUS, ROUTE_NAME_USER_HEURISTIC_PAGE,
  ROUTE_HEURISTIC_DESCRIPTORS, ROUTE_HEURISTICS,
} from '../../constants';
import NestedMenu from '../common/NestedMenu.vue';
import * as ht from '../../heuristicTree';
import {
  getCurrentDate, doPost, doGet, handleError,
} from '../../utilities';

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
  name: 'Editor',
  components: { TypeSelection, Details, NestedMenu },
  data() {
    return {
      icon: {
        mdiTransfer,
        mdiOpenInNew,
        mdiShapeSquareRoundedPlus,
        mdiFileDownloadOutline,
        mdiSourceBranchCheck,
        mdiDotsVertical,
        mdiAlertOctagon,
      },
      routeTransaction: ROUTE_NAME_TRANSACTION_PAGE,
      routeHeuristicOverview: ROUTE_NAME_USER_HEURISTIC_PAGE,
      routeHeuristicDescriptors: ROUTE_HEURISTIC_DESCRIPTORS,
      isHeuristicExecuting: false,
      banner: {
        // show is the switch for the warning banner
        // which gets displayed if the heuristic worker is not ready to accept requests
        show: false,
        display: true,
      },
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
          notReady: 5,
        },
      },
      isLinkMenuOpen: false,
      newUidPrefix: 'newUid_',
      uidCounter: 1,
      transactionHash: '',
      // dbState holds the state of the database.
      // It is used to detect changes in this.data.heuristics (computed)
      dbState: null,
      // changeSet holds all changes based on dbState and this.data.heuristics (computed)
      changeSet: [],
      // deletedData holds all uids of the heuristic which are deleted
      deletedData: [],
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
      heuristicDescriptors: [],
      heuristicTabItems: [],
      contextMenu: {
        display: false,
        x: 0,
        y: 0,
        items: [
          {
            title: 'Delete Heuristic',
            icon: mdiDelete,
            action: this.deleteSubTree,
            disabled: () => !this.banner.show,
          },
          { title: 'Show Properties', icon: mdiChartBar, action: ht.simulateClick },
          { isDivider: true },
          {
            title: 'Add Heuristic',
            icon: mdiShapeSquarePlus,
            action: () => {
              this.isAddHeuristicSheetOpen = true;
            },
            disabled: () => !this.banner.show,
          },
          {
            title: 'Actions',
            menu: [
              {
                title: 'Download Summary',
                icon: mdiFileDownloadOutline,
                action: this.downloadHeuristicSummary,
                disabled: this.doesDataExist,
              },
              {
                title: 'Execute Heuristics',
                icon: mdiSourceBranchCheck,
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
  },
  methods: {
    setErrorMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'error', temporary: true });
    },
    setInfoMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'info', temporary: true });
    },
    addNewHeuristic(heuristic) {
      const newHeuristic = { type: heuristic.type, uid: `${this.newUidPrefix}${this.uidCounter}` };
      if (heuristic.parameter) {
        newHeuristic.parameter = `${heuristic.parameter.value}`;
      }
      this.uidCounter += 1;

      this.data.heuristics.push(newHeuristic);
      this.updateGraph();
    },
    openPropertySheet(heuristic) {
      const sheet = this.heuristicSheet;

      // lookup type title from type id
      let displayType = '';
      this.heuristicDescriptors.some((d) => {
        if (d.type === heuristic.type) {
          displayType = d.title;
          return true;
        }
        return false;
      });

      sheet.heuristicParameter = heuristic.parameter;
      sheet.heuristicType = displayType;
      sheet.resultCount = heuristic.num_results;
      sheet.heuristicUid = heuristic.uid;

      // check if data has to be loaded from backend
      if (heuristic.num_results === undefined || heuristic.num_results === 0
          || this.heuristicDetails.has(heuristic.uid)
          || heuristic.uid.startsWith(this.newUidPrefix)) {
        sheet.isOpen = true;
        return;
      }

      // request data from backend
      this.$store.dispatch('updateHeuristicDetails', { uid: heuristic.uid }).then(() => {
        if (this.heuristicDetails === null || this.heuristicDetails.length === 0
            || !this.heuristicDetails.has(heuristic.uid)) return;

        const { results } = this.heuristicDetails.get(heuristic.uid);

        if (results.length === 0) return;

        this.heuristicDetailsMap.set(heuristic.uid, results);

        sheet.isOpen = true;
      });
    },
    isExecutable() {
      if (!this.banner.show && this.dbState !== null
          && ((this.changeSet !== null && this.changeSet.length > 0))) {
        return true;
      }

      return this.deletedData.length > 0;
    },
    doesDataExist() {
      if (!this.data || !this.data.heuristics) return false;

      // count elements which are not root or non-executed
      const numElements = this.data.heuristics.reduce(
        (sum, e) => (e.uid.startsWith(this.newUidPrefix) || e.uid === 'root' ? sum : sum + 1), 0,
      );

      return numElements > 0;
    },
    executeHeuristics() {
      // prevent execution if not data is available
      if (!this.isExecutable()) {
        return;
      }

      doPost(ROUTE_EXECUTE_HEURISTICS, this.$router, this.$store,
        prepareData(this.dbState, this.data.heuristics, this.changeSet, this.deletedData),
        this.transactionHash)
        .then((data) => {
          if (data.success === false) {
            if (data.msg) throw new Error(data.msg);
            throw new Error('execution did not succeed');
          }

          if (data.msg) this.setInfoMessage(data.msg);

          if (data.status === undefined && !data.msg) throw new Error('execution status is not defined');
          if (data.status === undefined && data.msg) return;
          this.setExecutionStatus(data.status);
          this.startActiveTimer();
        })
        .catch((error) => {
          this.setErrorMessage(error);
        });
    },
    setExecutionStatus(status) {
      switch (status) {
        case this.executionStatus.enum.added:
          this.executionStatus.value.processing = false;
          this.executionStatus.value.executing = true;
          this.banner.show = false;
          break;
        case this.executionStatus.enum.inQueue:
          this.executionStatus.value.processing = false;
          this.executionStatus.value.executing = true;
          this.banner.show = false;
          break;
        case this.executionStatus.enum.processing:
          this.executionStatus.value.processing = true;
          this.executionStatus.value.executing = true;
          this.banner.show = false;
          break;
        case this.executionStatus.enum.notReady:
          this.executionStatus.value.processing = false;
          this.executionStatus.value.executing = false;
          this.banner.show = true;
          break;
        default:
          this.executionStatus.value.processing = false;
          this.executionStatus.value.executing = false;
          this.banner.show = false;
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
    goToHeuristicOverviewPage() {
      this.$router.push({ name: this.routeHeuristicOverview });
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
    loadHeuristicData(transactionHash) {
      return doGet(ROUTE_HEURISTICS, this.$router, this.$store, transactionHash).then((data) => {
        this.data = data;
        this.$store.dispatch('resetMessages');
      }).catch((e) => {
        handleError(this.$store, e);
        return e;
      });
    },
    async refreshData() {
      await this.loadHeuristicData(this.transactionHash);

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
    async getDescriptors() {
      await doGet(this.routeHeuristicDescriptors, this.$router, this.$store)
        .then((data) => {
          if (data.success === undefined) throw Error('error getting heuristic descriptors');
          if (data.success === false) {
            throw Error(data.msg);
          }

          if (!data.descriptors) {
            throw Error('heuristic descriptor list is empty');
          }

          // add valid property
          this.heuristicDescriptors = data.descriptors.map((e) => {
            if (e.parameter) {
              e.parameter.valid = false;
            }

            return e;
          });
        })
        .catch((error) => {
          this.setErrorMessage(error);
        });
    },
    createTabs() {
      const tabSet = new Set();
      let isCategoryEmpty = false;
      this.heuristicDescriptors.forEach((e) => {
        if (e.category) {
          tabSet.add(e.category);
        } else {
          // if no category is set
          isCategoryEmpty = true;
        }
      });
      this.heuristicTabItems = Array.from(tabSet).sort();

      if (isCategoryEmpty) {
        this.heuristicTabItems.push('Other');
      }
    },
    async onMounted() {
      const svgCanvasId = 'svg_canvas';
      // remove previous svg children
      document.getElementById(svgCanvasId).innerHTML = '';

      // set transaction hashes for this page view
      this.transactionHash = this.$route.params.id;

      // set page title
      document.title = `Heuristic ${this.transactionHash}`;

      if (!ht.setHeuristicClickHandler(this.openPropertySheet)) {
        this.setErrorMessage('error setting heuristic click handler');
      }
      if (!ht.setContextMenuCallback(this.showContextMenu)) {
        this.setErrorMessage('error setting context menu handler');
      }

      // gets all heuristic type configurations
      await this.getDescriptors();

      // creates the tab descriptions based on the heuristic categories
      this.createTabs();

      ht.setupSvg(this, svgCanvasId, this.heuristicDescriptors);
      await this.refreshData();
      await ht.centerGraph();
    },
    onMenuItemClick(item) {
      if (item.action) {
        item.action();
      }
      this.contextMenu.display = false;
    },
    async updateExecutionStatus() {
      await doGet(ROUTE_HEURISTIC_STATUS, this.$router, this.$store, this.transactionHash)
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
          this.setErrorMessage(error);
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

<style scoped>
>>> .node text {
  font: 12px sans-serif;
  cursor: pointer;
}

>>> .link {
  fill: none;
  stroke: darkslategrey;
  stroke-width: 2px;
}

>>> .rect {
  stroke: #008ee5;
  fill-opacity: 0;
  cursor: pointer;
}

>>> .clicked {
  stroke: #FDD835;
}

>>> .modified {
  stroke-dasharray: 5;
}

>>> .selected {
  fill: #9CCC65;
  fill-opacity: 1;
}

>>> .valid-target {
  stroke: #2E7D32;
  stroke-width: 4px;
}

>>> #svg_canvas {
  height: 100%;
}

>>> .v-toolbar__content, .v-toolbar__extension {
  padding-right: 0
}
</style>
