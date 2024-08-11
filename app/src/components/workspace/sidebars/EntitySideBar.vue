<template>
  <side-bar
    v-model="model"
    :title="title"
    :icon="sideBarIcon"
    max-width="700px"
    :title-one-line="false"
  >
    <template #actions>
      <v-chip
        rounded
        class="me-2"
        variant="tonal"
        :prepend-icon="mdiDelete"
        @click="emitDeleteEntity"
      >
        Delete
      </v-chip>
      <v-chip
        rounded
        class="me-2"
        color="primary"
        variant="tonal"
        :prepend-icon="mdiNotePlus"
        :disabled="disableAddingNodes || auxiliaryData?.loading"
        @click="emitAddNote"
      >
        Add Note
      </v-chip>
      <template v-if="!isLoading && entityData">
        <template
          v-if="type === WORKSPACE_NODE_TYPE_SELECTOR ||
            (type === WORKSPACE_NODE_TYPE_TRANSACTION && entityData[0]?.privacytype >= 0)"
        >
          <v-chip
            v-if="type === WORKSPACE_NODE_TYPE_SELECTOR || isDestination(entityData[0].privacytype)"
            rounded
            color="primary"
            variant="tonal"
            class="me-2"
            :prepend-icon="mdiFilterPlus"
            :disabled="disableAddingNodes || auxiliaryData?.loading"
            @click="handleAddSelectorClick"
          >
            Add Selector
          </v-chip>
        </template>
        <fingerprint-chip
          v-if="type === WORKSPACE_NODE_TYPE_TRANSACTION && entityData[0] && isDestination(entityData[0].privacytype)"
          :transaction-hash="identifier"
          class="me-2"
        />
        <privacy-chip
          v-if="type === WORKSPACE_NODE_TYPE_TRANSACTION&& entityData[0]?.privacytype >= 0"
          :privacy-type="entityData[0].privacytype"
        />
        <exclusion-chip
          v-else-if="type === WORKSPACE_NODE_TYPE_CLUSTER && entityData?.addresshash"
          :address-hash="entityData.addresshash"
        />
        <v-chip
          v-else-if="type === WORKSPACE_NODE_TYPE_SELECTOR && entityData?.clusterCount > 0"
          rounded
          color="primary"
          variant="tonal"
          :prepend-icon="mdiFileDownloadOutline"
          @click="downloadReport"
        >
          Download
        </v-chip>
      </template>
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
      <fade-transition>
        <v-skeleton-loader
          v-if="isLoading"
          class="mx-auto"
          width="600px"
          type="list-item-three-line, list-item-three-line, list-item-three-line"
        />
        <template v-else>
          <template v-if="type === WORKSPACE_NODE_TYPE_TRANSACTION && entityData?.length">
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
                :show-title-bar="false"
              />
            </template>
          </template>
          <address-view
            v-else-if="entityData && type === WORKSPACE_NODE_TYPE_CLUSTER"
            :address-data="entityData"
            :show-title-bar="false"
          />
          <heuristic-details
            v-else-if="entityData.heuristicUid && type === WORKSPACE_NODE_TYPE_SELECTOR"
            :heuristic-data="entityData"
          />
          <div v-else>
            Type not recognized
          </div>
        </template>
      </fade-transition>
    </template>
  </side-bar>
</template>

<script setup>
import {
	mdiCardBulletedOutline,
	mdiChartBar,
	mdiDelete,
	mdiFileDownloadOutline, mdiFilterPlus,
	mdiNotePlus,
	mdiShapeCirclePlus,
	mdiTransfer,
} from '@mdi/js';
import SideBar from '@/components/common/SideBar.vue';
import {
	computed, inject, onUpdated, ref,
} from 'vue';
import Transaction from '@/components/explorer/transaction/Transaction.vue';
import AddressView from '@/components/explorer/address/Address.vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg.js';
import PrivacyChip from '@/components/common/PrivacyChip.vue';
import FadeTransition from '@/components/common/FadeTransition.vue';
import ExclusionChip from '@/components/explorer/address/ExclusionChip.vue';
import {useCacheStore} from '@/pinia/cache.js';
import HeuristicDetails from '@/components/workspace/sidebars/HeuristicDetails.vue';
import {getCurrentDate, isDestination} from '@/utilities/index.js';
import {
	SELECTOR_TYPE_HEURISTIC,
	WORKSPACE_NODE_TYPE_CLUSTER,
	WORKSPACE_NODE_TYPE_SELECTOR,
	WORKSPACE_NODE_TYPE_TRANSACTION,
} from '@/constants/index.js';
import FingerprintChip from '@/components/explorer/transaction/FingerprintChip.vue';
import {useWorkspaceStore} from '@/pinia/workspace.js';
import AddNodesChip from '@/components/workspace/sidebars/AddNodesChip.vue';

