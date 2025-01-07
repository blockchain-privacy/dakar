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
        style="max-width:700px"
      >
        <adaptive-toolbar
          :name="workspaceName"
          :selected-item-count="lassoSelectedNodes.length"
          :delete-disabled="!isLassoDeletionEnabled"
          :shortest-path-enabled="isShortestPathLookupEnabled"
          :add-entity-enabled="!isModifyingWorkspace"
          :node-type-items="nodeTypeLabels"
          :transaction-type-items="transactionTypeLabels"
          show-search-field
          show-delete-button
          show-add-selector-button
          @is-selection-enabled="(flag) => nodeGraph.setLassoEnabled(flag)"
          @rearrange="handleMenuRearrange"
          @center="handleMenuCenter"
          @delete-selected="handleMenuDeleteSelected"
          @add-entity="handleGraphQuery"
          @filter-changed="handleMenuFilterChanged"
          @shortest-path-lookup="handleShortestPathLookup"
          @add-selector="showCreateSelectorSheetFromButton"
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
          v-if="isLoadingWorkspace"
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
                  Loading workspace
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
        <create-selector-side-bar
          v-model="isCreateSelectorSheetOpen"
          :descriptors="heuristicDescriptors"
          :selector-type="selectorCreationType"
          :parent-node="selectorParent"
          @add-selector="addNewSelector"
        />
        <entity-side-bar
          v-model="isEntitySideBarOpen"
          :identifier="entityIdentifier"
          :workspace-uid="workspaceUID"
          :auxiliary-data="entityAuxiliaryData"
          :type="entityType"
          :disable-adding-nodes="isModifyingWorkspace"
          :node-actions="nodeActions"
          @add-nodes="checkNodeCount"
          @fingerprint-transaction="openFingerprintSidebar"
        />
        <connection-side-bar
          v-model="isConnectionSideBarOpen"
          :connection="connectionData"
          :workspace-uid="workspaceUID"
          :disable-adding-nodes="isModifyingWorkspace"
          @add-nodes="checkNodeCount"
        />
        <shortest-path-side-bar
          v-model="isShortestPathSideBarOpen"
          :from="shortestPathTransactions[0]"
          :to="shortestPathTransactions[1]"
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
        <fingerprint-side-bar
          v-model="isFingerprintSideBarOpen"
          :transaction-hash="fingerprintTransaction"
          :disable-adding-nodes="isModifyingWorkspace"
          @add-nodes="checkNodeCount"
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
          <v-list
            class="py-0"
            slim
          >
            <template
              v-for="(item, index) in nodeActions"
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
	mdiDelete, mdiFilterPlus,
	mdiNoteEdit,
	mdiNotePlus,
} from '@mdi/js';
import CreateSelectorSideBar from './sidebars/CreateSelectorSideBar.vue';
import {
	APPLICATION_NAME,
	ROUTE_NAME_WORKSPACE_PAGE,
	WORKSPACE_NODE_TYPE_CLUSTER,
	WORKSPACE_NODE_TYPE_SELECTOR,
	WORKSPACE_NODE_TYPE_TRANSACTION,
	WORKSPACE_NODE_TYPE_NOTE,
	PRIVACY_TYPE_DESTINATION,
	SELECTOR_TYPE_HEURISTIC,
	SELECTOR_TYPE_TX_PROP,
	SELECTOR_STATUS_WAITING,
	SELECTOR_TYPE_TX_GRAPH,
	PRIVACY_TYPE_ORIGIN,
	SELECTOR_STATUS_SUCCESS,
	PRIVACY_TYPE_MIXING,
	PRIVACY_TYPE_WASABI_2_ORIGIN,
	PRIVACY_TYPE_WASABI_2_MIXING,
	PRIVACY_TYPE_WASABI_2_DESTINATION,
} from '@/constants';
import {
	capitalize,
	getColorMap, getDakarClient, handleError,
} from '@/utilities';
import {
	computed, nextTick, onMounted, onUnmounted, ref, watch,
} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import {useWorkspaceStore} from '@/pinia/workspace.js';
import NodeGraph from '@/d3Documents/nodeGraph';
import {sleep} from '@/d3Documents/util';
import EntitySideBar from '@/components/workspace/sidebars/EntitySideBar.vue';
import AdaptiveToolbar from '@/components/common/AdaptiveToolbar.vue';
import ConnectionSideBar from '@/components/workspace/sidebars/ConnectionSideBar.vue';
import RoutingDialog from '@/components/workspace/RoutingDialog.vue';
import TextDialog from '@/components/common/TextDialog.vue';
import ConfirmDialog from '@/components/common/ConfirmDialog.vue';
import ShortestPathSideBar from '@/components/workspace/sidebars/ShortestPathSideBar.vue';
import {setNodesDisplayAttributes} from '@/d3Documents/nodeDisplay.js';
import {blenderPlus, graphPlus} from '@/customIcons/index.js';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';
import FingerprintSideBar from '@/components/workspace/sidebars/FingerprintSideBar.vue';

