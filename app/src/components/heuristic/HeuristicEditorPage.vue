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
          :icon="mdiAlertOctagon"
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
        <v-icon>{{ mdiTransfer }}</v-icon>
        Transaction {{ transactionHash }}
      </v-toolbar-title>

      <v-btn
        style="min-width: 32px !important;"
        class="ms-3 pa-2"
        variant="outlined"
        :disabled="banner.show || executionStatus.executing"
        @click="openTypeSelectionSheet"
      >
        <v-icon>{{ mdiShapeSquareRoundedPlus }}</v-icon>
        <div class="hidden-sm-and-down">
          Add Heuristic
        </div>
      </v-btn>
      <v-btn
        style="min-width: 32px !important;"
        class="ms-3 pa-2"
        variant="outlined"
        :disabled="banner.show || executionStatus.executing || !isExecutable"
        @click="executeHeuristics"
      >
        <v-icon>{{ mdiSourceBranchCheck }}</v-icon>
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
            <v-icon>{{ mdiDotsVertical }}</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: transactionHash }}">
            <template #prepend>
              <v-icon>{{ mdiOpenInNew }}</v-icon>
            </template>
            <v-list-item-title>Transaction Page</v-list-item-title>
          </v-list-item>
          <v-list-item :to="{ name: ROUTE_NAME_USER_HEURISTIC_PAGE}">
            <template #prepend>
              <v-icon>{{ mdiOpenInNew }}</v-icon>
            </template>
            <v-list-item-title>Heuristic Overview</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-toolbar>
    <!-- position: relative; is needed so the dialog is contained in its parent -->
    <div style="position: relative; height: 100%; width: 100%; overflow: hidden">
      <v-dialog
        :model-value="executionStatus.executing"
        :persistent="true"
        max-width="700px"
        :contained="true"
        :no-click-animation="true"
      >
        <v-card>
          <v-card-text class="text-subtitle-1 d-flex align-center">
            <v-icon
              :icon="mdiTimerSand"
              size="50"
              class="me-3"
            />
            <div>
              <p
                v-if="executionStatus.processing"
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
                :color="executionStatus.processing?'primary':''"
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