const props = defineProps({
	identifier: {type: String, required: true},
	type: {type: String, required: true},
	workspaceUid: {type: String, required: true},
	auxiliaryData: {type: Object, required: false, default: null},
	disableAddingNodes: {type: Boolean, required: true},
});
const emit = defineEmits(['addSelector', 'addNote', 'deleteEntity', 'addNodes']);
const model = defineModel({type: Boolean});

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();
const cacheStore = useCacheStore();
const workspaceStore = useWorkspaceStore();

const isLoading = ref(true);
const entityData = ref();
const showSelectAddresses = ref(true);
const showSelectTransactions = ref(true);

let oldIdentifier = null;

// Holds all transactions which can be selected and added to the workspace
const selectableEntities = new Map();

// Computed
const title = computed(() => {
	switch (props.type) {
		case WORKSPACE_NODE_TYPE_TRANSACTION:
			return `Transaction ${props.identifier}`;
		case WORKSPACE_NODE_TYPE_CLUSTER:
			return `Address ${props.identifier}`;
		case WORKSPACE_NODE_TYPE_SELECTOR:
			if (props.auxiliaryData?.selectorType === SELECTOR_TYPE_HEURISTIC) {
				return 'Heuristic Properties';
			}

			return 'Selector Properties';
		default:
			return 'Unknown entity type';
	}
});

// Hooks
onUpdated(async () => {
	if (props.identifier && props.identifier !== oldIdentifier) {
		workspaceStore.workspaceNodes.clear();
		isLoading.value = true;
		oldIdentifier = props.identifier;
		// Check if value is in cache, otherwise get data from backend
		const cacheValue = cacheStore.getValue(props.identifier);
		entityData.value = null;
		if (cacheValue !== undefined) {
			entityData.value = cacheValue;
		} else if (props.type === WORKSPACE_NODE_TYPE_TRANSACTION) {
			await getTransactionData();
		} else if (props.type === WORKSPACE_NODE_TYPE_CLUSTER) {
			await getAddressData();
		} else if (props.type === WORKSPACE_NODE_TYPE_SELECTOR) {
			await getHeuristicData();
		}

		setSelectableEntities();

		isLoading.value = false;
	}
});

// Computed
const sideBarIcon = computed(() => {
	switch (props.type) {
		case WORKSPACE_NODE_TYPE_TRANSACTION:
			return mdiTransfer;
		case WORKSPACE_NODE_TYPE_CLUSTER:
			return mdiCardBulletedOutline;
		case WORKSPACE_NODE_TYPE_SELECTOR:
			return mdiChartBar;
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
					t.inputs.forEach(addOutputToSelectableEntities);
				}

				if (t.outputs) {
					t.outputs.forEach(addOutputToSelectableEntities);
				}
			}

			showSelectTransactions.value = true;
			showSelectAddresses.value = true;

			break;
		case WORKSPACE_NODE_TYPE_CLUSTER:
			for (const output of entityData.value.addr_outputs) {
				if (output.input_transaction) {
					selectableEntities.set(output.input_transaction, {id: output.input_transaction, type: WORKSPACE_NODE_TYPE_TRANSACTION});
				}

				if (output.output_transaction) {
					selectableEntities.set(output.output_transaction, {id: output.output_transaction, type: WORKSPACE_NODE_TYPE_TRANSACTION});
				}
			}

			showSelectTransactions.value = true;
			showSelectAddresses.value = false;

			break;
		case WORKSPACE_NODE_TYPE_SELECTOR:
			for (const cluster of entityData.value.clusters) {
				for (const tx of cluster.txs) {
					if (tx.txhash) {
						selectableEntities.set(tx.txhash, {id: tx.txhash, type: WORKSPACE_NODE_TYPE_TRANSACTION});
					}
				}
			}

			showSelectTransactions.value = true;
			showSelectAddresses.value = false;

			break;
		default:
	}
}

