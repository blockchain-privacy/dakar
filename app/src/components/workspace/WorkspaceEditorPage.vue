<template>
  <div
    class="flex-column d-flex"
    style="height: 100%;position:relative"
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
    <div style="height: 100%; width:100%; position: relative">
      <v-card
        style="position:absolute; left: 10px; top:10px; z-index:1005; background-color: rgb(var(--v-theme-surface))"
      >
        <v-card-text class="d-flex align-center pa-0">
          <p class="mx-3 text-h6">
            XYZ
          </p>
          <v-text-field
            v-model="graphQuery"
            style="min-width:220px; max-width:300px"
            :hide-details="true"
            variant="outlined"
            density="compact"
            color="primary"
            :single-line="true"
            label="Add transactions or clusters"
            :append-inner-icon="mdiMagnify"
            @click:append-inner="handleGraphQuery(graphQuery)"
            @keydown.enter="handleGraphQuery(graphQuery)"
          />
          <v-progress-circular
            v-if="isLoading"
            :indeterminate="true"
          />
          <v-btn
            style="min-width: 32px !important;"
            class="ms-3 px-2"
            variant="text"
            :disabled="banner.show || executionStatus.executing"
            @click="modifyNode()"
          >
            <v-icon>{{ mdiShapeSquareRoundedPlus }}</v-icon>
            <div class="hidden-sm-and-down">
              Modify Node
            </div>
          </v-btn>
          <v-btn
            style="min-width: 32px !important;"
            class="ms-3 px-2"
            variant="text"
            :disabled="banner.show || executionStatus.executing"
            @click="addRandomNode()"
          >
            <v-icon>{{ mdiShapeSquareRoundedPlus }}</v-icon>
            <div class="hidden-sm-and-down">
              Add Node
            </div>
          </v-btn>
          <v-btn
            style="min-width: 32px !important;"
            class="ms-3 px-2"
            variant="text"
            :disabled="banner.show || executionStatus.executing"
            @click="hg.centerGraph()"
          >
            <v-icon>{{ mdiImageFilterCenterFocus }}</v-icon>
            <div class="hidden-sm-and-down">
              Center Graph
            </div>
          </v-btn>
          <v-btn
            style="min-width: 32px !important;"
            class="ms-3 px-2"
            variant="text"
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
            class="ms-3 px-2"
            variant="text"
            :disabled="banner.show || executionStatus.executing"
            @click="hg.setEnableInteractions(!hg.getEnableInteractions())"
          >
            <v-icon>{{ mdiShapeSquareRoundedPlus }}</v-icon>
            <div class="hidden-sm-and-down">
              Toggle graph interaction
            </div>
          </v-btn>
          <v-menu location="bottom">
            <template #activator="{ props }">
              <v-btn
                :icon="true"
                variant="text"
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
        </v-card-text>
      </v-card>
      <!-- position: relative; is needed so the dialog is contained in its parent -->
      <div style="position: relative; height: 100%; width: 100%; overflow: hidden">
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

        <context-menu
          v-model="contextMenuModel.display"
          :position-x="contextMenuModel.x"
          :position-y="contextMenuModel.y"
          :menu-items="contextMenuModel.items"
        />
        <svg id="svg_canvas" />
      </div>
    </div>
  </div>
</template>

<script setup>
import {
	mdiAlertOctagon, mdiChartBar, mdiDelete, mdiDotsVertical, mdiImageFilterCenterFocus, mdiMagnify, mdiOpenInNew,
	mdiShapeSquarePlus, mdiShapeSquareRoundedPlus,
} from '@mdi/js';
import HeuristicTypeSelectionSideBar from '../heuristic/HeuristicTypeSelectionSideBar.vue';
import {
	APPLICATION_NAME,
	CLUSTER_TYPE_CUSTOM,
	ROUTE_NAME_HEURISTIC_PAGE,
	ROUTE_NAME_TRANSACTION_PAGE,
	ROUTE_NAME_USER_HEURISTIC_PAGE,
} from '@/constants';
import ContextMenu from '../common/ContextMenu.vue';
import {getColorMap, handleError} from '@/utilities';
import HeuristicDetailsSidebar from '@/components/heuristic/HeuristicDetailsSideBar.vue';
import {onBeforeUnmount, onMounted, ref, watch, nextTick, inject} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import HeuristicGraph from '@/d3Documents/heuristicGraph';

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();
const context = {addMessage: msgStore.addMessage, $route: route};

const newUidPrefix = 'newUid_';

const colorMap = getColorMap();
colorMap.set('heuristic', 'striped');
colorMap.set('cluster', 'checkers');
// Non-privacy transaction
colorMap.set('transaction', 'grey');

const hg = new HeuristicGraph(colorMap);

let uidCounter = 1;
let data = null;
// HeuristicDetailsMap: map[heuristicUid]map[addressHash]array[originHash]
const heuristicDetailsMap = new Map();

const isLoading = ref(false);
const graphQuery = ref('');
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
const contextMenuModel = ref({
	display: false,
	x: 0,
	y: 0,
	items: [
		{title: 'Show Properties', icon: mdiChartBar, action: () => hg.contextMenuNodeClick()},
		{isDivider: true},
		{title: 'Add Heuristic', icon: mdiShapeSquarePlus, action: openTypeSelectionSheet, disabled: () => !banner.value.show},
		{title: 'Delete Node', icon: mdiDelete, action: () => hg.removeContextMenuNode(), disabled: () => !banner.value.show},
	],
});

