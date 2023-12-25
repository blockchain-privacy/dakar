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
          Server is not ready to accept request for new heuristics. Please try again later. Existing heuristic results
          can be viewed.
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
        class="workspace-toolbar"
      >
        <v-card-text class="d-flex align-center pa-0">
          <v-icon
            class="mx-3"
            icon="$graphIcon"
            size="x-large"
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
            :disabled="isLoading"
            :append-inner-icon="mdiMagnify"
            @click:append-inner="handleGraphQuery(graphQuery)"
            @keydown.enter="handleGraphQuery(graphQuery)"
          />
          <v-btn
            style="min-width: 32px !important;"
            class="ms-3 px-2"
            variant="text"
            :disabled="banner.show || executionStatus.executing"
            @click="nodeGraph.centerGraph()"
          >
            <v-icon>{{ mdiImageFilterCenterFocus }}</v-icon>
            <div class="hidden-sm-and-down">
              Center Graph
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
        <v-progress-linear
          v-if="isLoading"
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
        <v-dialog
          :model-value="isLoading"
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
          :auxiliary-data="entityAuxiliaryData"
          :type="entityType"
          @add-heuristic="openTypeSelectionSheet"
          @add-node="handleGraphQuery"
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
} from '@mdi/js';
import HeuristicTypeSelectionSideBar from '../heuristic/HeuristicTypeSelectionSideBar.vue';
import {
	APPLICATION_NAME,
	CLUSTER_TYPE_CUSTOM,
	ROUTE_NAME_WORKSPACE_PAGE,
	ROUTE_NAME_WORKSPACES_PAGE,
} from '@/constants';
import ContextMenu from '../common/ContextMenu.vue';
import {getColorMap, handleError, isDestination} from '@/utilities';
import {inject, nextTick, onMounted, onUnmounted, ref, watch} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import NodeGraph from '@/d3Documents/nodeGraph';
import {sleep} from '@/d3Documents/util';
import EntitySideBar from '@/components/workspace/EntitySideBar.vue';

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();
const context = {addMessage: msgStore.addMessage, $route: route};

const newUidPrefix = 'newUid_';

const colorMap = getColorMap();
colorMap.set('heuristic', '#4CAF50');
colorMap.set('cluster', '#CDDC39');
// Non-privacy transaction
colorMap.set('transaction', '#607D8B');

const nodeGraph = new NodeGraph(colorMap);

let uidCounter = 1;
let data = null;
// Node which triggered the heuristic type selection,
// and to which will be the parent of the new heuristic.
// May be a destination transaction or another heuristic
let newHeuristicParentNodeUID = '';

const wasAutoSaved = ref(false);
const isLoading = ref(false);
const graphQuery = ref('');
const workspaceUID = ref('');
const workspaceName = ref('');
const workspaceModificationTime = ref();
const isSaving = ref(false);
const isAddHeuristicSheetOpen = ref(false);
const isEntitySideBarOpen = ref(false);
const entityIdentifier = ref('');
const entityAuxiliaryData = ref(null);
const entityType = ref('');
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

const showContextMenuAddHeuristic = ref(false);
const contextMenuModel = ref({
	display: false,
	x: 0,
	y: 0,
	items: [
		{title: 'Show Properties', icon: mdiChartBar, action: () => nodeGraph.contextMenuNodeClick()},
		{isDivider: true},
		{
			title: 'Add Heuristic',
			icon: mdiShapeSquarePlus,
			action: () => contextMenuOpenTypeSelection(nodeGraph.getContextMenuNode()),
			disabled: () => !banner.value.show && showContextMenuAddHeuristic.value,
		},
		{title: 'Delete Node', icon: mdiDelete, action: removeGraphNode, disabled: () => !banner.value.show},
	],
});

let autoSaveTimer = null;

// Watchers
watch(route, () => {
	newRouting();
});

