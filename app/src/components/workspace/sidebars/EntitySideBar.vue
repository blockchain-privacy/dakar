<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <side-bar
    v-model="model"
    :title="title"
    :icon="sideBarIcon"
    max-width="700px"
  >
    <template
      v-if="!isLoading && entityData"
      #actions
    >
      <!-- sort in reverse so delete action is in first place -->
      <template
        v-for="(item, index) in nodeActions.toSorted((a,b) => b.title.localeCompare(a.title))"
        :key="index"
      >
        <v-chip
          v-if="!item.show || item.show()"
          :disabled="item.disabled && item.disabled()"
          rounded
          class="me-2"
          :color="item.color"
          :prepend-icon="item.icon"
          @click="() => {item.action(item); model=false;}"
        >
          {{ item.title }}
        </v-chip>
      </template>
      <v-chip
        v-if="type === WORKSPACE_NODE_TYPE_TRANSACTION && isDestination(entityData[0]?.txtype)"
        rounded
        color="primary"
        class="me-2"
        @click="emitFingerprint"
      >
        <v-icon
          :icon="mdiFingerprint"
          start
        />
        Fingerprint
      </v-chip>
      <privacy-chip
        v-if="type === WORKSPACE_NODE_TYPE_TRANSACTION && entityData[0]?.txtype"
        :transaction-type="entityData[0].txtype"
      />
      <v-chip
        v-else-if="type === WORKSPACE_NODE_TYPE_SELECTOR && (entityData?.clusterCount > 0 || entityData?.selectorCount > 0)"
        rounded
        color="primary"
        variant="tonal"
        :prepend-icon="mdiFileDownloadOutline"
        @click="downloadReport"
      >
        Download
      </v-chip>
    </template>
    <template #secondaryActions>
      <add-nodes-chip
        :disabled="disableAddingNodes || auxiliaryData?.loading"
        :show-select-all-addresses="showSelectAddresses"
        :show-select-all-transactions="showSelectTransactions"
        @add-nodes="emitAddNodes"
        @select-all-addresses="selectAllAddresses"
        @deselect-all-addresses="deselectAllAddresses"
        @select-all-transactions="selectAllTransactions"
        @deselect-all-transactions="deselectAllTransactions"
      />
    </template>
    <template #body>
      <alert :text="errorMsg" />
      <v-skeleton-loader
        v-if="isLoading"
        class="mx-auto"
        width="600px"
        type="list-item-three-line, list-item-three-line, list-item-three-line"
      />
      <template v-else>
        <div v-if="type === WORKSPACE_NODE_TYPE_TRANSACTION && entityData?.length">
          <!-- duplicate transaction hashes can exist -> loop through all results
            (e.g. d5d27987d2a3dfc724e359870c6644b40e497bdc0589a033220fe15429d88599 in Bitcoin) -->
          <template
            v-for="t in entityData"
            :key="t.txhash+t.bid"
          >
            <transaction
              :tx="t"
              :show-heuristic-editor-link="false"
              show-fingerprint-link
              show-details
              :embed="false"
            />
          </template>
        </div>
        <address-view
          v-else-if="type === WORKSPACE_NODE_TYPE_CLUSTER && entityData"
          :address-data="entityData"
        />
        <selector-details
          v-else-if="isHeuristic || isTxProp || isTxGraph"
          :selector-type="auxiliaryData.selectorType"
          :selector-data="entityData"
          @cluster-selected="handleClusterSelected"
          @cluster-deselected="handleClusterDeselected"
        />
        <div v-else>
          Type not recognized
        </div>
      </template>
    </template>
  </side-bar>
</template>

<script setup>
import {
	mdiBlender,
	mdiCardBulletedOutline,
	mdiFileDownloadOutline,
	mdiFilter,
	mdiFingerprint,
	mdiGraph,
	mdiShapeCirclePlus,
	mdiTransfer,
} from '@mdi/js';
import {
	computed,
	onUpdated,
	ref,
	watch,
} from 'vue';
import {useRoute} from 'vue-router';
import SideBar from '@/components/common/SideBar.vue';
import Transaction from '@/components/explorer/transaction/Transaction.vue';
import AddressView from '@/components/explorer/address/Address.vue';
import PrivacyChip from '@/components/common/PrivacyChip.vue';
import {useCacheStore} from '@/pinia/cache.js';
import {getCurrentDate, getDakarClient, isDestination} from '@/utilities/index.js';
import {
	SELECTOR_TYPE_HEURISTIC,
	SELECTOR_TYPE_TX_GRAPH,
	SELECTOR_TYPE_TX_PROP,
	WORKSPACE_NODE_TYPE_CLUSTER,
	WORKSPACE_NODE_TYPE_SELECTOR,
	WORKSPACE_NODE_TYPE_TRANSACTION,
} from '@/constants/index.js';
import {useWorkspaceStore} from '@/pinia/workspace.js';
import AddNodesChip from '@/components/workspace/sidebars/AddNodesChip.vue';
import SelectorDetails from '@/components/workspace/sidebars/SelectorDetails.vue';
import Alert from '@/components/common/Alert.vue';

