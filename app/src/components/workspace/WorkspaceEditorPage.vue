<template>
  <div
    class="flex-column d-flex"
    style="height: 100%;position:relative"
  >
    <div style="height: 100%; width:100%; position: relative">
      <v-card
        v-if="workspaceName"
        class="workspace-toolbar"
      >
        <v-card-text class="d-flex align-center pa-0">
          <v-icon
            class="mx-3"
            icon="$graphIcon"
            size="32"
          />
          <p class="me-3 text-h6 workspace-name hidden-sm-and-down">
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
            :disabled="isModifyingWorkspace"
            :append-inner-icon="mdiMagnify"
            @click:append-inner="handleGraphQuery(graphQuery)"
            @keydown.enter="handleGraphQuery(graphQuery)"
          />
          <adaptive-menu :items="menuItems" />
        </v-card-text>
        <v-progress-linear
          v-if="isModifyingWorkspace"
          :indeterminate="true"
          :rounded="true"
          location="bottom"
        />
      </v-card>
      <div
        v-if="workspaceName && wasAutoSaved"
        style=""
        :class="{'text-caption':true, 'auto-save-small-screen': $vuetify.display.smAndDown, 'auto-save-large-screen': $vuetify.display.mdAndUp }"
      >
        <template v-if="isAutoSaving">
          Saving ...
        </template>
        <template v-else>
          <v-icon :icon="mdiCheckCircle" />
          Saved
        </template>
      </div>
      <!-- position: relative; is needed so the dialog is contained in its parent -->
      <div style="position: relative; height: 100%; width: 100%; overflow: hidden">
        <v-dialog
          :model-value="isLoadingWorkspace"
          :persistent="true"
          max-width="350px"
          :contained="true"
          :no-click-animation="true"
        >
          <v-card>
            <v-card-text class="text-subtitle-1 d-flex align-center">
              <div style="width:100%">
                <p class="text-center mb-3">
                  Loading workspace ...
                </p>
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
        <heuristic-type-selection-side-bar
          v-model="isAddHeuristicSheetOpen"
          :tab-items="heuristicTabItems"
          :descriptors="heuristicDescriptors"
          @add-heuristic="addNewHeuristic"
        />
        <entity-side-bar
          v-model="isEntitySideBarOpen"
          :identifier="entityIdentifier"
          :workspace-uid="workspaceUID"
          :auxiliary-data="entityAuxiliaryData"
          :type="entityType"
          :disable-adding-nodes="isModifyingWorkspace"
          @add-heuristic="openTypeSelectionSheet"
          @add-node="handleGraphQuery"
          @delete-entity="removeGraphNode"
        />
        <v-menu
          v-model="contextMenuModel.display"
          :open-on-hover="false"
          transition="fade-transition"
          :target="[contextMenuModel.x,contextMenuModel.y]"
        >
          <v-list>
            <template
              v-for="(item, index) in contextMenuModel.items"
              :key="index"
            >
              <v-divider v-if="item.isDivider" />
              <v-list-item
                v-else
                :key="index"
                :disabled="item.disabled && item.disabled()"
                @click="item.action(item)"
              >
                <template
                  v-if="item.icon"
                  #prepend
                >
                  <v-icon :icon="item.icon" />
                </template>
                <v-list-item-title>{{ item.title }}</v-list-item-title>
              </v-list-item>
            </template>
          </v-list>
        </v-menu>
        <svg id="svg_canvas" />
      </div>
    </div>
  </div>
</template>

<script setup>
import {
	mdiCached,
	mdiCheckCircle,
	mdiDelete,
	mdiImageFilterCenterFocus,
	mdiMagnify,
	mdiOpenInNew,
	mdiShapeCirclePlus,
} from '@mdi/js';
import HeuristicTypeSelectionSideBar from './HeuristicTypeSelectionSideBar.vue';
import {
	APPLICATION_NAME,
	CLUSTER_TYPE_CUSTOM,
	ROUTE_NAME_WORKSPACE_PAGE,
	ROUTE_NAME_WORKSPACES_PAGE,
	WORKSPACE_NODE_TYPE_CLUSTER,
	WORKSPACE_NODE_TYPE_HEURISTIC,
	WORKSPACE_NODE_TYPE_TRANSACTION,
} from '@/constants';
import {getColorMap, handleError, isDestination} from '@/utilities';
import {
	inject, nextTick, onMounted, onUnmounted, ref, watch,
} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import NodeGraph from '@/d3Documents/nodeGraph';
import {sleep} from '@/d3Documents/util';
import EntitySideBar from '@/components/workspace/EntitySideBar.vue';
import AdaptiveMenu from '@/components/workspace/AdaptiveMenu.vue';

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();
const context = {addMessage: msgStore.addMessage, $route: route};

