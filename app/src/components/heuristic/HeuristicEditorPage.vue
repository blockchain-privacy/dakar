<template>
  <div
    class="flex-column d-flex"
    style="height: 100%;"
  >
    <v-expand-transition>
      <div
        v-if="banner.show && banner.display"
        class="d-flex align-center py-2"
        style="background-color: #FB8C00; z-index: 1004"
      >
        <v-icon
          :icon="icon.mdiAlertOctagon"
          size="large"
          class="mx-2"
          color="white"
        />
        <p
          class="text-h6"
          style="color:white"
        >
          Server is not ready to accept request for new heuristics. Please try again later. Existing heuristic results can be viewed.
        </p>
        <v-btn
          variant="text"
          class="ms-auto"
          @click="banner.display = false"
        >
          Dismiss
        </v-btn>
      </div>
    </v-expand-transition>
    <nested-menu
      v-model="contextMenu.display"
      origin="center center"
      :position-x="contextMenu.x"
      :position-y="contextMenu.y"
      absolute
      is-offset
      :close-on-click="true"
      style="max-width: 600px"
      name="File"
      :menu-items="contextMenu.items"
      @nested-menu-click="onMenuItemClick"
    />
    <v-toolbar
      density="compact"
      color="rgb(var(--v-theme-surface))"
      class="heuristicToolbar"
    >
      <v-toolbar-title class="hidden-md-and-up">
        {{ transactionHash }}
      </v-toolbar-title>
      <v-toolbar-title class="hidden-sm-and-down">
        <v-icon>{{ icon.mdiTransfer }}</v-icon>
        Transaction {{ transactionHash }}
      </v-toolbar-title>

      <v-btn
        style="min-width: 32px !important;"
        class="ms-3 pa-2"
        variant="outlined"
        :disabled="banner.show || executionStatus.value.executing"
        @click="openTypeSelectionSheet"
      >
        <v-icon>{{ icon.mdiShapeSquareRoundedPlus }}</v-icon>
        <div class="hidden-sm-and-down">
          Add Heuristic
        </div>
      </v-btn>
      <v-btn
        style="min-width: 32px !important;"
        class="ms-3 pa-2"
        variant="outlined"
        :disabled="banner.show || executionStatus.value.executing || !isExecutable()"
        @click="executeHeuristics"
      >
        <v-icon>{{ icon.mdiSourceBranchCheck }}</v-icon>
        <div class="hidden-sm-and-down">
          Execute
        </div>
      </v-btn>
      <v-menu location="bottom">
        <template #activator="{ props }">
          <v-btn
            icon
            v-bind="props"
            style="outline: 0"
          >
            <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item :to="{ name: routeTransaction, params: { id: transactionHash }}">
            <template #prepend>
              <v-icon>{{ icon.mdiOpenInNew }}</v-icon>
            </template>
            <v-list-item-title>Transaction Page</v-list-item-title>
          </v-list-item>
          <v-list-item :to="{ name: routeHeuristicOverview}">
            <template #prepend>
              <v-icon>{{ icon.mdiOpenInNew }}</v-icon>
            </template>
            <v-list-item-title>Heuristic Overview</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-toolbar>
    <!-- position: relative; is needed so the dialog is contained in its parent -->
    <div style="position: relative; height: 100%; width: 100%; overflow: hidden">
      <v-dialog
        :model-value="executionStatus.value.executing"
        :persistent="true"
        max-width="700px"
        :contained="true"
        :no-click-animation="true"
      >
        <v-card>
          <v-card-text class="text-subtitle-1 d-flex align-center">
            <v-icon
              :icon="icon.mdiTimerSand"
              size="50"
              class="me-3"
            />
            <div>
              <p
                v-if="executionStatus.value.processing"
                class="text-center mb-3"
              >
                Heuristics are executing now
              </p>
              <p
                v-else
                class="text-center mb-3"
              >
                Heuristics are waiting to be processed
              </p>
              This may take several minutes depending on the chosen parameters and
              number of heuristics. You can wait or close this page and come back later.
              <v-progress-linear
                class="mt-3"
                :indeterminate="true"
                rounded
                :color="executionStatus.value.processing?'primary':''"
              />
            </div>
          </v-card-text>
        </v-card>
      </v-dialog>
      <heuristic-details-sidebar
        v-model="heuristicSheet.isOpen"
        :heuristic-data="heuristicSheet"
        :new-heuristic-prefix="newUidPrefix"
        :clusters="heuristicSheet.clusters"
      />
      <heuristic-type-selection-side-bar
        v-model="isAddHeuristicSheetOpen"
        :tab-items="heuristicTabItems"
        :descriptors="heuristicDescriptors"
        @add-heuristic="addNewHeuristic"
      />
      <svg
        id="svg_canvas"
        style="width: 100%; height:100%"
      />
    </div>
  </div>