const props = defineProps({
	identifier: {type: String, required: true},
	type: {type: String, required: true},
	workspaceUid: {type: String, required: true},
	auxiliaryData: {type: Object, required: false, default: null},
	disableAddingNodes: {type: Boolean, required: true},
	nodeActions: {type: Array, required: true},
});
const emit = defineEmits(['addNodes', 'fingerprint-transaction']);
const model = defineModel({type: Boolean});

const route = useRoute();
const cacheStore = useCacheStore();
const workspaceStore = useWorkspaceStore();

const dakar = getDakarClient(route.params.blockchainMode);

const isLoading = ref(true);
const entityData = ref();
const showSelectAddresses = ref(true);
const showSelectTransactions = ref(true);
const errorMsg = ref('');

let oldIdentifier = null;

// Holds all transactions which can be selected and added to the workspace
const selectableEntities = new Map();

// Computed
const title = computed(() => {
	const unknownType = 'Unknown Entity Type';
	switch (props.type) {
		case WORKSPACE_NODE_TYPE_TRANSACTION:
			return `Transaction ${props.identifier}`;
		case WORKSPACE_NODE_TYPE_CLUSTER:
			return `Address ${props.identifier}`;
		case WORKSPACE_NODE_TYPE_SELECTOR:
			switch (props.auxiliaryData?.selectorType) {
				case SELECTOR_TYPE_HEURISTIC: return 'CoinJoin Heuristic';
				case SELECTOR_TYPE_TX_GRAPH: return 'Graph Selector';
				case SELECTOR_TYPE_TX_PROP: return 'Property Selector';
				default:
					return unknownType;
			}

		default:
			return unknownType;
	}
});

const isHeuristic = computed(() => props.type === WORKSPACE_NODE_TYPE_SELECTOR
	&& props.auxiliaryData.selectorType === SELECTOR_TYPE_HEURISTIC);
const isTxProp = computed(() => props.type === WORKSPACE_NODE_TYPE_SELECTOR
	&& props.auxiliaryData.selectorType === SELECTOR_TYPE_TX_PROP);
const isTxGraph = computed(() => props.type === WORKSPACE_NODE_TYPE_SELECTOR
	&& props.auxiliaryData.selectorType === SELECTOR_TYPE_TX_GRAPH);

// Watchers

watch(() => props.identifier, async () => {
	await updateEntityData();
});

// Hooks
onUpdated(async () => {
	await updateEntityData();
});

async function updateEntityData() {
	if (props.identifier && props.identifier !== oldIdentifier) {
		workspaceStore.workspaceNodes.clear();
		isLoading.value = true;
		oldIdentifier = props.identifier;
		// Check if value is in cache, otherwise get data from backend
		const cacheValue = cacheStore.get(props.identifier);
		entityData.value = null;
		if (cacheValue === undefined) {
			switch (props.type) {
				case WORKSPACE_NODE_TYPE_TRANSACTION: {
					await getTransactionData();
					break;
				}

				case WORKSPACE_NODE_TYPE_CLUSTER: {
					await getAddressData();
					break;
				}

				case WORKSPACE_NODE_TYPE_SELECTOR: {
					await getSelectorData();
					break;
				}
 // No default
			}
		} else {
			entityData.value = cacheValue;
		}

		setSelectableEntities();

		isLoading.value = false;
	}
}

// Computed
const sideBarIcon = computed(() => {
	switch (props.type) {
		case WORKSPACE_NODE_TYPE_TRANSACTION:
			return mdiTransfer;
		case WORKSPACE_NODE_TYPE_CLUSTER:
			return mdiCardBulletedOutline;
		case WORKSPACE_NODE_TYPE_SELECTOR:
			switch (props.auxiliaryData.selectorType) {
				case SELECTOR_TYPE_HEURISTIC: return mdiBlender;
				case SELECTOR_TYPE_TX_PROP: return mdiFilter;
				case SELECTOR_TYPE_TX_GRAPH: return mdiGraph;
				default: return mdiShapeCirclePlus;
			}

		default:
			return mdiShapeCirclePlus;
	}
});

