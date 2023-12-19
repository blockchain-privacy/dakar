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
        v-if="workspaceName"
        style="position:absolute; left: 10px; top:10px; z-index:1005; background-color: rgb(var(--v-theme-surface))"
      >
        <v-card-text class="d-flex align-center pa-0">
          <p class="mx-3 text-h6">
            <v-icon icon="$graphIcon" />
            {{ workspaceName }}
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
              <v-list-item :to="{ name: ROUTE_NAME_WORKSPACES_PAGE}">
                <template #prepend>
                  <v-icon>{{ mdiOpenInNew }}</v-icon>
                </template>
                <v-list-item-title>Workspaces Overview</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-menu>
        </v-card-text>
      </v-card>
      <div
        v-if="workspaceName"
        style="position:absolute; top: 10px; right:10px"
        class="text-caption"
      >
        <template v-if="isSaving">
          Saving ...
        </template>
        <template v-else>
          <v-icon :icon="mdiCheckCircle" />
          Saved
        </template>
      </div>
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
	mdiAlertOctagon,
	mdiChartBar,
	mdiCheckCircle,
	mdiDelete,
	mdiDotsVertical,
	mdiImageFilterCenterFocus,
	mdiMagnify,
	mdiOpenInNew,
	mdiShapeSquarePlus,
	mdiShapeSquareRoundedPlus,
} from '@mdi/js';
import HeuristicTypeSelectionSideBar from '../heuristic/HeuristicTypeSelectionSideBar.vue';
import {
	APPLICATION_NAME, CLUSTER_TYPE_CUSTOM, ROUTE_NAME_WORKSPACE_PAGE, ROUTE_NAME_WORKSPACES_PAGE,
} from '@/constants';
import ContextMenu from '../common/ContextMenu.vue';
import {getColorMap, handleError} from '@/utilities';
import HeuristicDetailsSidebar from '@/components/heuristic/HeuristicDetailsSideBar.vue';
import {onMounted, ref, watch, nextTick, inject, onUnmounted} from 'vue';
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
const workspaceUID = ref('');
const workspaceName = ref('');
const workspaceModificationTime = ref();
const isSaving = ref(false);
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
		{title: 'Delete Node', icon: mdiDelete, action: removeGraphNode, disabled: () => !banner.value.show},
	],
});

const autoSave = {
	timer: null,
	wasSaved: false,
};

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
function onDocumentClose() {
	if (document.visibilityState === 'hidden') {
		if (autoSave.timer !== null) {
			clearTimeout(autoSave.timer);
			autoSave.timer = null;
			doAutoSave();
		}
	}
}

onMounted(async () => {
	await whenMounted();
	document.addEventListener('visibilitychange', onDocumentClose);
});

onUnmounted(() => {
	// Immediately save queued up auto save
	if (autoSave.timer !== null) {
		clearTimeout(autoSave.timer);
		autoSave.timer = null;
		doAutoSave();
	}

	document.removeEventListener('visibilitychange', onDocumentClose);
});

// Functions
function removeGraphNode() {
	hg.removeContextMenuNode();
	queueAutoSave();
}

async function handleGraphQuery(query) {
	if (isSaving.value) {
		msgStore.resetMessages();
		setInfoMessage('currently saving, please try again later');
		return;
	}

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
			workspaceUID: workspaceUID.value,
		}});
		if (response.nodes) {
			hg.addNodes(response.nodes);
			hg.centerOnNewNodes();
		}
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
	if (id === undefined || route.name !== ROUTE_NAME_WORKSPACE_PAGE) {
		return;
	}

	await whenMounted();
}

function setErrorMessage(msg) {
	msgStore.addMessage({text: msg, type: 'error', temporary: true, category: route.name});
}

function setInfoMessage(msg) {
	msgStore.addMessage({text: msg, type: 'info', temporary: true, category: route.name});
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

function showContextMenu(e) {
	contextMenuModel.value.display = false;

	e.preventDefault();
	contextMenuModel.value.x = e.clientX;
	contextMenuModel.value.y = e.clientY;

	nextTick(() => {
		contextMenuModel.value.display = true;
	});
}

async function refreshData() {
	isLoading.value = true;
	try {
		const response = await dakar.workspace.getWorkspaceUidGet({uid: workspaceUID.value});
		if (response.workspace) {
			data = response.workspace;
			workspaceName.value = data.name;
			workspaceModificationTime.value = new Date(data.ts);
			if (data.state) {
				data.state = JSON.parse(data.state);
			}
		} else {
			data = null;
		}

		msgStore.resetMessages();
	} catch (e) {
		handleError(context, e);
	}

	isLoading.value = false;

	if (!data) {
		return false;
	}

	// If the workspace does not yet contain any nodes
	if (!data.state) {
		data.state = [];
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

function queueAutoSave() {
	isSaving.value = true;
	if (autoSave.timer !== null) {
		clearTimeout(autoSave.timer);
	}

	autoSave.timer = setTimeout(doAutoSave, 5000);
}

async function doAutoSave() {
	isSaving.value = true;
	autoSave.timer = null;
	try {
		const response = await dakar.workspace.updateWorkspacePost({state: {
			workspaceUID: workspaceUID.value,
			currentState: hg.exportNodes(),
		}});
		if (response.ts) {
			workspaceModificationTime.value = new Date(response.ts);
			autoSave.wasSaved = true;
		}
	} catch (e) {
		setErrorMessage(e);
	}

	isSaving.value = false;
}

async function whenMounted() {
	const svgCanvasId = 'svg_canvas';
	// Remove previous svg children
	document.getElementById(svgCanvasId).innerHTML = '';

	// Set workspace UID for this page view
	workspaceUID.value = route.params.id;

	// Set page title
	document.title = `Workspace - ${APPLICATION_NAME}`;

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

	if (!hg.setDragEndCallback(queueAutoSave)) {
		setErrorMessage('error setting drag end handler');
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

	if (!await refreshData()) {
		return false;
	}

	// Update page title
	document.title = `Workspace ${workspaceName.value} - ${APPLICATION_NAME}`;

	hg.addNodes(data.state);

	await hg.centerGraph();

	return true;
}

</script>

<style scoped>

:deep( #svg_canvas ) {
  height: 100%;
  width: 100%;
  filter: drop-shadow(-4px 4px 2px var(--v-shadow-key-penumbra-opacity, rgba(0, 0, 0, 0.2)));
}

</style>