async function getTransactionData() {
	if (props.identifier === '') {
		return;
	}

	try {
		const response = await dakar.data.blockchainTransactionsHashGet({hash: props.identifier});
		entityData.value = response.transactions;
		cacheStore.setValue(props.identifier, response.transactions);
	} catch (e) {
		setErrorMessage(e);
	}
}

async function getAddressData() {
	if (props.identifier === '') {
		return;
	}

	try {
		const response = await dakar.data.blockchainAddressesHashGet({hash: props.identifier});
		entityData.value = response.address;
		cacheStore.setValue(props.identifier, response.address);
	} catch (e) {
		setErrorMessage(e);
	}
}

async function getHeuristicData() {
	if (!props.identifier || !props.workspaceUid || !props.auxiliaryData?.heuristicOptions) {
		return;
	}

	const opt = props.auxiliaryData.heuristicOptions;

	const tmp = {
		heuristicParameter: opt.parameter,
		heuristicExcludeAddresses: opt.excludeAddresses,
		heuristicExcludeSpendingGaps: opt.excludeSpendingGaps,
		heuristicCustomClusters: opt.clusterTypes?.length > 0,
		heuristicTypeTitle: props.auxiliaryData.displayType,
		clusterCount: props.auxiliaryData.selectorResultCount,
		heuristicUid: props.auxiliaryData.uid,
		heuristicTimestamp: new Date(props.auxiliaryData.selectorModified),
		clusters: [],
	};

	// Check if data has to be loaded from backend
	if (!tmp.clusterCount) {
		entityData.value = tmp;
		return;
	}

	try {
		const response = await dakar.heuristic.heuristicDetailsPost({
			heuristic: {
				heuristicUID: props.identifier,
				workspaceUID: props.workspaceUid,
			},
		});

		if (!response.heuristic) {
			throw new Error('response contains no heuristics');
		}

		tmp.clusters = response.heuristic.clusters;
		entityData.value = tmp;
		cacheStore.setValue(props.identifier, tmp);
	} catch (e) {
		setErrorMessage(e);
	}
}

function setErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: true, category: route.name,
	});
}

async function downloadReport() {
	if (!entityData.value.heuristicUid || !props.workspaceUid) {
		return;
	}

	try {
		const response = await dakar.heuristic.heuristicsReportPost({
			work: {
				workspaceUID: props.workspaceUid,
				heuristicUID: entityData.value.heuristicUid,
			},
		});
		// Looks hacky, but it is the only way with good UX
		const a = document.createElement('a');
		a.href = URL.createObjectURL(response);

		a.setAttribute(
			'download',
			`heuristic_report_${getCurrentDate()}_${entityData.value.heuristicUid}.csv`,
		);
		a.click();
		a.remove();
	} catch (e) {
		setErrorMessage(e);
	}
}

function emitDeleteEntity() {
	emit('deleteEntity', props.identifier);
	model.value = false;
}

function emitAddNote() {
	emit('addNote', props.identifier);
	model.value = false;
}

function emitAddNodes(nodes) {
	emit('addNodes', nodes);
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
		.filter(d => d.type === WORKSPACE_NODE_TYPE_TRANSACTION)
		.map(d => d.id));
}

function deselectAllAddresses() {
	workspaceStore.removeNodesFromMap([...workspaceStore.workspaceNodes.values()]
		.filter(d => d.type === WORKSPACE_NODE_TYPE_CLUSTER)
		.map(d => d.id));
}

function handleAddSelectorClick() {
	emit('addSelector');
}

</script>

<style scoped>

</style>