const newUidPrefix = 'newUid_';

const colorMap = getColorMap();
colorMap.set(WORKSPACE_NODE_TYPE_HEURISTIC, '#4CAF50');
colorMap.set(WORKSPACE_NODE_TYPE_CLUSTER, '#CDDC39');
// Non-privacy transaction
colorMap.set(WORKSPACE_NODE_TYPE_TRANSACTION, '#607D8B');

const nodeGraph = new NodeGraph(colorMap);

let uidCounter = 1;
let data = null;
// Node which triggered the heuristic type selection,
// and to which will be the parent of the new heuristic.
// It may be a destination transaction or another heuristic
let newHeuristicParentNodeUID = '';
// Holds the references of all timers
const heuristicTimers = [];

const isAutoSaving = ref(false);
const wasAutoSaved = ref(false);
const isLoadingWorkspace = ref(false);
const isModifyingWorkspace = ref(false);
const graphQuery = ref('');
const workspaceUID = ref('');
const workspaceName = ref('');
const isAddHeuristicSheetOpen = ref(false);
const isEntitySideBarOpen = ref(false);
const entityIdentifier = ref('');
const entityAuxiliaryData = ref(null);
const entityType = ref('');
const heuristicDescriptors = ref([]);
const heuristicTabItems = ref([]);
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

const showContextMenuAddHeuristic = ref(false);
const contextMenuModel = ref({
	display: false,
	x: 0,
	y: 0,
	items: [
		{
			title: 'Add Heuristic',
			icon: mdiShapeCirclePlus,
			action: () => contextMenuOpenTypeSelection(nodeGraph.getContextNode()),
			disabled: () => !showContextMenuAddHeuristic.value || nodeGraph.getContextNode()?.loading,
		},
		{
			title: 'Delete',
			icon: mdiDelete,
			action: removeGraphNode,
			disabled: () => !isDeleteEnabled(nodeGraph.getContextNode()),
		},
	],
});

const menuItems = [
	{
		title: 'Reorder Nodes', icon: mdiCached, action() {
			nodeGraph.reorderNodes();
			queueAutoSave();
		},
	},
	{title: 'Center Graph', icon: mdiImageFilterCenterFocus, action: () => nodeGraph.centerGraph()},
	{title: 'Workspaces Overview', icon: mdiOpenInNew, to: {name: ROUTE_NAME_WORKSPACES_PAGE}},
];

let autoSaveTimer = null;

// Watchers
watch(route, () => {
	newRouting();
});

watch(isAddHeuristicSheetOpen, newVal => {
	// If sheet is being closed reset click state of graph
	if (!newVal) {
		nodeGraph.resetClick();
	}
});

watch(isEntitySideBarOpen, newVal => {
	// If sheet is being closed reset click state of graph
	if (!newVal) {
		nodeGraph.resetClick();
	}
});

