<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <side-bar
    v-model="model"
    :title="`Fingerprint ${transactionHash}`"
    :icon="mdiFingerprint"
    max-width="648px"
  >
    <template #actions>
      <add-nodes-chip
        v-if="selectableEntities.size > 0"
        :disabled="disableAddingNodes"
        :show-select-all-transactions="showSelectTransactions"
        @add-nodes="emitAddNodes"
        @select-all-transactions="selectAllTransactions"
        @deselect-all-transactions="deselectAllTransactions"
      />
    </template>
    <template #body>
      <v-card-text>
        <fingerprint-transactions
          v-if="transactionHash"
          :transaction-hash="transactionHash"
          @received-transactions="receivedTransactions"
        />
      </v-card-text>
    </template>
  </side-bar>
</template>

<script setup>
import SideBar from '@/components/common/SideBar.vue';
import {mdiFingerprint} from '@mdi/js';
import FingerprintTransactions from '@/components/explorer/transaction/FingerprintTransactions.vue';
import AddNodesChip from '@/components/workspace/sidebars/AddNodesChip.vue';
import {ref} from 'vue';
import {WORKSPACE_NODE_TYPE_TRANSACTION} from '@/constants/index.js';
import {useWorkspaceStore} from '@/pinia/workspace.js';

const emit = defineEmits(['addNodes']);
const model = defineModel({type: Boolean});
const workspaceStore = useWorkspaceStore();

defineProps({
	transactionHash: {type: String, required: true},
	disableAddingNodes: {type: Boolean, required: true},
});

const showSelectTransactions = ref(true);

// Holds all transactions which can be selected and added to the workspace
const selectableEntities = ref(new Map());

// Function
function selectAllTransactions() {
	workspaceStore.setWorkspaceNodes([...selectableEntities.value.values()]
		.filter(d => d.type === WORKSPACE_NODE_TYPE_TRANSACTION));
}

function deselectAllTransactions() {
	workspaceStore.removeNodesFromMap([...workspaceStore.workspaceNodes.values()]
		.filter(d => d.type === WORKSPACE_NODE_TYPE_TRANSACTION)
		.map(d => d.id));
}

function emitAddNodes(nodes) {
	emit('addNodes', nodes);
	model.value = false;
}

function receivedTransactions(transactions) {
	selectableEntities.value.clear();
	transactions.forEach(d => {
		selectableEntities.value.set(d, {id: d, type: WORKSPACE_NODE_TYPE_TRANSACTION});
	});
}

</script>

<style scoped>
.textBorder {
  border-style: solid;
  border-radius: 5px;
  border-width: 1px;
}
</style>
