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
          <transaction-item
            v-for="(tx) in transactions"
            v-else
            :key="tx.txhash"
            :tx="tx"
            class="mx-auto mt-3"
            max-width="1000"
          />
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
import {WORKSPACE_NODE_TYPE_HEURISTIC, WORKSPACE_NODE_TYPE_CLUSTER} from '@/constants/index.js';
import TransactionItem from '@/components/common/TransactionItem.vue';

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
const transactions = ref([]);
const showEmptyText = ref(false);

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

	transactions.value = [];

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
		transactions.value = response.transactions;
	} catch (e) {
		setErrorMessage(e);
	}
}

</script>

<style scoped>

</style>