const {getSettings} = storeToRefs(useLocalStore());
const route = useRoute();
const msgStore = useMsgStore();
const workspaceStore = useWorkspaceStore();
const context = {addMessage: msgStore.addMessage, $route: route};
const dakar = getDakarClient(getSettings.value.blockchainMode);

const colorMap = getColorMap(getSettings.value.blockchainMode);
colorMap.set(WORKSPACE_NODE_TYPE_CLUSTER, '#ffe119');
// Non-privacy transaction
colorMap.set(WORKSPACE_NODE_TYPE_TRANSACTION, '#607d8b');

colorMap.set(SELECTOR_TYPE_HEURISTIC, '#4363d8');
colorMap.set(SELECTOR_TYPE_TX_GRAPH, '#42d4f4');
colorMap.set(SELECTOR_TYPE_TX_PROP, '#000075');

const nodeTypeLabels = [
	{text: WORKSPACE_NODE_TYPE_SELECTOR, color: colorMap.get(SELECTOR_TYPE_TX_PROP)},
	{text: WORKSPACE_NODE_TYPE_CLUSTER, color: colorMap.get(WORKSPACE_NODE_TYPE_CLUSTER)},
	{text: WORKSPACE_NODE_TYPE_TRANSACTION, color: colorMap.get(WORKSPACE_NODE_TYPE_TRANSACTION)},
];

const nodeGraph = new NodeGraph(colorMap);
let data = null;
// Holds the references to all selector timers
const selectorTimers = [];

const isAutoSaving = ref(false);
const wasAutoSaved = ref(false);
const isLoadingWorkspace = ref(false);
const isModifyingWorkspace = ref(false);
const workspaceUID = ref('');
const workspaceName = ref('');
// SelectorParent is the parent node of the new selector
const selectorParent = ref(undefined);
const isCreateSelectorSheetOpen = ref(false);
const isEntitySideBarOpen = ref(false);
const isConnectionSideBarOpen = ref(false);
const selectorCreationType = ref('');
const isShortestPathSideBarOpen = ref(false);
const isFingerprintSideBarOpen = ref(false);
const entityIdentifier = ref('');
const entityAuxiliaryData = ref(null);
const entityType = ref('');
const heuristicDescriptors = ref([]);
// Holds a mapping between heuristic type and title
const heuristicTypeMap = new Map();
const connectionData = ref({});
const showRouteGuardDialogModel = ref(false);
const routeGuardTo = ref({});
const showAddNoteDialogModel = ref(false);
const showEditNoteDialogModel = ref(false);
const showWarningDialogModel = ref(false);
const editNoteDialogValue = ref('');
const warningDialogNodes = ref([]);
const lassoSelectedNodes = ref([]);
const fingerprintTransaction = ref('');
const shortestPathTransactions = ref(['', '']);
const contextMenuModel = ref({display: false, x: 0, y: 0});
const nodeActions = ref([
	{
		title: 'Add CoinJoin Heuristic',
		color: 'primary',
		icon: blenderPlus,
		show: () => isHeuristicNode(nodeGraph.getContextNode())
		|| isCoinJoinTransactionNode(nodeGraph.getContextNode()),
		action: () => openCreateSelectorSheet(SELECTOR_TYPE_HEURISTIC, nodeGraph.getContextNode()),
		disabled: () => !acceptsChild(nodeGraph.getContextNode()),
	},
	{
		title: 'Add Property Selector',
		color: 'primary',
		icon: mdiFilterPlus,
		show: () => isTxPropNode(nodeGraph.getContextNode())
		|| isTxGraphNode(nodeGraph.getContextNode())
		|| isHeuristicNode(nodeGraph.getContextNode()),
		action: () => openCreateSelectorSheet(SELECTOR_TYPE_TX_PROP, nodeGraph.getContextNode()),
		disabled: () => isModifyingWorkspace.value || !acceptsChild(nodeGraph.getContextNode()),
	},
	{
		title: 'Add Graph Selector',
		color: 'primary',
		icon: graphPlus,
		show: () => isTransactionNode(nodeGraph.getContextNode()),
		action: () => openCreateSelectorSheet(SELECTOR_TYPE_TX_GRAPH, nodeGraph.getContextNode()),
		disabled: () => isModifyingWorkspace.value || !acceptsChild(nodeGraph.getContextNode()),
	},
	{
		title: 'Add Note',
		color: 'primary',
		icon: mdiNotePlus,
		action: showAddNoteDialog,
		show: () => isModifyingWorkspace.value || !isNote(nodeGraph.getContextNode()),
	},
	{
		title: 'Edit',
		color: 'primary',
		icon: mdiNoteEdit,
		show: () => isNote(nodeGraph.getContextNode()),
		action: () => editNote(nodeGraph.getContextNode()),
		disabled: () => isModifyingWorkspace.value || !acceptsChild(nodeGraph.getContextNode()),
	},
	{
		title: 'Delete',
		icon: mdiDelete,
		action: removeContextNode,
		disabled: () => isModifyingWorkspace.value || !isDeleteEnabled(nodeGraph.getContextNode()),
	},
]);

