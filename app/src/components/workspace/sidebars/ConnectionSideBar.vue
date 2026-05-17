<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <side-bar
    v-model="model"
    :title="title"
    :icon="mdiArrowLeftRight"
    max-width="700px"
  >
    <template #actions>
      <add-nodes-chip
        v-if="!showEmptyText"
        :disabled="disableAddingNodes"
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
      <v-card flat>
        <v-card-text>
          <fade-transition>
            <div
              v-if="showEmptyText"
              class="d-flex flex-column align-center"
            >
              <v-icon
                class="text-grey"
                :icon="mdiCancel"
                size="90"
              />
              <div class="text-title-large mt-2">
                No connection data available
              </div>
            </div>
            <div v-else-if="transactionList !== null">
              <v-card-text>
                The following transactions connect the two nodes.
              </v-card-text>
              <v-data-table
                v-model:sort-by="identitiesSortBy"
                :headers="filteredHeaders"
                :items="transactionList?transactionList:[]"
                item-key="txhash"
                :loading="!transactionList"
                items-per-page="50"
              >
                <template #item.txhash="{ item }">
                  <workspace-link
                    style="max-width:200px"
                    :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
                           params: { id: item.txhash, blockchainMode: route.params.blockchainMode }}"
                  >
                    {{ item.txhash }}
                  </workspace-link>
                </template>
                <template #item.txtype="{ item }">
                  <span>{{ capitalize(item.txtype) }}</span>
                </template>
                <template #item.ts="{ item }">
                  <span>{{ item.ts.toLocaleString() }}</span>
                </template>
                <template #item.fee="{ item }">
                  <span>{{ convertAmount(item.fee) }}</span>
                </template>
                <template #item.inputAmount="{ item }">
                  <span>{{ convertAmount(item.inputAmount) }}</span>
                </template>
                <template #item.outputAmount="{ item }">
                  <span>{{ convertAmount(item.outputAmount) }}</span>
                </template>
              </v-data-table>
            </div>
            <div v-else-if="transactions !== null">
              <v-alert
                color="info"
                variant="tonal"
                density="compact"
                class="mb-5"
              >
                <div class="d-flex align-center">
                  <div>
                    Outputs which connect the two nodes are <span class="textBorder">outlined</span>. Only show outlined outputs?
                  </div>
                  <v-spacer />
                  <!-- need to set min-width so the switch does not shrink when less space is available -->
                  <v-switch
                    v-model="showOnlyHighlightedOutputs"
                    class="ms-2"
                    inset
                    density="compact"
                    hide-details
                    min-width="55px"
                  />
                </div>
              </v-alert>
              <!-- duplicate transaction hashes can exist -> loop through all results
               (e.g. d5d27987d2a3dfc724e359870c6644b40e497bdc0589a033220fe15429d88599 in Bitcoin) -->
              <template
                v-for="t in transactions"
                :key="t.txhash+t.bid"
              >
                <transaction
                  show-fingerprint-link
                  :show-heuristic-editor-link="false"
                  :tx="t"
                  show-details
                  :embed="false"
                  :highlight-transaction="connectionTarget.transactionHash"
                  :filter-highlighted-outputs="showOnlyHighlightedOutputs"
                />
              </template>
            </div>
          </fade-transition>
        </v-card-text>
      </v-card>
    </template>
  </side-bar>
</template>

<script setup>
import {mdiArrowLeftRight, mdiCancel} from '@mdi/js';
import {computed, onUpdated, ref} from 'vue';
import {useRoute} from 'vue-router';
import {capitalize, convertAmount, getDakarClient} from '../../../utilities/index.js';
import {
	WORKSPACE_NODE_TYPE_SELECTOR,
	WORKSPACE_NODE_TYPE_CLUSTER,
	ROUTE_NAME_TRANSACTION_PAGE,
	WORKSPACE_NODE_TYPE_TRANSACTION,
} from '@/constants/index.js';
import SideBar from '@/components/common/SideBar.vue';
import Transaction from '@/components/explorer/transaction/Transaction.vue';
import FadeTransition from '@/components/common/FadeTransition.vue';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';
import AddNodesChip from '@/components/workspace/sidebars/AddNodesChip.vue';
import {useWorkspaceStore} from '@/pinia/workspace.js';
import Alert from '@/components/common/Alert.vue';

const props = defineProps({
	connection: {type: Object, required: true},
	workspaceUid: {type: String, required: true},
	disableAddingNodes: {type: Boolean, required: true},
});

const emit = defineEmits(['addNodes']);

const model = defineModel({type: Boolean});
const route = useRoute();
const workspaceStore = useWorkspaceStore();
const dakar = getDakarClient(route.params.blockchainMode);

let oldConnection = null;
const connectionSource = ref(null);
const connectionTarget = ref(null);
// For cluster <-> cluster and cluster <-> heuristic
const transactionList = ref(null);
// For transaction <-> transaction
const transactions = ref(null);
const showEmptyText = ref(false);
const identitiesSortBy = ref([{key: 'ts', order: 'desc'}]);
const filteredHeaders = ref([]);
const showOnlyHighlightedOutputs = ref(false);
const showSelectAddresses = ref(true);
const showSelectTransactions = ref(true);
const errorMsg = ref('');

const headers = [
	{
		title: 'Hash', key: 'txhash', align: 'start', sortable: false,
	},
	{title: 'Type', key: 'txtype'},
	{title: 'Timestamp', key: 'ts'},
	{title: 'Fee', key: 'fee'},
	{title: 'Input Amount', key: 'inputAmount'},
	{title: 'Output Amount', key: 'outputAmount'},
];

