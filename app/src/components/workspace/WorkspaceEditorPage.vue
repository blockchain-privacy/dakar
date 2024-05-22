<template>
  <div
    class="flex-column d-flex"
    style="height: 100%;position:relative"
  >
    <div style="height: 100%; width:100%; position: relative">
      <v-card
        v-if="workspaceName"
        :rounded="$vuetify.display.xs?'0':undefined"
        :class="{'toolbar-sm': $vuetify.display.xs, 'toolbar': $vuetify.display.smAndUp}"
      >
        <adaptive-toolbar
          :name="workspaceName"
          :selected-item-count="lassoSelectedNodes.length"
          :delete-enabled="isLassoDeletionEnabled"
          :add-entity-enabled="!isModifyingWorkspace"
          @is-selection-enabled="(flag) => nodeGraph.setLassoEnabled(flag)"
          @rearrange="handleMenuRearrange"
          @center="handleMenuCenter"
          @delete-selected="handleMenuDeleteSelected"
          @add-entity="handleGraphQuery"
        />
        <v-progress-linear
          v-if="isModifyingWorkspace"
          indeterminate
          rounded
          location="bottom"
        />
      </v-card>
      <div
        v-if="workspaceName && wasAutoSaved"
        style=""
        :class="{'text-caption':true, 'auto-save-sm': $vuetify.display.smAndDown, 'auto-save': $vuetify.display.mdAndUp }"
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
          persistent
          max-width="350px"
          contained
          no-click-animation
        >
          <v-card>
            <v-card-text class="text-subtitle-1 d-flex align-center">
              <div style="width:100%">
                <p class="text-center mb-3">
                  Loading workspace ...
                </p>
                <v-progress-linear
                  class="mt-3"
                  indeterminate
                  rounded
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
          @add-note="showAddNoteDialog"
          @add-nodes="checkNodeCount"
          @delete-entity="removeContextNode"
        />
        <connection-side-bar
          v-model="isConnectionSideBarOpen"
          :connection="connectionData"
          :workspace-uid="workspaceUID"
          :disable-adding-nodes="isModifyingWorkspace"
          @add-nodes="checkNodeCount"
        />
        <routing-dialog
          v-model="showRouteGuardDialogModel"
          :to="routeGuardTo"
          :disable-adding-nodes="isModifyingWorkspace"
          @add-node="handleGraphQuery"
        />
        <text-dialog
          v-if="showAddNoteDialogModel"
          v-model="showAddNoteDialogModel"
          title="New Note"
          submit-label="Create"
          input-label="Note content"
          :maxlength="maxNoteLength"
          text-area
          @submit="addNewNote"
        />
        <text-dialog
          v-if="showEditNoteDialogModel"
          v-model="showEditNoteDialogModel"
          title="Edit Note"
          submit-label="OK"
          input-label="Note content"
          :input-value="editNoteDialogValue"
          :maxlength="maxNoteLength"
          text-area
          @submit="changeNote"
        />
        <confirm-dialog
          v-if="showWarningDialogModel"
          v-model="showWarningDialogModel"
          title="Adding Entities"
          confirm-label="Add"
          @confirm="handleWarningDialogConfirm"
        >
          <p class="text-subtitle-1">
            You are about to add <strong>{{ warningDialogNodes.length }}</strong> entities to your workspace.
            Depending on their connections this might take several minutes.
          </p>
        </confirm-dialog>
        <v-menu
          v-model="contextMenuModel.display"
          :open-on-hover="false"
          transition="fade-transition"
          :target="[contextMenuModel.x,contextMenuModel.y]"
        >
          <v-list class="py-0">
            <template
              v-for="(item, index) in contextMenuModel.items"
              :key="index"
            >
              <v-divider v-if="item.isDivider" />
              <v-list-item
                v-else-if="!item.show || item.show()"
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
	mdiCheckCircle,
	mdiDelete,
	mdiNoteEdit,
	mdiNotePlus,
	mdiShapeCirclePlus,
} from '@mdi/js';
import HeuristicTypeSelectionSideBar from './HeuristicTypeSelectionSideBar.vue';
import {
	APPLICATION_NAME,
	CLUSTER_TYPE_CUSTOM,
	ROUTE_NAME_WORKSPACE_PAGE,
	WORKSPACE_NODE_TYPE_CLUSTER,
	WORKSPACE_NODE_TYPE_HEURISTIC,
	WORKSPACE_NODE_TYPE_TRANSACTION,
	WORKSPACE_NODE_TYPE_NOTE,
	PRIVACY_TYPE_DESTINATION,
} from '@/constants';
import {
	getColorMap, handleError, getPrivacyTypeLabel,
} from '@/utilities';
import {
	computed,
	inject, nextTick, onMounted, onUnmounted, ref, watch,
} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import {useWorkspaceStore} from '@/pinia/workspace.js';
import NodeGraph from '@/d3Documents/nodeGraph';
import {sleep} from '@/d3Documents/util';
import EntitySideBar from '@/components/workspace/EntitySideBar.vue';
import AdaptiveToolbar from '@/components/common/AdaptiveToolbar.vue';
import ConnectionSideBar from '@/components/workspace/ConnectionSideBar.vue';
import RoutingDialog from '@/components/workspace/RoutingDialog.vue';
import TextDialog from '@/components/common/TextDialog.vue';
import ConfirmDialog from '@/components/common/ConfirmDialog.vue';

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();
const workspaceStore = useWorkspaceStore();
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
const workspaceUID = ref('');
const workspaceName = ref('');
const isAddHeuristicSheetOpen = ref(false);
const isEntitySideBarOpen = ref(false);
const isConnectionSideBarOpen = ref(false);
const entityIdentifier = ref('');
const entityAuxiliaryData = ref(null);
const entityType = ref('');
const heuristicDescriptors = ref([]);
const heuristicTabItems = ref([]);
const connectionData = ref({});
const showRouteGuardDialogModel = ref(false);
const routeGuardTo = ref({});
const showAddNoteDialogModel = ref(false);
const showEditNoteDialogModel = ref(false);
const showWarningDialogModel = ref(false);
const editNoteDialogValue = ref('');
const warningDialogNodes = ref([]);
const lassoSelectedNodes = ref([]);
const contextMenuModel = ref({
	display: false,
	x: 0,
	y: 0,
	items: [
		{
			title: 'Add Heuristic',
			icon: mdiShapeCirclePlus,
			show: () => nodeGraph.getContextNode()?.type === WORKSPACE_NODE_TYPE_HEURISTIC
        || nodeGraph.getContextNode().privacyTypeLabel === PRIVACY_TYPE_DESTINATION,
			action: () => contextMenuOpenTypeSelection(nodeGraph.getContextNode()),
			disabled: () => nodeGraph.getContextNode()?.loading,
		},
		{
			title: 'Add Note',
			icon: mdiNotePlus,
			action: showAddNoteDialog,
			show: () => nodeGraph.getContextNode()?.type !== WORKSPACE_NODE_TYPE_NOTE,
			disabled: () => nodeGraph.getContextNode()?.loading,
		},
		{
			title: 'Edit',
			icon: mdiNoteEdit,
			show: () => isEditEnabled(nodeGraph.getContextNode()),
			action: () => editNote(nodeGraph.getContextNode()),
			disabled: () => nodeGraph.getContextNode()?.loading,
		},
		{
			title: 'Delete',
			icon: mdiDelete,
			action: removeContextNode,
			disabled: () => !isDeleteEnabled(nodeGraph.getContextNode()),
		},
	],
});