let autoSaveTimer = null;
const maxNoteLength = 100;

// Watchers
watch(route, () => {
	newRouting();
});

watch(isCreateSelectorSheetOpen, newVal => {
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
const transactionTypeLabels = computed(() => {
	const labels = [];

	getColorMap(getSettings.value.blockchainMode).forEach((v, k) => {
		labels.push({text: k, color: v});
	});

	return labels;
});

const isLassoDeletionEnabled = computed(() => !lassoSelectedNodes.value.some(d => !isDeleteEnabled(d)));
const isShortestPathLookupEnabled = computed(() =>
	lassoSelectedNodes.value.length === 2 && !lassoSelectedNodes.value.some(d => !d.transactionHash));

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
	selectorTimers.forEach(d => clearTimeout(d));

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

function isTransactionNode(node) {
	if (!node) {
		return false;
	}

	return node.type === WORKSPACE_NODE_TYPE_TRANSACTION;
}

function isTxGraphNode(node) {
	if (!node) {
		return false;
	}

	return node.type === WORKSPACE_NODE_TYPE_SELECTOR && node.selectorType === SELECTOR_TYPE_TX_GRAPH;
}

function isTxPropNode(node) {
	if (!node) {
		return false;
	}

	return node.type === WORKSPACE_NODE_TYPE_SELECTOR && node.selectorType === SELECTOR_TYPE_TX_PROP;
}

function isHeuristicNode(node) {
	if (!node) {
		return false;
	}

	return node.type === WORKSPACE_NODE_TYPE_SELECTOR && node.selectorType === SELECTOR_TYPE_HEURISTIC;
}

function isCoinJoinTransactionNode(node) {
	if (!node) {
		return false;
	}

	switch (node.txtype) {
		case PRIVACY_TYPE_ORIGIN:
		case PRIVACY_TYPE_MIXING:
		case PRIVACY_TYPE_DESTINATION:
		case PRIVACY_TYPE_WASABI_2_ORIGIN:
		case PRIVACY_TYPE_WASABI_2_MIXING:
		case PRIVACY_TYPE_WASABI_2_DESTINATION:
			return true;
		default: return false;
	}
}

function isNote(node) {
	if (!node) {
		return false;
	}

	return node.type === WORKSPACE_NODE_TYPE_NOTE;
}

// Checks if a node can be deleted. If a heuristic or a node
// in a heuristic sub graph is loading it return false.
function isDeleteEnabled(contextNode) {
	if (!contextNode || contextNode.selectorStatus === SELECTOR_STATUS_WAITING) {
		return false;
	}

	if (contextNode.type !== WORKSPACE_NODE_TYPE_SELECTOR && !isCoinJoinTransactionNode(contextNode)) {
		return true;
	}

	if (contextNode.children) {
		for (const child of contextNode.children) {
			const childNode = nodeGraph.getNode(child);
			if (!childNode || childNode.type !== WORKSPACE_NODE_TYPE_SELECTOR) {
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

function handleShortestPathLookup() {
	openShortestPathSidebar();
}

function handleMenuFilterChanged(nodeFilter, privacyFilter) {
	nodeGraph.filterNodes(nodeFilter, privacyFilter);
	nodeGraph.draw();
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
			response.nodes = setNodesDisplayAttributes(response.nodes, heuristicTypeMap);
			nodeGraph.addNodes(response.nodes);
			queueAutoSave();
			nodeGraph.centerOnNewNodes();
		} else if (response.duplicateNodeUID) {
			let defaultText = 'Node already in workspace';
			const n = nodeGraph.getNode(response.duplicateNodeUID);
			if (n !== undefined) {
				n.showMarker = true;
				nodeGraph.addNode(n);
				if (n.nodeDisplayTitle && n.type) {
					defaultText = `${capitalize(n.type)} ${n.nodeDisplayTitle} already in workspace`;
				}
			}

			setInfoMessage(defaultText);
		}
	} catch (e) {
		setErrorMessage(e);
	}

	releaseAutosaveLock();
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
			response.nodes = setNodesDisplayAttributes(response.nodes, heuristicTypeMap);
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

function setInfoMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'info', temporary: true, category: route.name,
	});
}

// Sends the new selector to the backend
async function addNewSelector(type, options) {
	const currentNode = nodeGraph.getContextNode();

	let parent;
	let heuristicOptions;
	let txPropOptions;
	let txGraphOptions;

	switch (type) {
		case SELECTOR_TYPE_HEURISTIC:
			{
				if (!currentNode || !currentNode.uid) {
					setErrorMessage('could not determine parent node');
					return;
				}

				parent = currentNode.uid;

				heuristicOptions = options;
				const txHash = getHeuristicTransaction(nodeGraph.getNodes(), currentNode.uid);
				if (!txHash) {
					setErrorMessage('could not determine heuristic transaction');
					return;
				}

				heuristicOptions.transactionHash = txHash;
			}

			break;
		case SELECTOR_TYPE_TX_PROP:
			if (currentNode?.uid) {
				parent = currentNode.uid;
			}

			txPropOptions = options;
			break;
		case SELECTOR_TYPE_TX_GRAPH:
			if (!currentNode || !currentNode.uid) {
				setErrorMessage('parent not set');
				return;
			}

			parent = currentNode.uid;
			txGraphOptions = options;
			break;
		default:
			setErrorMessage('invalid selector type');
			return;
	}

	await lockAutosave();

	try {
		const response = await dakar.workspace.workspacesSelectorPost({
			selector: {
				parent, type, heuristicOptions, txPropOptions, txGraphOptions, workspaceUID: workspaceUID.value,
			},
		});

		response.nodes = setNodesDisplayAttributes(response.nodes, heuristicTypeMap);
		nodeGraph.addNodes(response.nodes);
		startWaitingforSelectors(response.nodes);

		// Immediatly auto save to store coordinates of new node
		queueAutoSave(0);
	} catch (e) {
		setErrorMessage(e);
	}

	releaseAutosaveLock();
}

function startWaitingforSelectors(nodes) {
	// Stop all timers
	selectorTimers.forEach(d => clearTimeout(d));
	// Start new timers
	nodes
		.filter(n => n.selectorStatus === SELECTOR_STATUS_WAITING)
		.forEach(n => addWork(n.uid));
}

function addWork(selectorUID) {
	selectorTimers.push(setTimeout(checkWork, 3000, selectorUID));
}

// Checks if the requested selector is finished executing
async function checkWork(selectorUID) {
	try {
		const response = await dakar.workspace.workspacesSelectorStatusPost({
			selector: {workspaceUID: workspaceUID.value, selectorUID},
		});
		if (response.nodes) {
			response.nodes = setNodesDisplayAttributes(response.nodes, heuristicTypeMap);
			nodeGraph.addNodes(response.nodes);
		} else {
			addWork(selectorUID);
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

	if (node.type === WORKSPACE_NODE_TYPE_SELECTOR) {
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

function openConnectionSheet(d) {
	connectionData.value = d;

	isEntitySideBarOpen.value = false;
	isCreateSelectorSheetOpen.value = false;
	isConnectionSideBarOpen.value = true;
	isShortestPathSideBarOpen.value = false;
	isFingerprintSideBarOpen.value = false;

	// Next tick so watcher actions are executed first
	nextTick(() => nodeGraph.setContextObjectClicked());
}

function showCreateSelectorSheetFromButton() {
	// So no parent is set
	nodeGraph.resetContextNode();
	openCreateSelectorSheet(SELECTOR_TYPE_TX_PROP);
}

function openCreateSelectorSheet(selectorType, parentNode) {
	selectorCreationType.value = selectorType;
	selectorParent.value = parentNode;
	isEntitySideBarOpen.value = false;
	isConnectionSideBarOpen.value = false;
	isCreateSelectorSheetOpen.value = true;
	isShortestPathSideBarOpen.value = false;
	isFingerprintSideBarOpen.value = false;

	// Next tick so watcher actions are executed first
	nextTick(() => nodeGraph.setContextObjectClicked());
}

function openEntitySideBar(nodeData) {
	entityAuxiliaryData.value = null;
	entityType.value = nodeData.type;

	switch (entityType.value) {
		case WORKSPACE_NODE_TYPE_CLUSTER:
			entityIdentifier.value = nodeData.addressHash;
			break;
		case WORKSPACE_NODE_TYPE_TRANSACTION:
			entityIdentifier.value = nodeData.transactionHash;
			break;
		case WORKSPACE_NODE_TYPE_SELECTOR:
			// Brackets so variables have a local scope (more info: https://eslint.org/docs/latest/rules/no-case-declarations)
			{
				entityAuxiliaryData.value = nodeData;
				entityIdentifier.value = nodeData.uid;

				let displayType = '';

				switch (nodeData.selectorType) {
					case SELECTOR_TYPE_HEURISTIC:
						for (const descriptor of heuristicDescriptors.value) {
							if (descriptor.type === nodeData.heuristicOptions?.type) {
								displayType = descriptor.title;
								break;
							}
						}

						break;
					case SELECTOR_TYPE_TX_PROP:
						displayType = 'selector (change me)';
						break;
					default:
				}

				entityAuxiliaryData.value.displayType = displayType;
			}

			break;
		default:
	}

	isCreateSelectorSheetOpen.value = false;
	isConnectionSideBarOpen.value = false;
	isShortestPathSideBarOpen.value = false;
	isFingerprintSideBarOpen.value = false;
	isEntitySideBarOpen.value = true;
	// Next tick so watcher actions are executed first
	nextTick(() => nodeGraph.setContextObjectClicked());
}

function openShortestPathSidebar() {
	shortestPathTransactions.value = lassoSelectedNodes.value.map(d => d.transactionHash);

	isEntitySideBarOpen.value = false;
	isCreateSelectorSheetOpen.value = false;
	isConnectionSideBarOpen.value = false;
	isShortestPathSideBarOpen.value = true;
	isFingerprintSideBarOpen.value = false;
}

function openFingerprintSidebar(txhash) {
	fingerprintTransaction.value = txhash;

	isEntitySideBarOpen.value = false;
	isCreateSelectorSheetOpen.value = false;
	isConnectionSideBarOpen.value = false;
	isShortestPathSideBarOpen.value = false;
	isFingerprintSideBarOpen.value = true;
}

function closeSideBars() {
	isCreateSelectorSheetOpen.value = false;
	isEntitySideBarOpen.value = false;
	isConnectionSideBarOpen.value = false;
	isShortestPathSideBarOpen.value = false;
	isFingerprintSideBarOpen.value = false;
}

function showContextMenu(e) {
	contextMenuModel.value.display = false;

	e.preventDefault();

	contextMenuModel.value.x = e.clientX;
	contextMenuModel.value.y = e.clientY;

	// Need to hide sidebar, otherwise the context node is also used by side bar
	closeSideBars();

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

		heuristicTypeMap.clear();
		heuristicDescriptors.value.forEach(e => heuristicTypeMap.set(e.type, e.title));

		if (response.workspace) {
			data = response.workspace;
			workspaceName.value = data.name;
			data.nodes &&= setNodesDisplayAttributes(data.nodes, heuristicTypeMap);
		} else {
			data = null;
		}

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

	// Update page title
	document.title = `${workspaceName.value} - Workspace - ${APPLICATION_NAME}`;

	nodeGraph.addNodes(data.nodes);
	nodeGraph.centerGraph();

	startWaitingforSelectors(data.nodes);
	return true;
}

// Returns true if to the given node a child can be added.
// This is generally not the case
// - if the node is a selector waiting to be executed
// - if the node is a selector which failed executing
function acceptsChild(node) {
	if (!node) {
		return false;
	}

	if (node.type !== WORKSPACE_NODE_TYPE_SELECTOR) {
		return true;
	}

	return node.selectorStatus && node.selectorStatus === SELECTOR_STATUS_SUCCESS;
}

</script>

<style scoped>

:deep( #svg_canvas ) {
  height: 100%;
  width: 100%;
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