// Functions
function addOutputToSelectableEntities(output) {
	if (output.txhash) {
		selectableEntities.set(output.txhash, {id: output.txhash, type: WORKSPACE_NODE_TYPE_TRANSACTION});
	}

	if (output.addresshash) {
		selectableEntities.set(output.addresshash, {id: output.addresshash, type: WORKSPACE_NODE_TYPE_CLUSTER});
	}
}

function setSelectableEntities() {
	selectableEntities.clear();
	switch (props.type) {
		case WORKSPACE_NODE_TYPE_TRANSACTION:
			for (const t of entityData.value) {
				if (t.inputs) {
					t.inputs.forEach(element => {
						addOutputToSelectableEntities(element);
					});
				}

				if (t.outputs) {
					t.outputs.forEach(element => {
						addOutputToSelectableEntities(element);
					});
				}
			}

			showSelectTransactions.value = true;
			showSelectAddresses.value = true;

			break;
		case WORKSPACE_NODE_TYPE_CLUSTER:
			for (const output of entityData.value.outputs) {
				if (output.inputTransactionHash) {
					selectableEntities.set(output.inputTransactionHash, {id: output.inputTransactionHash, type: WORKSPACE_NODE_TYPE_TRANSACTION});
				}

				if (output.outputTransactionHash) {
					selectableEntities.set(output.outputTransactionHash, {id: output.outputTransactionHash, type: WORKSPACE_NODE_TYPE_TRANSACTION});
				}
			}

			showSelectTransactions.value = true;
			showSelectAddresses.value = false;

			break;
		case WORKSPACE_NODE_TYPE_SELECTOR:
			switch (props.auxiliaryData.selectorType) {
				case SELECTOR_TYPE_HEURISTIC:
				case SELECTOR_TYPE_TX_GRAPH:
				case SELECTOR_TYPE_TX_PROP:
					setSelectableSelectorElements();
					break;
				default:
			}

			showSelectTransactions.value = true;
			showSelectAddresses.value = false;

			break;
		default:
	}
}

function setSelectableSelectorElements() {
	for (const tx of entityData.value.transactions) {
		if (tx.txhash) {
			selectableEntities.set(tx.txhash, {id: tx.txhash, type: WORKSPACE_NODE_TYPE_TRANSACTION, cluster: tx.cluster});
		}
	}
}

async function getTransactionData() {
	if (props.identifier === '') {
		return;
	}

	errorMsg.value = '';
	try {
		const response = await dakar.data.blockchainTransactionsHashGet({hash: props.identifier});
		entityData.value = response.transactions;
		cacheStore.set(props.identifier, response.transactions);
	} catch (error) {
		setErrorMessage(error);
	}
}

async function getAddressData() {
	if (props.identifier === '') {
		return;
	}

	errorMsg.value = '';
	try {
		const response = await dakar.data.blockchainAddressesHashGet({hash: props.identifier});
		entityData.value = response.address;
		cacheStore.setTTL(props.identifier, response.address, 30);
	} catch (error) {
		setErrorMessage(error);
	}
}