// Hooks
function onDocumentClose() {
	if (document.visibilityState === 'hidden') {
		if (autoSaveTimer !== null) {
			clearTimeout(autoSaveTimer);
			autoSaveTimer = null;
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
	if (autoSaveTimer !== null) {
		clearTimeout(autoSaveTimer);
		autoSaveTimer = null;
		doAutoSave();
	}

	// Stop all heuristic timers
	heuristicTimers.forEach(d => clearTimeout(d));

	document.removeEventListener('visibilitychange', onDocumentClose);
});

// Functions
async function removeGraphNode() {
	const node = nodeGraph.getContextNode();
	if (!node || node.loading) {
		return;
	}

	try {
		const response = await dakar.workspace.workspacesNodeDelete({
			state: {
				nodeUID: node.uid,
				workspaceUID: workspaceUID.value,
			},
		});

		nodeGraph.removeNodes(response.deletedNodeUIDs);
	} catch (e) {
		setErrorMessage(e);
	}
}

// Getlock prevents further actions causing an autosave event to occur,
// and waits until the current autosave event is done.
async function lockAutosave() {
	nodeGraph.setEnableInteractions(false);
	isModifyingWorkspace.value = true;

	// Wait for auto save to finish
	while (isAutoSaving.value) {
		// eslint-disable-next-line no-await-in-loop
		await sleep(200);
	}
}

// Checks if a node can be deleted. If a heuristic or a node
// in a heuristic sub graph is loading it return false.
function isDeleteEnabled(contextNode) {
	if (!contextNode || contextNode.loading) {
		return false;
	}

	if (contextNode.type !== WORKSPACE_NODE_TYPE_HEURISTIC && !isDestination(contextNode.privacyType)) {
		return true;
	}

	if (contextNode.children) {
		for (const child of contextNode.children) {
			const childNode = nodeGraph.getNode(child);
			if (!childNode || childNode.type !== WORKSPACE_NODE_TYPE_HEURISTIC) {
				continue;
			}

			if (!isDeleteEnabled(childNode)) {
				return false;
			}
		}
	}

	return true;
}

function releaseAutosaveLock() {
	isModifyingWorkspace.value = false;
	nodeGraph.setEnableInteractions(true);
}

async function handleGraphQuery(query) {
	if (isModifyingWorkspace.value) {
		return;
	}

	const trimmedQuery = query.trim();
	if (!trimmedQuery) {
		return;
	}

	graphQuery.value = '';

	await lockAutosave();

	try {
		const response = await dakar.workspace.workspacesNodePost({
			query: {
				query: trimmedQuery,
				workspaceUID: workspaceUID.value,
			},
		});
		if (response.nodes) {
			nodeGraph.removeAllNodes(false);
			nodeGraph.addNodes(response.nodes);
			queueAutoSave();
			nodeGraph.centerOnNewNodes();
		}
	} catch (e) {
		setErrorMessage(e);
	}

	releaseAutosaveLock();
}

async function newRouting() {
	const {id} = route.params;
	if (id === undefined || route.name !== ROUTE_NAME_WORKSPACE_PAGE) {
		return;
	}

	await whenMounted();
}

function setErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: true, category: route.name,
	});
}

async function addNewHeuristic(heuristic) {
	await lockAutosave();

	const newHeuristic = {
		uid: `${newUidPrefix}${uidCounter}`,
		type: heuristic.type,
		clusterTypes: heuristic.useCustomClusters ? [CLUSTER_TYPE_CUSTOM] : [],
		useAddressExclusionList: heuristic.useAddressExclusionList,
		excludeSpendingGaps: heuristic.excludeSpendingGaps,
	};

	if (heuristic.parameter) {
		newHeuristic.parameter = `${heuristic.parameter.value}`;
	}

	uidCounter += 1;

	const parentNode = nodeGraph.getNode(newHeuristicParentNodeUID);
	if (!parentNode) {
		releaseAutosaveLock();
		return;
	}

	const nodes = nodeGraph.getNodes();
	const txHash = getHeuristicTransaction(nodes, newHeuristicParentNodeUID);
	if (!txHash) {
		releaseAutosaveLock();
		return;
	}

	newHeuristic.transactionHash = txHash;

	if (parentNode.type === WORKSPACE_NODE_TYPE_HEURISTIC) {
		// Only set parent if the direct parent is a heuristic
		newHeuristic.parentUID = newHeuristicParentNodeUID;
	}

	try {
		const response = await dakar.heuristic.executeHeuristicsPost({
			heuristic: {newHeuristic, workspaceUID: workspaceUID.value},
		});

		if (parentNode.children) {
			parentNode.children.push(response.workID);
		} else {
			parentNode.children = [response.workID];
		}

		nodeGraph.addNodes([parentNode, {
			uid: response.workID,
			type: WORKSPACE_NODE_TYPE_HEURISTIC,
			loading: true,
			heuristicType: newHeuristic.type,
			heuristicExcludeAddresses: newHeuristic.excludeAddresses,
			heuristicExcludeSpendingGaps: newHeuristic.excludeSpendingGaps,
			heuristicClusterTypes: newHeuristic.clusterTypes,
			heuristicParameter: newHeuristic.parameter,
		}]);
		addWork(response.workID);
	} catch (e) {
		setErrorMessage(e);
	}

	releaseAutosaveLock();
}

function addWork(workID) {
	heuristicTimers.push(setTimeout(checkWork, 3000, workID));
}