// Holds all transactions which can be selected and added to the workspace
const selectableEntities = new Map();

// Hooks

// eslint-disable-next-line complexity
onUpdated(async () => {
	if (props.connection?.target?.uid && props.connection.source?.uid) {
		const sourceUID = props.connection.source.uid;
		const targetUID = props.connection.target.uid;

		if (oldConnection && (sourceUID === oldConnection.source.uid && targetUID === oldConnection.target.uid)) {
			return;
		}

		workspaceStore.workspaceNodes.clear();
		selectableEntities.clear();
		showOnlyHighlightedOutputs.value = false;
		showEmptyText.value = false;
		oldConnection = props.connection;
		connectionSource.value = props.connection.source;
		connectionTarget.value = props.connection.target;

		// Only pull data if the pair is [cluster,cluster] or [heuristic,cluster]
		if (
		// Cluster <-> cluster
			(connectionSource.value.type === WORKSPACE_NODE_TYPE_CLUSTER
				&& connectionTarget.value.type === WORKSPACE_NODE_TYPE_CLUSTER)
			// Cluster <-> transaction
			|| (connectionSource.value.type === WORKSPACE_NODE_TYPE_SELECTOR
				&& connectionTarget.value.type === WORKSPACE_NODE_TYPE_CLUSTER)
			|| (connectionSource.value.type === WORKSPACE_NODE_TYPE_CLUSTER
				&& connectionTarget.value.type === WORKSPACE_NODE_TYPE_SELECTOR)
			// Cluster <-> transaction
			|| (connectionSource.value.type === WORKSPACE_NODE_TYPE_TRANSACTION
				&& connectionTarget.value.type === WORKSPACE_NODE_TYPE_CLUSTER)
			|| (connectionSource.value.type === WORKSPACE_NODE_TYPE_CLUSTER
				&& connectionTarget.value.type === WORKSPACE_NODE_TYPE_TRANSACTION)
		) {
			await getConnectionData();
		} else if (connectionSource.value.type === WORKSPACE_NODE_TYPE_TRANSACTION
			&& connectionTarget.value.type === WORKSPACE_NODE_TYPE_TRANSACTION) {
			await getTransactionData();
		} else {
			showEmptyText.value = true;
		}
	}
});

// Computed
const title = computed(() => {
	if (showEmptyText.value) {
		return 'Connections';
	}

	if (transactionList.value !== null) {
		return 'Connection List';
	}

	if (transactions.value !== null && transactions.value[0]?.txhash) {
		return `Transaction ${transactions.value[0].txhash}`;
	}

	return 'Connections';
});

// Functions
async function getConnectionData() {
	if (!connectionSource.value?.uid || !connectionTarget.value?.uid || !props.workspaceUid) {
		return;
	}

	transactionList.value = null;
	transactions.value = null;
	errorMsg.value = '';

	try {
		const response = await dakar.workspace.workspacesConnectionPost({
			state: {
				firstNode: {
					uid: connectionSource.value.uid,
					type: connectionSource.value.type,
				},
				secondNode: {
					uid: connectionTarget.value.uid,
					type: connectionTarget.value.type,
				},
				workspaceUID: props.workspaceUid,
			},
		});

		if (response.amountTransactions) {
			let hasTxType = false;
			transactionList.value = response.amountTransactions.map(d => {
				if (d.txhash) {
					selectableEntities.set(d.txhash, {id: d.txhash, type: WORKSPACE_NODE_TYPE_TRANSACTION});
				}

				if (d.txtype) {
					hasTxType = true;
				}

				d.ts = new Date(d.ts);
				return d;
			});

			showSelectTransactions.value = true;
			showSelectAddresses.value = false;

			// If data has no transaction type, so remove it from header
			filteredHeaders.value = hasTxType ? headers : headers.filter(d => d.key !== 'txtype');
		} else if (response.frontendTransactions) {
			transactions.value = response.frontendTransactions;
			addTransactionEntitiesToMap(response.frontendTransactions);
		} else {
			transactionList.value = [];
		}
	} catch (error) {
		errorMsg.value = error;
	}
}

function addOutputToSelectableEntities(output) {
	if (output.txhash) {
		selectableEntities.set(output.txhash, {id: output.txhash, type: WORKSPACE_NODE_TYPE_TRANSACTION});
	}

	if (output.addresshash) {
		selectableEntities.set(output.addresshash, {id: output.addresshash, type: WORKSPACE_NODE_TYPE_CLUSTER});
	}
}

function addTransactionEntitiesToMap(txs) {
	for (const t of txs) {
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
}

async function getTransactionData() {
	if (!connectionSource.value?.transactionHash) {
		return;
	}

	transactionList.value = null;
	transactions.value = null;
	errorMsg.value = '';

	try {
		const response = await dakar.data.blockchainTransactionsHashGet({hash: connectionSource.value.transactionHash});

		if (response.transactions) {
			transactions.value = response.transactions;
			addTransactionEntitiesToMap(response.transactions);
		}
	} catch (error) {
		errorMsg.value = error;
	}
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

function emitAddNodes(nodes) {
	emit('addNodes', nodes);
	model.value = false;
}

</script>

<style scoped>
.textBorder {
  border-style: solid;
  border-radius: 5px;
  border-width: 1px;
}
</style>
