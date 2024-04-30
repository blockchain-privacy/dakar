<template>
  <side-bar
    v-model="model"
    title="Connections"
    :icon="mdiArrowLeftRight"
    max-width="648px"
  >
    <template #body>
      <v-card flat>
        <v-card-text>
          <p v-if="showEmptyText">
            empty
          </p>
          <v-data-table
            v-else
            v-model:sort-by="identitiesSortBy"
            :headers="headers"
            :items="transactions? transactions: []"
            item-key="txhash"
            :loading="!transactions"
            items-per-page="50"
          >
            <template #item.txhash="{ item }">
              <router-link
                :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: item.txhash }}"
                target="_blank"
              >
                <span
                  class="shorten"
                >{{ item.txhash }}</span>
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
        </v-card-text>
      </v-card>
    </template>
  </side-bar>
</template>

<script setup>
import {mdiArrowLeftRight} from '@mdi/js';
import SideBar from '@/components/common/SideBar.vue';
import {inject, onUpdated, ref} from 'vue';
import {useMsgStore} from '@/pinia/msg.js';
import {useRoute} from 'vue-router';
import {WORKSPACE_NODE_TYPE_HEURISTIC, WORKSPACE_NODE_TYPE_CLUSTER, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants/index.js';
import {convertAmount, getPrivacyTypeLabel} from '../../utilities/index.js';

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
const transactions = ref(null);
const showEmptyText = ref(false);
const identitiesSortBy = ref([{key: 'ts', order: 'desc'}]);

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
		if ((connectionSource.value.type === WORKSPACE_NODE_TYPE_CLUSTER
        && connectionTarget.value.type === WORKSPACE_NODE_TYPE_CLUSTER)
      || (connectionSource.value.type === WORKSPACE_NODE_TYPE_HEURISTIC
        && connectionTarget.value.type === WORKSPACE_NODE_TYPE_CLUSTER)
     || (connectionSource.value.type === WORKSPACE_NODE_TYPE_CLUSTER
        && connectionTarget.value.type === WORKSPACE_NODE_TYPE_HEURISTIC)) {
			await getConnectionData();
		} else {
			showEmptyText.value = true;
		}
	}
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

		transactions.value = response.transactions.map(d => {
			d.ts = new Date(d.ts);
			return d;
		});
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