// Periodically check
async function checkWork(workID) {
	try {
		const response = await dakar.heuristic.heuristicByWorkIDPost({
			work: {
				workspaceUID: workspaceUID.value,
				id: workID,
			},
		});
		if (response.nodes) {
			nodeGraph.removeAllNodes(false);
			nodeGraph.addNodes(response.nodes);

			// ReplaceTemporaryHeuristic(workID, response.heuristic);
		} else {
			addWork(workID);
		}
	} catch (e) {
		setErrorMessage(e);
	}
}

function replaceTemporaryHeuristic(workID, heuristic) {
	// If temporary node exists in graph, collect coordinates
	const n = nodeGraph.getNode(workID);
	if (n) {
		heuristic.x = n.x;
		heuristic.y = n.y;
	}

	// If parent exists, set connection to new heuristic
	const parent = nodeGraph.getParent(n.uid);
	if (!parent) {
		// Heuristic without a parent can not exist
		return;
	}

	const childPos = parent.children.indexOf(workID);
	if (childPos === -1) {
		// Temporary node not found in children -> create new connection
		parent.children.push(heuristic.uid);
	} else {
		// Temporary node found in children -> replace connection
		parent.children[childPos] = heuristic.uid;
	}

	nodeGraph.removeNode(workID, false);
	nodeGraph.addNodes([parent, {
		uid: heuristic.uid,
		type: WORKSPACE_NODE_TYPE_HEURISTIC,
		heuristicType: heuristic.type,
		heuristicExcludeAddresses: heuristic.excludeAddresses,
		heuristicExcludeSpendingGaps: heuristic.excludeSpendingGaps,
		heuristicClusterTypes: heuristic.clusterTypes,
		heuristicParameter: heuristic.parameter,
		heuristicClusterCount: heuristic.clusterCount,
		x: heuristic.x,
		y: heuristic.y,
	}]);
}

// Returns the transaction hash of the given heuristic
function getHeuristicTransaction(nodes, uid) {
	const node = nodeGraph.getNode(uid);
	if (!node) {
		return '';
	}

	if (node.type === WORKSPACE_NODE_TYPE_TRANSACTION) {
		// Found it
		return node.transactionHash;
	}

	if (node.type === WORKSPACE_NODE_TYPE_HEURISTIC) {
		// Find parent and do recursive call
		const parent = nodes.find(v => v.children?.includes(uid));

		// Parent not found -> something went wrong
		if (!parent) {
			return '';
		}

		return getHeuristicTransaction(nodes, parent.uid);
	}

	return '';
}

function contextMenuOpenTypeSelection(node) {
	if (!node) {
		return;
	}

	newHeuristicParentNodeUID = node.uid;

	openTypeSelectionSheet();
}

function openTypeSelectionSheet() {
	isEntitySideBarOpen.value = false;
	isAddHeuristicSheetOpen.value = true;
	// Next tick so watcher actions are executed first
	nextTick(() => nodeGraph.setContextNodeClicked());
}

function openEntitySideBar(nodeData) {
	// Do not show sidebar for heuristic placeholder
	if (nodeData.uid.startsWith(newUidPrefix)) {
		return;
	}

	newHeuristicParentNodeUID = nodeData.uid;
	entityAuxiliaryData.value = null;
	entityType.value = nodeData.type;

	switch (entityType.value) {
		case WORKSPACE_NODE_TYPE_CLUSTER:
			entityIdentifier.value = nodeData.addressHash;
			break;
		case WORKSPACE_NODE_TYPE_TRANSACTION:
			entityIdentifier.value = nodeData.transactionHash;
			break;
		case WORKSPACE_NODE_TYPE_HEURISTIC:
			// Brackets so local variables stay local (more info: https://eslint.org/docs/latest/rules/no-case-declarations)
			{
				let displayType = '';
				for (const descriptor of heuristicDescriptors.value) {
					if (descriptor.type === nodeData.heuristicType) {
						displayType = descriptor.title;
						break;
					}
				}

				entityAuxiliaryData.value = nodeData;
				entityAuxiliaryData.value.displayType = displayType;
				entityIdentifier.value = nodeData.uid;
			}

			break;
		default:
	}

	isAddHeuristicSheetOpen.value = false;
	isEntitySideBarOpen.value = true;
	// Next tick so watcher actions are executed first
	nextTick(() => nodeGraph.setContextNodeClicked());
}