</template>

<script>
import {
	mdiAlertOctagon,
	mdiChartBar,
	mdiDelete,
	mdiDotsVertical,
	mdiFileDownloadOutline,
	mdiOpenInNew,
	mdiShapeSquarePlus,
	mdiShapeSquareRoundedPlus,
	mdiSourceBranchCheck,
	mdiTransfer,
	mdiTimerSand,
} from '@mdi/js';
import HeuristicTypeSelectionSideBar from './HeuristicTypeSelectionSideBar.vue';
import {
	APPLICATION_NAME,
	CLUSTER_TYPE_CUSTOM,
	ROUTE_NAME_HEURISTIC_PAGE,
	ROUTE_NAME_TRANSACTION_PAGE,
	ROUTE_NAME_USER_HEURISTIC_PAGE,
} from '@/constants';
import NestedMenu from '../common/NestedMenu.vue';
import {HeuristicTree, rootIdentifier} from '@/d3Documents/heuristicTree';
import {handleError} from '@/utilities';
import HeuristicDetailsSidebar from '@/components/heuristic/HeuristicDetailsSideBar.vue';

function getDeletedData(oldStateMap, newStateMap) {
	// Search for deleted items
	const deletedUIDs = [];
	oldStateMap.forEach((value, key) => {
		if (!newStateMap.has(key)) {
			deletedUIDs.push(key);
		}
	});

	return deletedUIDs;
}

// PrepareData prepares the heuristic data, so it can be sent to be executed
function prepareData(oldStateMap, newState, changeSet, deletedData) {
	const changedItems = [];
	const newStateMap = new Map(newState.map(d => [d.uid, d]));
	changeSet.forEach(changedUid => {
		changedItems.push(newStateMap.get(changedUid));
	});

	const filteredData = [];
	// Filter properties which do not need to be sent over the wire: timestamp and result count
	changedItems.forEach(d => {
		// Filter out the dummy element
		if (d.uid === rootIdentifier) {
			return;
		}

		filteredData.push({
			uid: d.uid,
			type: d.type,
			parameter: d.parameter,
			children: d.children,
			parent: d.parent,
			useAddressExclusionList: d.excludeAddresses,
			clusterTypes: d.clusterTypes,
			excludeSpendingGaps: d.excludeSpendingGaps,
		});
	});

	return {changed: filteredData, deleted: deletedData};
}

function newRouting(context) {
	const {id} = context.$route.params;
	if (id === undefined || context.$route.name !== ROUTE_NAME_HEURISTIC_PAGE) {
		return;
	}

	context.onMounted();
}

function areDataElementsEqual(a, b) {
	if (a.uid !== b.uid || a.parameter !== b.parameter || a.type !== b.type) {
		return false;
	}

	if (a.parent !== undefined && b.parent !== undefined) {
		return a.parent[0].uid === b.parent[0].uid;
	}

	return !((a.parent !== undefined && b.parent === undefined)
      || (b.parent !== undefined && a.parent === undefined));
}