watch(isAddHeuristicSheetOpen, newVal => {
	// If sheet is being closed reset click state of graph
	if (!newVal.value) {
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

	document.removeEventListener('visibilitychange', onDocumentClose);
});

// Functions
function removeGraphNode() {
	nodeGraph.removeContextMenuNode();
	queueAutoSave();
}

async function handleGraphQuery(query) {
	const trimmedQuery = query.trim();
	if (!trimmedQuery) {
		return;
	}

	graphQuery.value = '';

	nodeGraph.setEnableInteractions(false);
	isLoading.value = true;

	// Wait for auto save to finish
	while (isSaving.value) {
		// eslint-disable-next-line no-await-in-loop
		await sleep(200);
	}

	try {
		const response = await dakar.workspace.addWorkspaceNodePost({
			query: {
				query: trimmedQuery,
				currentState: nodeGraph.exportNodes(),
				workspaceUID: workspaceUID.value,
			},
		});
		if (response.nodes) {
			nodeGraph.addNodes(response.nodes);
			nodeGraph.centerOnNewNodes();
		}
	} catch (e) {
		setErrorMessage(e);
	}

	isLoading.value = false;
	nodeGraph.setEnableInteractions(true);
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

	const parentNode = nodeGraph.getNode(newHeuristicParentNodeUID);
	if (!parentNode) {
		return;
	}

	if (parentNode.children) {
		parentNode.children.push(newHeuristic.uid);
	} else {
		parentNode.children = [newHeuristic.uid];
	}

	nodeGraph.addNodes([parentNode, {
		uid: newHeuristic.uid,
		type: 'heuristic',
		status: 'loading',
		heuristicType: newHeuristic.type,
		heuristicExcludeAddresses: newHeuristic.excludeAddresses,
		heuristicExcludeSpendingGaps: newHeuristic.excludeSpendingGaps,
		heuristicClusterTypes: newHeuristic.clusterTypes,
		heuristicParameter: newHeuristic.parameter,
	}], true);
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
}

function openEntitySideBar(nodeData) {
	// Do not show sidebar for hollow heuristic
	if (nodeData.uid.startsWith(newUidPrefix)) {
		return;
	}

	newHeuristicParentNodeUID = nodeData.uid;

	isAddHeuristicSheetOpen.value = false;
	entityAuxiliaryData.value = null;
	entityType.value = nodeData.type;

	switch (entityType.value) {
		case 'cluster':
			entityIdentifier.value = nodeData.addressHash;
			break;
		case 'transaction':
			entityIdentifier.value = nodeData.transactionHash;
			break;
		case 'heuristic':
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

	isEntitySideBarOpen.value = true;
}

function closeSideBars() {
	isAddHeuristicSheetOpen.value = false;
	isEntitySideBarOpen.value = false;
}

function showContextMenu(e, nodeData) {
	contextMenuModel.value.display = false;

	e.preventDefault();

	if (nodeData?.type === 'heuristic' || isDestination(nodeData.privacyType)) {
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
	wasAutoSaved.value = true;
	if (autoSaveTimer !== null) {
		clearTimeout(autoSaveTimer);
	}

	autoSaveTimer = setTimeout(doAutoSave, 5000);
}

async function doAutoSave() {
	isSaving.value = true;
	autoSaveTimer = null;
	try {
		const response = await dakar.workspace.updateWorkspacePost({
			state: {
				workspaceUID: workspaceUID.value,
				currentState: nodeGraph.exportNodes(),
			},
		});
		if (response.ts) {
			workspaceModificationTime.value = new Date(response.ts);
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

	// Gets all heuristic type configurations
	await getDescriptors();
	if (heuristicDescriptors.value.length === 0) {
		return false;
	}

	// Creates the tab descriptions based on the heuristic categories
	createTabs();
	nodeGraph.populateHeuristicMap(heuristicDescriptors.value);
	nodeGraph.initSvg(svgCanvasId);
	if (!await refreshData()) {
		return false;
	}

	// Update page title
	document.title = `${workspaceName.value} - Workspace - ${APPLICATION_NAME}`;

	nodeGraph.addNodes(data.state);

	await nodeGraph.centerGraph();

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
