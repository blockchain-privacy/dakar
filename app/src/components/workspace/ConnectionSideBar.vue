<template>
  <side-bar
    v-model="model"
    :title="title"
    :icon="mdiArrowLeftRight"
    max-width="648px"
  >
    <template #body>
      <v-card flat>
        <v-card-text>
          <fade-transition>
            <p v-if="showEmptyText">
              empty
            </p>
            <div v-else-if="transactionList !== null">
              <v-card-text>
                The following transactions transfer value between the two clusters.
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
                  <router-link
                    :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: item.txhash }}"
                    target="_blank"
                  >
                    <span class="shorten">{{ item.txhash }}</span>
                  </router-link>
                </template>
                <template #item.privacytype="{ item }">
                  <span>{{ getPrivacyTypeLabel(item.privacytype) }}</span>
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
              <v-card-text>
                Outputs which connect the two nodes are marked in red.
              </v-card-text>
              <!-- duplicate transaction hashes can exist -> loop through all results
               (e.g. d5d27987d2a3dfc724e359870c6644b40e497bdc0589a033220fe15429d88599 in Bitcoin) -->
              <template
                v-for="t in transactions"
                :key="t.txhash+t.bid"
              >
                <transaction
                  :show-fingerprint-link="true"
                  :show-heuristic-editor-link="false"
                  :tx="t"
                  show-details
                  :embed="false"
                  :show-title-bar="false"
                  :highlight-transaction="connectionTarget.transactionHash"
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
import {mdiArrowLeftRight} from '@mdi/js';
import SideBar from '@/components/common/SideBar.vue';
import {
	computed, inject, onUpdated, ref,
} from 'vue';
import {useMsgStore} from '@/pinia/msg.js';
import {useRoute} from 'vue-router';
import {
	WORKSPACE_NODE_TYPE_HEURISTIC, WORKSPACE_NODE_TYPE_CLUSTER,
	ROUTE_NAME_TRANSACTION_PAGE, WORKSPACE_NODE_TYPE_TRANSACTION,
} from '@/constants/index.js';
import {convertAmount, getPrivacyTypeLabel} from '../../utilities/index.js';
import Transaction from '@/components/explorer/transaction/Transaction.vue';
import FadeTransition from '@/components/common/FadeTransition.vue';

const props = defineProps({
	connection: {type: Object, required: true},
	workspaceUid: {type: String, required: true},
});

const model = defineModel({type: Boolean});
const route = useRoute();
const msgStore = useMsgStore();
const dakar = inject('dakar');

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

const headers = [
	{
		title: 'Hash', key: 'txhash', align: 'start', sortable: false,
	},
	{title: 'Privacy Type', key: 'privacytype'},
	{title: 'Timestamp', key: 'ts'},
	{title: 'Fee', key: 'fee'},
	{title: 'Input Amount', key: 'inputAmount'},
	{title: 'Output Amount', key: 'outputAmount'},
];

// Hooks

// eslint-disable-next-line complexity
onUpdated(async () => {
	if (props.connection?.target?.uid && props.connection.source?.uid) {
		const sourceUID = props.connection.source.uid;
		const targetUID = props.connection.target.uid;

		if (oldConnection && (sourceUID === oldConnection.source.uid && targetUID === oldConnection.target.uid)) {
			return;
		}

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
      || (connectionSource.value.type === WORKSPACE_NODE_TYPE_HEURISTIC
        && connectionTarget.value.type === WORKSPACE_NODE_TYPE_CLUSTER)
     || (connectionSource.value.type === WORKSPACE_NODE_TYPE_CLUSTER
        && connectionTarget.value.type === WORKSPACE_NODE_TYPE_HEURISTIC)
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
	if (transactionList.value !== null) {
		return 'Connection List';
	}

	if (transactions.value !== null && transactions.value[0]?.txhash) {
		return `Transaction ${transactions.value[0].txhash}`;
	}

	return 'Connections';
});

// Functions

function setErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: true, category: route.name,
	});
}

async function getConnectionData() {
	if (!connectionSource.value?.uid || !connectionTarget.value?.uid || !props.workspaceUid) {
		return;
	}

	transactionList.value = null;
	transactions.value = null;

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
			let hasPrivacyType = false;
			transactionList.value = response.amountTransactions.map(d => {
				if (d.privacytype !== undefined) {
					hasPrivacyType = true;
				}

				d.ts = new Date(d.ts);
				return d;
			});

			if (hasPrivacyType) {
				filteredHeaders.value = headers;
			} else {
				// Data has no privacy type, so remove it from header
				filteredHeaders.value = headers.filter(d => d.key !== 'privacytype');
			}
		} else if (response.frontendTransactions) {
			transactions.value = response.frontendTransactions;
		} else {
			transactionList.value = [];
		}
	} catch (e) {
		setErrorMessage(e);
	}
}

async function getTransactionData() {
	if (!connectionSource.value?.transactionHash) {
		return;
	}

	transactionList.value = null;
	transactions.value = null;

	try {
		const response = await dakar.data.blockchainTransactionsHashGet(
			{hash: connectionSource.value.transactionHash});

		if (response.transactions) {
			transactions.value = response.transactions;
		}
	} catch (e) {
		setErrorMessage(e);
	}
}

</script>

<style scoped>
.shorten {
  display: block;
  max-width: 125px;
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}
</style>