let autoSaveTimer = null;
const maxNoteLength = 100;

// Watchers
watch(route, () => {
	newRouting();
});

watch(isAddHeuristicSheetOpen, newVal => {
	// If sheet is being closed reset click state of graph
	if (!newVal) {
		nodeGraph.resetClick();
		nodeGraph.resetLasso();
	}
});

watch(isEntitySideBarOpen, newVal => {
	// If sheet is being closed reset click state of graph
	if (!newVal) {
		nodeGraph.resetClick();
		nodeGraph.resetLasso();
	}
});

watch(isConnectionSideBarOpen, newVal => {
	// If sheet is being closed reset click state of graph
	if (!newVal) {
		nodeGraph.resetClick();
		nodeGraph.resetLasso();
	}
});

watch(
	() => workspaceStore.workspaceNode,
	newVal => {
		routeGuardTo.value = newVal.to;
		showRouteGuardDialogModel.value = true;
	},
);

// Computed

const isLassoDeletionEnabled = computed(() => !lassoSelectedNodes.value.some(d => !isDeleteEnabled(d)));

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
	workspaceStore.setWorkspaceActive(true);
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
	workspaceStore.setWorkspaceActive(false);
});

// Functions
async function removeGraphNodes(nodes) {
	if (!nodes.length) {
		return;
	}

	if (nodes.some(d => d.loading)) {
		return;
	}

	try {
		const response = await dakar.workspace.workspacesNodeDelete({
			state: {
				nodeUIDs: nodes,
				workspaceUID: workspaceUID.value,
			},
		});

		nodeGraph.removeNodes(response.deletedNodeUIDs);
	} catch (e) {
		setErrorMessage(e);
	}
}