async function getSelectorData() {
	if (!props.identifier || !props.workspaceUid) {
		return;
	}

	let tmpEntityData;
	errorMsg.value = '';
	switch (props.auxiliaryData.selectorType) {
		case SELECTOR_TYPE_HEURISTIC:
			{
				const opt = props.auxiliaryData.heuristicOptions;

				tmpEntityData = {
					heuristicParameter: opt.parameter,
					heuristicExcludeSpendingGaps: opt.excludeSpendingGaps,
					heuristicCustomClusters: opt.clusterTypes?.length > 0,
					heuristicTypeTitle: props.auxiliaryData.displayType,
					heuristicParameterTitle: props.auxiliaryData.parameterTitle,
					clusterCount: props.auxiliaryData.selectorResultCount,
					selectorUid: props.auxiliaryData.uid,
					selectorStatus: props.auxiliaryData.selectorStatus,
					selectorErrorCode: props.auxiliaryData.selectorErrorCode,
					heuristicTimestamp: new Date(props.auxiliaryData.selectorModified),
					transactions: [],
				};
			}

			// Check if data has to be loaded from backend
			if (!tmpEntityData.clusterCount) {
				entityData.value = tmpEntityData;
				return;
			}

			break;
		case SELECTOR_TYPE_TX_PROP:
			tmpEntityData = props.auxiliaryData.txPropOptions;
			tmpEntityData.selectorUid = props.auxiliaryData.uid;
			tmpEntityData.selectorTimestamp = new Date(props.auxiliaryData.selectorModified);
			tmpEntityData.selectorCount = props.auxiliaryData.selectorResultCount;
			tmpEntityData.selectorStatus = props.auxiliaryData.selectorStatus;
			tmpEntityData.selectorErrorCode = props.auxiliaryData.selectorErrorCode;
			tmpEntityData.selectorTotalResultCount = props.auxiliaryData.selectorTotalResultCount;
			tmpEntityData.transactions = [];

			// Check if data has to be loaded from backend
			if (!tmpEntityData.selectorCount) {
				entityData.value = tmpEntityData;
				return;
			}

			break;
		case SELECTOR_TYPE_TX_GRAPH:
			tmpEntityData = props.auxiliaryData.txGraphOptions;
			tmpEntityData.selectorUid = props.auxiliaryData.uid;
			tmpEntityData.selectorTimestamp = new Date(props.auxiliaryData.selectorModified);
			tmpEntityData.selectorCount = props.auxiliaryData.selectorResultCount;
			tmpEntityData.selectorStatus = props.auxiliaryData.selectorStatus;
			tmpEntityData.selectorErrorCode = props.auxiliaryData.selectorErrorCode;
			tmpEntityData.selectorTotalResultCount = props.auxiliaryData.selectorTotalResultCount;
			tmpEntityData.transactions = [];

			// Check if data has to be loaded from backend
			if (!tmpEntityData.selectorCount) {
				entityData.value = tmpEntityData;
				return;
			}

			break;
		default:
			// Invalid type
			return;
	}

	try {
		const response = await dakar.workspace.workspacesSelectorResultsPost({
			selector: {selectorUID: props.identifier, workspaceUID: props.workspaceUid},
		});

		if (response.transactions?.length > 0) {
			tmpEntityData.transactions = response.transactions;
		}

		entityData.value = tmpEntityData;
		cacheStore.set(props.identifier, tmpEntityData);
	} catch (error) {
		setErrorMessage(error);
	}
}

function setErrorMessage(msg) {
	errorMsg.value = msg;
}

async function downloadReport() {
	if (!entityData.value.selectorUid || !props.workspaceUid) {
		return;
	}

	errorMsg.value = '';
	try {
		const response = await dakar.workspace.workspacesSelectorReportPost({
			selector: {
				workspaceUID: props.workspaceUid,
				selectorUID: entityData.value.selectorUid,
			},
		});
		// Looks hacky, but it is the only way with good UX
		const a = document.createElement('a');
		a.href = URL.createObjectURL(response);

		a.setAttribute(
			'download',
			`selector_report_${getCurrentDate()}_${entityData.value.selectorUid}.csv`,
		);
		a.click();
		a.remove();
	} catch (error) {
		setErrorMessage(error);
	}
}

function emitAddNodes(nodes) {
	emit('addNodes', nodes);
	model.value = false;
}

function emitFingerprint() {
	emit('fingerprint-transaction', props.identifier);
	model.value = false;
}

function selectAllTransactions() {
	workspaceStore.setWorkspaceNodes([...selectableEntities.values()]
		.filter(d => d.type === WORKSPACE_NODE_TYPE_TRANSACTION));
}

function selectAllAddresses() {
	workspaceStore.setWorkspaceNodes([...selectableEntities.values()]
		.filter(d => d.type === WORKSPACE_NODE_TYPE_CLUSTER));
}

function deselectAllTransactions() {
	workspaceStore.removeNodesFromMap([...workspaceStore.workspaceNodes.values()]
		.filter(d => d.type === WORKSPACE_NODE_TYPE_TRANSACTION).map(d => d.id));
}

function deselectAllAddresses() {
	workspaceStore.removeNodesFromMap([...workspaceStore.workspaceNodes.values()]
		.filter(d => d.type === WORKSPACE_NODE_TYPE_CLUSTER).map(d => d.id));
}

function handleClusterSelected(clusterID) {
	if (clusterID === undefined || clusterID === null) {
		return;
	}

	workspaceStore.setWorkspaceNodes([...selectableEntities.values()]
		.filter(d => d.cluster === clusterID));
}

function handleClusterDeselected(clusterID) {
	if (clusterID === undefined || clusterID === null) {
		return;
	}

	workspaceStore.removeNodesFromMap([...selectableEntities.values()]
		.filter(d => d.cluster === clusterID).map(d => d.id));
}

</script>

<style scoped>

</style>