export default {
	name: 'HeuristicEditorPage',
	components: {HeuristicDetailsSidebar, HeuristicTypeSelectionSideBar, NestedMenu},
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
				mdiTimerSand,
			},
			routeTransaction: ROUTE_NAME_TRANSACTION_PAGE,
			routeHeuristicOverview: ROUTE_NAME_USER_HEURISTIC_PAGE,
			isHeuristicExecuting: false,
			banner: {
				// Show is the switch for the warning banner
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
			// DbState holds the state of the database.
			// It is used to detect changes in this.data.heuristics (computed)
			dbState: null,
			// ChangeSet holds all changes based on dbState and this.data.heuristics (computed)
			changeSet: [],
			// DeletedData holds all UIDs of the heuristic which are deleted
			deletedData: [],
			isAddHeuristicSheetOpen: false,
			// HeuristicDetailsMap: map[heuristicUid]map[addressHash]array[originHash]
			heuristicDetailsMap: new Map(),
			heuristicSheet: {
				isOpen: false,
				isLoaded: false,
				heuristicUid: '',
				heuristicTypeTitle: '',
				heuristicParameter: '',
				heuristicCustomClusters: false,
				heuristicExcludeAddresses: false,
				heuristicTimestamp: null,
				clusterCount: null,
				clusters: [],
			},
			ht: new HeuristicTree(150, this),
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
					{title: 'Show Properties', icon: mdiChartBar, action: this.simulateClick},
					{isDivider: true},
					{
						title: 'Add Heuristic',
						icon: mdiShapeSquarePlus,
						action: this.openTypeSelectionSheet,
						disabled: () => !this.banner.show,
					},
					// Reminder: enable when https://github.com/vuetifyjs/vuetify/issues/17004 is fixed
					// {
					// 	title: 'Actions',
					// 	menu: [
					// 		{
					// 			title: 'Execute Heuristics',
					// 			icon: mdiSourceBranchCheck,
					// 			action: this.executeHeuristics,
					// 			disabled: this.isExecutable,
					// 		},
					// 	],
					// },
				],
			},
		};
	},
	watch: {
		$route() {
			newRouting(this);
		},
	},
	beforeUnmount() {
		// Reset memory
		this.resetExecutionStatus();
	},
	async mounted() {
		if (!await this.onMounted()) {
			return;
		}

		this.startDormantTimer();
	},
	methods: {
		simulateClick() {
			this.ht.simulateClick();
		},
		setErrorMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'error', temporary: true, category: this.$route.name});
		},
		setInfoMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'info', temporary: true, category: this.$route.name});
		},
		addNewHeuristic(heuristic) {
			const newHeuristic = {
				uid: `${this.newUidPrefix}${this.uidCounter}`,
				type: heuristic.type,
				clusterTypes: heuristic.useCustomClusters ? [CLUSTER_TYPE_CUSTOM] : [],
				excludeAddresses: heuristic.useAddressExclusionList,
				excludeSpendingGaps: heuristic.excludeSpendingGaps,
			};

			if (heuristic.parameter) {
				newHeuristic.parameter = `${heuristic.parameter.value}`;
			}

			this.uidCounter += 1;

			this.data.heuristics.push(newHeuristic);
			this.updateGraph();
		},
		async loadHeuristicDetails(uid) {
			try {
				const response = await this.dakar.heuristic.heuristicDetailsPost({heuristic: {uid}});

				if (!response.heuristic) {
					throw new Error('response contains no heuristics');
				}

				this.heuristicDetailsMap.set(response.heuristic.uid, response.heuristic);
				this.$store.dispatch('resetMessages');
			} catch (e) {
				handleError(this, e);
			}
		},
		openTypeSelectionSheet() {
			this.heuristicSheet.isOpen = false;
			this.isAddHeuristicSheetOpen = true;
		},
		async openPropertySheet(heuristic) {
			const sheet = this.heuristicSheet;

			// Lookup type title from type id
			let displayType = '';
			this.heuristicDescriptors.some(d => {
				if (d.type === heuristic.type) {
					displayType = d.title;
					return true;
				}

				return false;
			});

			// Open sheet immediately, but show skeleton loader
			this.isAddHeuristicSheetOpen = false;
			sheet.isOpen = true;
			sheet.isLoaded = false;

			sheet.heuristicParameter = heuristic.parameter;
			sheet.heuristicExcludeAddresses = heuristic.excludeAddresses;
			sheet.heuristicExcludeSpendingGaps = heuristic.excludeSpendingGaps;
			sheet.heuristicCustomClusters = heuristic.clusterTypes?.length > 0;
			sheet.heuristicTypeTitle = displayType;
			sheet.clusterCount = heuristic.clusterCount;
			sheet.heuristicUid = heuristic.uid;
			sheet.heuristicTimestamp = new Date(heuristic.ts);
			sheet.clusters = [];

			// Check if data has to be loaded from backend
			if (!heuristic.clusterCount || heuristic.uid.startsWith(this.newUidPrefix)) {
				sheet.isLoaded = true;
				return;
			}

			if (this.heuristicDetailsMap.has(heuristic.uid)) {
				sheet.clusters = this.heuristicDetailsMap.get(heuristic.uid).clusters;
				sheet.isLoaded = true;
				return;
			}

			// Request data from backend
			await this.loadHeuristicDetails(heuristic.uid);

			// Return if request was not successful
			if (this.heuristicDetailsMap.size === 0
            || !this.heuristicDetailsMap.has(heuristic.uid)) {
				return;
			}

			sheet.clusters = this.heuristicDetailsMap.get(heuristic.uid).clusters;
			sheet.isLoaded = true;
		},
		closeSideBars() {
			this.heuristicSheet.isOpen = false;
			this.isAddHeuristicSheetOpen = false;
		},
		isExecutable() {
			if (!this.banner.show && this.dbState !== null
          && ((this.changeSet !== null && this.changeSet.length > 0))) {
				return true;
			}

			return this.deletedData.length > 0;
		},
		async executeHeuristics() {
			// Prevent execution if data is not available
			if (!this.isExecutable()) {
				return;
			}

			// Close sidebars
			this.closeSideBars();

			try {
				const response = await this.dakar.heuristic.executeHeuristicsHashPost({
					hash: this.transactionHash,
					heuristic: prepareData(this.dbState, this.data.heuristics, this.changeSet, this.deletedData),
				});
				this.setExecutionStatus(response.status);
				this.startActiveTimer();
			} catch (e) {
				this.setErrorMessage(e);
			}
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
		// UpdateChangeSet updates the change set <this.changeSet> based
		// on the differences of this.data.heuristics and this.dbState
		updateChangeSet() {
			this.changeSet = [];
			const originChangeSet = [];
			this.data.heuristics.forEach(d => {
				if (this.dbState.has(d.uid)) {
					const thisElement = this.dbState.get(d.uid);
					if (!areDataElementsEqual(thisElement, d)) {
						// Changed element
						originChangeSet.push(d);
					}
				} else {
					// New element
					originChangeSet.push(d);
				}
			});

			// This set will have some duplicates, if changes are nested we get overlapping descendants
			const descendantArray = [];
			// Find descendants for each changed root element
			originChangeSet.forEach(d => {
				// Get subtree
				const descendants = this.ht.getDescendants(d.uid);
				descendantArray.push(...descendants);
			});

			// Remove duplicates
			const descendantMap = new Map(descendantArray.map(
				tempObject => [tempObject.data.data.uid, tempObject],
			));

			// Save in global changeSet
			descendantMap.forEach(e => this.changeSet.push(e.data.data.uid));

			this.ht.setNodesChanged(descendantMap);
		},
		deleteSubTree() {
			const toBeRemoved = this.ht.getRemovableNodes();
			const rel = this.ht.getRemovableRelationship();

			const updatedData = [];

			this.data.heuristics.forEach(e => {
				// Update children set of parent
				if (rel.parentUid !== '' && e.uid === rel.parentUid) {
					e.children = e.children.filter(c => c.uid !== rel.childUid);
				}

				// Remove removable nodes
				if (!toBeRemoved.includes(e.uid)) {
					updatedData.push(e);
				}
			});

			this.data = {heuristics: updatedData, status: 0};

			const newStateMap = new Map(this.data.heuristics.map(d => [d.uid, d]));
			this.deletedData = getDeletedData(this.dbState, newStateMap);

			// Update displayed graph
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
			// Maps the node data to the tree layout
			this.ht.processGraphData(this.data.heuristics);
			// UpdateChangeSet is called after a graph update,
			// because otherwise it gets an out of date descendant state
			this.updateChangeSet();
		},
		async loadHeuristicData(transactionHash) {
			try {
				this.data = await this.dakar.heuristic.heuristicsHashGet({
					hash: transactionHash,
				});
				this.$store.dispatch('resetMessages');
			} catch (e) {
				handleError(this, e);
			}
		},
		async refreshData() {
			await this.loadHeuristicData(this.transactionHash);

			if (!this.data) {
				return false;
			}

			this.setExecutionStatus(this.data.status);

			if (this.executionStatus.value.executing) {
				this.startActiveTimer();
			}

			// If the transaction has not yet any heuristics associated
			if (!this.data || !this.data.heuristics) {
				this.data.heuristics = [];
			}

			this.data.heuristics.push({uid: 'root'});

			// Reset deleted data
			this.deletedData = [];

			this.dbState = new Map(structuredClone(this.data.heuristics)
				.map(d => [d.uid, d]));
			this.updateGraph();

			return true;
		},
		async getDescriptors() {
			try {
				const response = await this.dakar.heuristic.heuristicDescriptorsGet();

				if (!response.descriptors) {
					throw Error('heuristic descriptor list is empty');
				}

				this.heuristicDescriptors = response.descriptors.map(e => {
					// Add valid property
					if (e.parameter) {
						e.parameter.valid = false;
					}

					return e;
				}).sort((a, b) => {
					if (a.title > b.title) {
						return 1;
					}

					if (a.title < b.title) {
						return -1;
					}

					return 0;
				});
			} catch (e) {
				this.setErrorMessage(e);
			}
		},
		createTabs() {
			const tabSet = new Set();
			let isCategoryEmpty = false;
			this.heuristicDescriptors.forEach(e => {
				if (e.category) {
					tabSet.add(e.category);
				} else {
					// If no category is set
					isCategoryEmpty = true;
				}
			});
			this.heuristicTabItems = Array.from(tabSet).sort().reverse();

			if (isCategoryEmpty) {
				this.heuristicTabItems.push('Other');
			}
		},
		async onMounted() {
			const svgCanvasId = 'svg_canvas';
			// Remove previous svg children
			document.getElementById(svgCanvasId).innerHTML = '';

			// Set transaction hashes for this page view
			this.transactionHash = this.$route.params.id;

			// Set page title
			document.title = `Heuristic ${this.transactionHash} - ${APPLICATION_NAME}`;

			if (!this.ht.setNodeClickHandler(this.openPropertySheet)) {
				this.setErrorMessage('error setting heuristic click handler');
				return false;
			}

			if (!this.ht.setSvgClickCallback(this.closeSideBars)) {
				this.setErrorMessage('error setting svg click handler');
				return false;
			}

			if (!this.ht.setContextMenuCallback(this.showContextMenu)) {
				this.setErrorMessage('error setting context menu handler');
				return false;
			}

			// Gets all heuristic type configurations
			await this.getDescriptors();
			if (this.heuristicDescriptors.length === 0) {
				return false;
			}

			// Creates the tab descriptions based on the heuristic categories
			this.createTabs();
			this.ht.populateHeuristicMap(this.heuristicDescriptors);
			this.ht.setupSvg(this, svgCanvasId);
			if (!await this.refreshData()) {
				return false;
			}

			await this.ht.centerGraph();
			return true;
		},
		onMenuItemClick(item) {
			if (item.action) {
				item.action();
			}

			this.contextMenu.display = false;
		},
		async updateExecutionStatus() {
			try {
				const response = await this.dakar.heuristic.heuristicStatusHashGet({hash: this.transactionHash});
				if (!response.status) {
					throw Error('execution status is not defined');
				}

				const oldExecutionStatus = this.executionStatus.value.executing;
				this.setExecutionStatus(response.status);
				// If it was previously executing refresh data
				if (oldExecutionStatus && !this.executionStatus.value.executing) {
					await this.refreshData();
					this.stopActiveTimer();
				}
			} catch (e) {
				this.setErrorMessage(e);
			}
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
};
</script>

<style scoped>

:deep( .node text ) {
  font: 12px sans-serif;
  cursor: pointer;
}

:deep( .link ) {
  fill: none;
  stroke: darkslategrey;
  stroke-width: 2px;
}

:deep( .rect ) {
  stroke: rgb(var(--v-theme-primary));
  fill: rgb(var(--v-theme-surface));
  fill-opacity: 1;
  cursor: pointer;
}

:deep( .clicked ) {
  stroke: #FDD835;
}

:deep( .modified ) {
  stroke-dasharray: 5;
}

:deep( .selected ){
  fill: #9CCC65;
  fill-opacity: 1;
}

:deep( .valid-target ) {
  stroke: #2E7D32;
  stroke-width: 4px;
}

:deep( #svg_canvas ) {
  height: 100%;
  filter: drop-shadow(-4px 4px 2px var(--v-shadow-key-penumbra-opacity, rgba(0, 0, 0, 0.2)));
}

:deep( .v-toolbar__content, .v-toolbar__extension ){
  padding-right: 0
}

.heuristicToolbar {
  border-top: 0 lightgrey solid;
  /* VAppbar has z-index 1004, therefore set z-index to the same so top shadow is not visible */
  z-index: 1004;
  filter: drop-shadow(-4px 4px 2px var(--v-shadow-key-penumbra-opacity, rgba(0, 0, 0, 0.2)));
}
</style>