async function removeContextNode() {
	const node = nodeGraph.getContextNode();
	if (!node || node.loading) {
		return;
	}

	await removeGraphNodes([node.uid]);
}

function editNote(note) {
	editNoteDialogValue.value = note.text;
	showEditNoteDialogModel.value = true;
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

function isEditEnabled(contextNode) {
	return contextNode?.type === WORKSPACE_NODE_TYPE_NOTE;
}

// Checks if a node can be deleted. If a heuristic or a node
// in a heuristic sub graph is loading it return false.
function isDeleteEnabled(contextNode) {
	if (!contextNode || contextNode.loading) {
		return false;
	}

	if (contextNode.type !== WORKSPACE_NODE_TYPE_HEURISTIC && contextNode.privacyTypeLabel !== PRIVACY_TYPE_DESTINATION) {
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

async function handleWarningDialogConfirm() {
	await addMultipleNodes(warningDialogNodes.value);
}

// Checks if the node count warning dialog needs to be shown
async function checkNodeCount(nodes) {
	if (nodes.length > 10) {
		showWarningDialogModel.value = true;
		warningDialogNodes.value = nodes;
		return;
	}

	await addMultipleNodes(nodes);
}

function handleMenuRearrange() {
	nodeGraph.reorderNodes();
	queueAutoSave();
}

function handleMenuCenter() {
	nodeGraph.centerGraph();
}

function handleMenuDeleteSelected() {
	removeGraphNodes(lassoSelectedNodes.value.map(d => d.uid));
}

// Receives a node array
async function addMultipleNodes(nodes) {
	if (isModifyingWorkspace.value) {
		return;
	}

	await lockAutosave();

	try {
		const response = await dakar.workspace.workspacesNodesPost({
			query: {
				queries: nodes,
				workspaceUID: workspaceUID.value,
			},
		});
		if (response.nodes) {
			response.nodes = setPrivacyLabels(response.nodes);
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

function setPrivacyLabels(nodes) {
	return nodes.map(d => {
		d.privacyTypeLabel = getPrivacyTypeLabel(d.privacyType);
		return d;
	});
}

async function handleGraphQuery(query) {
	if (isModifyingWorkspace.value) {
		return;
	}

	const trimmedQuery = query.trim();
	if (!trimmedQuery) {
		return;
	}

	await addMultipleNodes([trimmedQuery]);
}

async function changeNote(noteText) {
	const trimmed = noteText.trim();
	if (!trimmed) {
		return;
	}

	const note = nodeGraph.getContextNode();
	note.text = trimmed;

	await addNewNote(noteText, note.uid, note.children[0]);
}

async function addNewNote(noteText, noteUID, childUID) {
	if (isModifyingWorkspace.value) {
		return;
	}

	const trimmed = noteText.trim();
	if (!trimmed) {
		return;
	}

	if (!childUID) {
		// Child uid not set, therefore we have to get it from the context node
		const child = nodeGraph.getContextNode();
		if (!child) {
			return;
		}

		childUID = child.uid;
	}

	await lockAutosave();

	try {
		const response = await dakar.workspace.workspacesNotePost({
			note: {
				uid: noteUID ? noteUID : '',
				childUID,
				text: trimmed,
				workspaceUID: workspaceUID.value,
			},
		});
		if (response.nodes) {
			response.nodes = setPrivacyLabels(response.nodes);
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
		// Immediatly auto save to store coordinates of dummy node
		queueAutoSave(0);
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
			response.nodes = setPrivacyLabels(response.nodes);
			nodeGraph.removeAllNodes(false);
			nodeGraph.addNodes(response.nodes);
		} else {
			addWork(workID);
		}
	} catch (e) {
		setErrorMessage(e);
	}
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

function openConnectionSheet(d) {
	connectionData.value = d;

	isEntitySideBarOpen.value = false;
	isAddHeuristicSheetOpen.value = false;
	isConnectionSideBarOpen.value = true;

	// Next tick so watcher actions are executed first
	nextTick(() => nodeGraph.setContextObjectClicked());
}

function openTypeSelectionSheet() {
	isEntitySideBarOpen.value = false;
	isConnectionSideBarOpen.value = false;
	isAddHeuristicSheetOpen.value = true;
	// Next tick so watcher actions are executed first
	nextTick(() => nodeGraph.setContextObjectClicked());
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
	isConnectionSideBarOpen.value = false;
	isEntitySideBarOpen.value = true;
	// Next tick so watcher actions are executed first
	nextTick(() => nodeGraph.setContextObjectClicked());
}

function closeSideBars() {
	isAddHeuristicSheetOpen.value = false;
	isEntitySideBarOpen.value = false;
	isConnectionSideBarOpen.value = false;
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

function handleLassoSelection() {
	lassoSelectedNodes.value = nodeGraph.getLassoSelectedNodesData();
}

function handleLassoReset() {
	lassoSelectedNodes.value = [];
}

function showAddNoteDialog() {
	showAddNoteDialogModel.value = true;
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
			data.nodes &&= setPrivacyLabels(data.nodes);
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
	if (nodeGraph.isEmpty()) {
		return;
	}

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
		const exportedNodes = nodeGraph.exportNodes();
		if (exportedNodes.length === 0) {
			return;
		}

		await dakar.workspace.workspacesPut({
			state: {
				workspaceUID: workspaceUID.value,
				currentState: exportedNodes,
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

	if (!nodeGraph.setNodeClickCallback(openEntitySideBar)) {
		setErrorMessage('error setting node click handler');
		return false;
	}

	if (!nodeGraph.setLineClickCallback(openConnectionSheet)) {
		setErrorMessage('error setting line click handler');
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

	if (!nodeGraph.setLassoSelectionCallback(handleLassoSelection)) {
		setErrorMessage('error setting lasso selection handler');
		return false;
	}

	if (!nodeGraph.setLassoResetCallback(handleLassoReset)) {
		setErrorMessage('error setting lasso reset handler');
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
	nodeGraph.centerGraph();

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

.auto-save {
  position: absolute;
  top: 10px;
  right: 10px;
  z-index: 1004;
}

.auto-save-sm {
  position: absolute;
  bottom: 10px;
  left: 10px;
  z-index: 1004;
}

.toolbar {
  position: absolute;
  left: 10px;
  top: 10px;
  z-index: 1004;
  background-color: rgb(var(--v-theme-surface))
}

.toolbar-sm {
  position: absolute;
  left: 0;
  top: 0;
  right:0;
  z-index: 1004;
  background-color: rgb(var(--v-theme-surface))
}

</style>