// Watchers
watch(route, () => {
	newRouting();
});

watch(heuristicSheet.value, newVal => {
	// If sheet is being closed reset click state of graph
	if (!newVal.isOpen) {
		hg.resetClick();
	}
});

watch(isAddHeuristicSheetOpen, newVal => {
	// If sheet is being closed reset click state of graph
	if (!newVal.value) {
		hg.resetClick();
	}
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

// Functions
async function handleGraphQuery(query) {
	const trimmedQuery = query.trim();
	if (!trimmedQuery) {
		return;
	}

	graphQuery.value = '';

	hg.setEnableInteractions(false);
	isLoading.value = true;
	try {
		const response = await dakar.workspace.addWorkspaceNodePost({query: {
			query: trimmedQuery,
			currentState: hg.exportNodes(),
		}});
		console.log(structuredClone(response.nodes));
		hg.addNodes(response.nodes);
		hg.centerOnNewNodes();
	} catch (e) {
		setErrorMessage(e);
	}

	isLoading.value = false;
	hg.setEnableInteractions(true);
}

function addRandomNode() {
	const allNodes = hg.exportNodes();

	const childNode = allNodes[Math.floor(Math.random() * allNodes.length)];

	hg.addNode({uid: `${Math.random()}`, type: 'cluster', children: [childNode.uid]});
	hg.centerOnNewNodes();
}

let flag = false;
function modifyNode() {
	if (flag) {
		hg.addNode({uid: '0x100', type: 'heuristic', status: 'loading'});
	} else {
		hg.addNode({uid: '0x100', type: 'heuristic'});
	}

	flag = !flag;
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

function mockGetDBState() {
	return '[{"uid":"0x1","type":"cluster","children":["0x2","0x3"],"x":365.7393,"y":-279.538},{"uid":"0x2","type":"origin","children":["0x4","0x7"],"x":296.0357,"y":-400.9613},{"uid":"0x3","type":"destination","children":["0x6","0x5"],"x":507.9497,"y":-280.7019},{"uid":"0x4","type":"cluster","x":156.0119,"y":-401.265},{"uid":"0x5","type":"mixing","x":435.9277,"y":-400.7161},{"uid":"0x6","type":"transaction","x":437.8707,"y":-159.5429},{"uid":"0x7","type":"cluster","x":225.7315,"y":-279.86},{"uid":"0x8","type":"transaction","x":-216.1006,"y":-297.2454}]';
}

async function openPropertySheet(heuristic) {
	const sheet = heuristicSheet;

	// Lookup type title from type id
	let displayType = '';
	for (const descriptor of heuristicDescriptors.value) {
		if (descriptor.type === heuristic.type) {
			displayType = descriptor.title;
			break;
		}
	}

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

function showContextMenu(e) {
	contextMenuModel.value.display = false;

	e.preventDefault();
	contextMenuModel.value.x = e.clientX;
	contextMenuModel.value.y = e.clientY;

	nextTick(() => {
		contextMenuModel.value.display = true;
	});
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

	if (!hg.setNodeClickHandler(openPropertySheet)) {
		setErrorMessage('error setting heuristic click handler');
		return false;
	}

	if (!hg.setSvgZoomCallback(() => {
		contextMenuModel.value.display = false;
	})) {
		setErrorMessage('error setting zoom handler');
		return false;
	}

	if (!hg.setSvgClickCallback(closeSideBars)) {
		setErrorMessage('error setting svg click handler');
		return false;
	}

	if (!hg.setContextMenuCallback(showContextMenu)) {
		setErrorMessage('error setting svg context menu handler');
		return false;
	}

	// Gets all heuristic type configurations
	await getDescriptors();
	if (heuristicDescriptors.value.length === 0) {
		return false;
	}

	// Creates the tab descriptions based on the heuristic categories
	createTabs();

	hg.initSvg(svgCanvasId);

	// Const nodesFromDB = JSON.parse(mockGetDBState());

	// Const nodesFromDB = [
	// 	{uid: '0x1', type: 'cluster', children: ['0x2', '0x3']},
	// 	{uid: '0x2', type: 'transaction', children: ['0x4', '0x7']},
	// 	{uid: '0x3', type: 'transaction', children: ['0x6', '0x5']},
	// 	{uid: '0x4', type: 'cluster'},
	// 	{uid: '0x5', type: 'transaction'},
	// 	{uid: '0x6', type: 'transaction'},
	// 	{uid: '0x7', type: 'cluster'},
	// 	{uid: '0x8', type: 'transaction'},
	// ];
	// hg.addNodes(nodesFromDB);
	// hg.centerGraph();

	// Await sleep(1000);
	// hg.addNodes([
	// 	{uid: '0x10', type: 'heuristic', status: 'loading'},
	// 	{uid: '0x11', type: 'transaction', children: ['0x10']},
	// ]);
	// hg.centerOnNewNodes();

	if (!await refreshData()) {
		return false;
	}

	// Await hg.centerGraph();

	return true;
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

:deep( #svg_canvas ) {
  height: 100%;
  width: 100%;
  filter: drop-shadow(-4px 4px 2px var(--v-shadow-key-penumbra-opacity, rgba(0, 0, 0, 0.2)));
}

:deep( .v-toolbar__content, .v-toolbar__extension ){
  padding-right: 0
}

</style>