function closeSideBars() {
	isAddHeuristicSheetOpen.value = false;
	isEntitySideBarOpen.value = false;
}

function showContextMenu(e, nodeData) {
	contextMenuModel.value.display = false;

	e.preventDefault();

	if (nodeData?.type === WORKSPACE_NODE_TYPE_HEURISTIC || isDestination(nodeData.privacyType)) {
		showContextMenuAddHeuristic.value = true;
	} else {
		showContextMenuAddHeuristic.value = false;
	}

	contextMenuModel.value.x = e.clientX;
	contextMenuModel.value.y = e.clientY;

	nextTick(() => {
		contextMenuModel.value.display = true;
	});
}

async function refreshData() {
	isLoadingWorkspace.value = true;

	try {
		const response = await dakar.workspace.workspacesUidGet({uid: workspaceUID.value});
		if (!response.descriptors) {
			throw Error('heuristic descriptor list is empty');
		}

		if (response.workspace) {
			data = response.workspace;
			workspaceName.value = data.name;
		} else {
			data = null;
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

		msgStore.resetMessages();
	} catch (e) {
		handleError(context, e);
	}

	isLoadingWorkspace.value = false;

	if (!data) {
		return false;
	}

	// If the workspace does not yet contain any nodes, set an empty array
	data.nodes ??= [];

	return true;
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

function queueAutoSave(t = 5000) {
	isAutoSaving.value = true;
	wasAutoSaved.value = true;
	if (autoSaveTimer !== null) {
		clearTimeout(autoSaveTimer);
	}

	autoSaveTimer = setTimeout(doAutoSave, t);
}

async function doAutoSave() {
	isAutoSaving.value = true;
	autoSaveTimer = null;
	try {
		await dakar.workspace.workspacesPut({
			state: {
				workspaceUID: workspaceUID.value,
				currentState: nodeGraph.exportNodes(),
			},
		});
	} catch (e) {
		setErrorMessage(e);
	}

	isAutoSaving.value = false;
}

async function whenMounted() {
	const svgCanvasId = 'svg_canvas';
	// Remove previous svg children
	document.getElementById(svgCanvasId).innerHTML = '';

	// Set workspace UID for this page view
	workspaceUID.value = route.params.id;

	// Set page title
	document.title = `Workspace - ${APPLICATION_NAME}`;

	if (!nodeGraph.setNodeClickHandler(openEntitySideBar)) {
		setErrorMessage('error setting heuristic click handler');
		return false;
	}

	if (!nodeGraph.setSvgZoomCallback(() => {
		contextMenuModel.value.display = false;
	})) {
		setErrorMessage('error setting zoom handler');
		return false;
	}

	if (!nodeGraph.setSvgClickCallback(closeSideBars)) {
		setErrorMessage('error setting svg click handler');
		return false;
	}

	if (!nodeGraph.setContextMenuCallback(showContextMenu)) {
		setErrorMessage('error setting svg context menu handler');
		return false;
	}

	if (!nodeGraph.setDragEndCallback(queueAutoSave)) {
		setErrorMessage('error setting drag end handler');
		return false;
	}

	nodeGraph.initSvg(svgCanvasId);
	if (!await refreshData()) {
		return false;
	}

	// Creates the tab descriptions based on the heuristic categories
	createTabs();
	nodeGraph.populateHeuristicMap(heuristicDescriptors.value);

	// Update page title
	document.title = `${workspaceName.value} - Workspace - ${APPLICATION_NAME}`;

	nodeGraph.addNodes(data.nodes);

	await nodeGraph.centerGraph();

	// Add for each dummy heuristics a timer which checks if their heuristic is done executing
	for (const node of data.nodes) {
		if (node.loading) {
			addWork(node.uid);
		}
	}

	return true;
}

</script>

<style scoped>

:deep( #svg_canvas ) {
  height: 100%;
  width: 100%;
}

.workspace-name {
  max-width: 275px;
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}

.auto-save-large-screen {
  position: absolute;
  top: 10px;
  right: 10px;
  z-index: 1004;
}

.auto-save-small-screen {
  position: absolute;
  top: 65px;
  left: 10px;
  z-index: 1004;
}

.workspace-toolbar {
  position: absolute;
  left: 10px;
  top: 10px;
  z-index: 1004;
  background-color: rgb(var(--v-theme-surface))
}
</style>