<script setup>
import {
	mdiAlertOctagon, mdiChartBar, mdiDelete, mdiDotsVertical,
	mdiOpenInNew, mdiShapeSquarePlus, mdiShapeSquareRoundedPlus,
	mdiSourceBranchCheck, mdiTransfer, mdiTimerSand,
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
import {onBeforeUnmount, onMounted, ref, watch, nextTick, inject, computed} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();
const context = {addMessage: msgStore.addMessage, $route: route};

const newUidPrefix = 'newUid_';
const ht = new HeuristicTree(150);
let uidCounter = 1;
let data = null;
// HeuristicDetailsMap: map[heuristicUid]map[addressHash]array[originHash]
const heuristicDetailsMap = new Map();

// DeletedData holds all UIDs of the heuristic which are deleted
const deletedData = ref([]);
// ChangeSet holds all changes based on dbState and data.heuristics (computed)
const changeSet = ref([]);
// DbState holds the state of the database.
// It is used to detect changes in data.heuristics (computed)
const dbState = ref(null);
const transactionHash = ref('');
const isAddHeuristicSheetOpen = ref(false);
const heuristicDescriptors = ref([]);
const heuristicTabItems = ref([]);
const banner = ref({
	// Show is the switch for the warning banner
	// which gets displayed if the heuristic worker is not ready to accept requests
	show: false,
	display: true,
});
const executionStatus = ref({
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
});
const heuristicSheet = ref({
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
});
const contextMenu = ref({
	display: false,
	x: 0,
	y: 0,
	items: [
		{
			title: 'Delete Heuristic',
			icon: mdiDelete,
			action: deleteSubTree,
			disabled: () => !banner.value.show,
		},
		{title: 'Show Properties', icon: mdiChartBar, action: simulateClick},
		{isDivider: true},
		{
			title: 'Add Heuristic',
			icon: mdiShapeSquarePlus,
			action: openTypeSelectionSheet,
			disabled: () => !banner.value.show,
		},
		// Reminder: enable when https://github.com/vuetifyjs/vuetify/issues/17004 is fixed
		// {
		// 	title: 'Actions',
		// 	menu: [
		// 		{
		// 			title: 'Execute Heuristics',
		// 			icon: mdiSourceBranchCheck,
		// 			action: executeHeuristics,
		// 			disabled: isExecutable,
		// 		},
		// 	],
		// },
	],
});

// Watchers
watch(route, () => {
	newRouting();
});

// Hooks
onBeforeUnmount(() => {
	// Reset memory
	resetExecutionStatus();
});

onMounted(async () => {
	if (!await whenMounted()) {
		return;
	}

	startDormantTimer();
});

// Computed
const isExecutable = computed(() => {
	if (!banner.value.show && dbState.value !== null && changeSet.value !== null && changeSet.value.length > 0) {
		return true;
	}

	return deletedData.value.length > 0;
});

// Functions
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

async function newRouting() {
	const {id} = route.params;
	if (id === undefined || route.name !== ROUTE_NAME_HEURISTIC_PAGE) {
		return;
	}

	if (!await whenMounted()) {
		return;
	}

	startDormantTimer();
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

function simulateClick() {
	ht.simulateClick();
}

function setErrorMessage(msg) {
	msgStore.addMessage({text: msg, type: 'error', temporary: true, category: route.name});
}

function addNewHeuristic(heuristic) {
	const newHeuristic = {
		uid: `${newUidPrefix}${uidCounter}`,
		type: heuristic.type,
		clusterTypes: heuristic.useCustomClusters ? [CLUSTER_TYPE_CUSTOM] : [],
		excludeAddresses: heuristic.useAddressExclusionList,
		excludeSpendingGaps: heuristic.excludeSpendingGaps,
	};

	if (heuristic.parameter) {
		newHeuristic.parameter = `${heuristic.parameter.value}`;
	}

	uidCounter += 1;

	data.heuristics.push(newHeuristic);
	updateGraph();
}

async function loadHeuristicDetails(uid) {
	try {
		const response = await dakar.heuristic.heuristicDetailsPost({heuristic: {uid}});

		if (!response.heuristic) {
			throw new Error('response contains no heuristics');
		}

		heuristicDetailsMap.set(response.heuristic.uid, response.heuristic);
		msgStore.resetMessages();
	} catch (e) {
		handleError(context, e);
	}
}

function openTypeSelectionSheet() {
	heuristicSheet.value.isOpen = false;
	isAddHeuristicSheetOpen.value = true;
}

async function openPropertySheet(heuristic) {
	const sheet = heuristicSheet;

	// Lookup type title from type id
	let displayType = '';
	heuristicDescriptors.value.some(d => {
		if (d.type === heuristic.type) {
			displayType = d.title;
			return true;
		}

		return false;
	});

	// Open sheet immediately, but show skeleton loader
	isAddHeuristicSheetOpen.value = false;
	sheet.value.isOpen = true;
	sheet.value.isLoaded = false;

	sheet.value.heuristicParameter = heuristic.parameter;
	sheet.value.heuristicExcludeAddresses = heuristic.excludeAddresses;
	sheet.value.heuristicExcludeSpendingGaps = heuristic.excludeSpendingGaps;
	sheet.value.heuristicCustomClusters = heuristic.clusterTypes?.length > 0;
	sheet.value.heuristicTypeTitle = displayType;
	sheet.value.clusterCount = heuristic.clusterCount;
	sheet.value.heuristicUid = heuristic.uid;
	sheet.value.heuristicTimestamp = new Date(heuristic.ts);
	sheet.value.clusters = [];

	// Check if data has to be loaded from backend
	if (!heuristic.clusterCount || heuristic.uid.startsWith(newUidPrefix)) {
		sheet.value.isLoaded = true;
		return;
	}

	if (heuristicDetailsMap.has(heuristic.uid)) {
		sheet.value.clusters = heuristicDetailsMap.get(heuristic.uid).clusters;
		sheet.value.isLoaded = true;
		return;
	}

	// Request data from backend
	await loadHeuristicDetails(heuristic.uid);

	// Return if request was not successful
	if (heuristicDetailsMap.size === 0
        || !heuristicDetailsMap.has(heuristic.uid)) {
		return;
	}

	sheet.value.clusters = heuristicDetailsMap.get(heuristic.uid).clusters;
	sheet.value.isLoaded = true;
}

function closeSideBars() {
	heuristicSheet.value.isOpen = false;
	isAddHeuristicSheetOpen.value = false;
}

async function executeHeuristics() {
	// Prevent execution if data is not available
	if (!isExecutable.value) {
		return;
	}

	// Close sidebars
	closeSideBars();

	try {
		const response = await dakar.heuristic.executeHeuristicsHashPost({
			hash: transactionHash.value,
			heuristic: prepareData(dbState.value, data.heuristics, changeSet.value, deletedData.value),
		});
		setExecutionStatus(response.status);
		startActiveTimer();
	} catch (e) {
		setErrorMessage(e);
	}
}

function setExecutionStatus(status) {
	switch (status) {
		case executionStatus.value.enum.added:
			executionStatus.value.processing = false;
			executionStatus.value.executing = true;
			banner.value.show = false;
			break;
		case executionStatus.value.enum.inQueue:
			executionStatus.value.processing = false;
			executionStatus.value.executing = true;
			banner.value.show = false;
			break;
		case executionStatus.value.enum.processing:
			executionStatus.value.processing = true;
			executionStatus.value.executing = true;
			banner.value.show = false;
			break;
		case executionStatus.value.enum.notReady:
			executionStatus.value.processing = false;
			executionStatus.value.executing = false;
			banner.value.show = true;
			break;
		default:
			executionStatus.value.processing = false;
			executionStatus.value.executing = false;
			banner.value.show = false;
	}
}

// UpdateChangeSet updates the change set <changeSet> based
// on the differences of data.heuristics and dbState
function updateChangeSet() {
	changeSet.value = [];
	const originChangeSet = [];
	data.heuristics.forEach(d => {
		if (dbState.value.has(d.uid)) {
			const thisElement = dbState.value.get(d.uid);
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
		const descendants = ht.getDescendants(d.uid);
		descendantArray.push(...descendants);
	});

	// Remove duplicates
	const descendantMap = new Map(descendantArray.map(
		tempObject => [tempObject.data.data.uid, tempObject],
	));

	// Save in global changeSet
	descendantMap.forEach(e => changeSet.value.push(e.data.data.uid));

	ht.setNodesChanged(descendantMap);
}

function deleteSubTree() {
	const toBeRemoved = ht.getRemovableNodes();
	const rel = ht.getRemovableRelationship();

	const updatedData = [];

	data.heuristics.forEach(e => {
		// Update children set of parent
		if (rel.parentUid !== '' && e.uid === rel.parentUid) {
			e.children = e.children.filter(c => c.uid !== rel.childUid);
		}

		// Remove removable nodes
		if (!toBeRemoved.includes(e.uid)) {
			updatedData.push(e);
		}
	});

	data = {heuristics: updatedData, status: 0};

	const newStateMap = new Map(data.heuristics.map(d => [d.uid, d]));
	deletedData.value = getDeletedData(dbState.value, newStateMap);

	// Update displayed graph
	updateGraph();
}

function showContextMenu(e) {
	contextMenu.value.display = false;

	e.preventDefault();
	contextMenu.value.x = e.clientX;
	contextMenu.value.y = e.clientY;

	nextTick(() => {
		contextMenu.value.display = true;
	});
}

function updateGraph() {
	// Maps the node data to the tree layout
	ht.processGraphData(data.heuristics);
	// UpdateChangeSet is called after a graph update,
	// because otherwise it gets an out of date descendant state
	updateChangeSet();
}

async function loadHeuristicData() {
	try {
		data = await dakar.heuristic.heuristicsHashGet({hash: transactionHash.value});
		msgStore.resetMessages();
	} catch (e) {
		handleError(context, e);
	}
}

async function refreshData() {
	await loadHeuristicData();

	if (!data) {
		return false;
	}

	setExecutionStatus(data.status);

	if (executionStatus.value.executing) {
		startActiveTimer();
	}

	// If the transaction has not yet any heuristics associated
	if (!data || !data.heuristics) {
		data.heuristics = [];
	}

	data.heuristics.push({uid: 'root'});

	// Reset deleted data
	deletedData.value = [];

	dbState.value = new Map(structuredClone(data.heuristics)
		.map(d => [d.uid, d]));
	updateGraph();

	return true;
}

async function getDescriptors() {
	try {
		const response = await dakar.heuristic.heuristicDescriptorsGet();

		if (!response.descriptors) {
			throw Error('heuristic descriptor list is empty');
		}

		heuristicDescriptors.value = response.descriptors.map(e => {
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
		setErrorMessage(e);
	}
}

function createTabs() {
	const tabSet = new Set();
	let isCategoryEmpty = false;
	heuristicDescriptors.value.forEach(e => {
		if (e.category) {
			tabSet.add(e.category);
		} else {
			// If no category is set
			isCategoryEmpty = true;
		}
	});
	heuristicTabItems.value = Array.from(tabSet).sort().reverse();

	if (isCategoryEmpty) {
		heuristicTabItems.value.push('Other');
	}
}

async function whenMounted() {
	const svgCanvasId = 'svg_canvas';
	// Remove previous svg children
	document.getElementById(svgCanvasId).innerHTML = '';

	// Set transaction hashes for this page view
	transactionHash.value = route.params.id;

	// Set page title
	document.title = `Heuristic ${transactionHash.value} - ${APPLICATION_NAME}`;

	if (!ht.setNodeClickHandler(openPropertySheet)) {
		setErrorMessage('error setting heuristic click handler');
		return false;
	}

	if (!ht.setSvgZoomCallback(() => {
		contextMenu.value.display = false;
	})) {
		setErrorMessage('error setting zoom handler');
		return false;
	}

	if (!ht.setSvgClickCallback(closeSideBars)) {
		setErrorMessage('error setting svg click handler');
		return false;
	}

	if (!ht.setContextMenuCallback(showContextMenu)) {
		setErrorMessage('error setting context menu handler');
		return false;
	}

	if (!ht.setDragEndCallback(() => updateGraph())) {
		setErrorMessage('error setting context drag end handler');
		return false;
	}

	if (!ht.setSubTreeMoveCallback(moveSubTree)) {
		setErrorMessage('error setting sub tree move handler');
		return false;
	}

	// Gets all heuristic type configurations
	await getDescriptors();
	if (heuristicDescriptors.value.length === 0) {
		return false;
	}

	// Creates the tab descriptions based on the heuristic categories
	createTabs();
	ht.populateHeuristicMap(heuristicDescriptors.value);
	ht.setupSvg(svgCanvasId);
	if (!await refreshData()) {
		return false;
	}

	await ht.centerGraph();
	return true;
}

function onMenuItemClick(item) {
	if (item.action) {
		item.action();
	}

	contextMenu.value.display = false;
}

function moveSubTree(childUID, parentUID, formerParentUID) {
	if (!data || !data.heuristics) {
		return;
	}

	const newData = data.heuristics;

	for (let i = 0; i < newData.length; i += 1) {
		const dataElement = newData[i];
		if (dataElement.uid === parentUID) {
			if (dataElement.children === undefined) {
				dataElement.children = [];
			}

			let alreadyExists = false;
			dataElement.children.forEach(c => {
				if (c.uid === childUID) {
					alreadyExists = true;
				}
			});

			if (!alreadyExists) {
				dataElement.children.push({uid: childUID});
			}
		} else if (dataElement.uid === childUID) {
			if (dataElement.parent === undefined) {
				dataElement.parent = [];
			}

			dataElement.parent = [];
			dataElement.parent.push({uid: parentUID});
		} else if (dataElement.uid === formerParentUID) {
			dataElement.children = dataElement.children.filter(c => c.uid !== childUID);
		}
	}

	// Set new state
	data.heuristics = newData;
}

async function updateExecutionStatus() {
	try {
		const response = await dakar.heuristic.heuristicStatusHashGet({hash: transactionHash.value});
		if (!response.status) {
			throw Error('execution status is not defined');
		}

		const oldExecutionStatus = executionStatus.value.executing;
		setExecutionStatus(response.status);
		// If it was previously executing refresh data
		if (oldExecutionStatus && !executionStatus.value.executing) {
			await refreshData();
			stopActiveTimer();
		}
	} catch (e) {
		setErrorMessage(e);
	}
}

function startDormantTimer() {
	executionStatus.value.dormantTimer.timer = setInterval(async () => {
		await updateExecutionStatus();
	}, executionStatus.value.dormantTimer.refreshRate);
}

function startActiveTimer() {
	executionStatus.value.activeTimer.timer = setInterval(async () => {
		await updateExecutionStatus();
	}, executionStatus.value.activeTimer.refreshRate);
}

function stopDormantTimer() {
	clearInterval(executionStatus.value.dormantTimer.timer);
}

function stopActiveTimer() {
	clearInterval(executionStatus.value.activeTimer.timer);
}

function resetExecutionStatus() {
	stopDormantTimer();
	stopActiveTimer();
}

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
