<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-chip
    v-if="showSelectAllTransactions"
    rounded
    class="me-2"
    color="primary"
    variant="tonal"
    :prepend-icon="mdiCheckboxMultipleOutline"
    @click="emitSelectAllTransactions"
  >
    {{ transactionLabel }}
  </v-chip>
  <v-chip
    v-if="showSelectAllAddresses"
    rounded
    class="me-2"
    color="primary"
    variant="tonal"
    :prepend-icon="mdiCheckboxMultipleOutline"
    @click="emitSelectAllAddresses"
  >
    {{ addressLabel }}
  </v-chip>
  <fade-transition>
    <v-chip
      v-if="selectionCount"
      rounded
      class="me-2"
      color="primary"
      variant="tonal"
      :prepend-icon="mdiPlus"
      :disabled="disabled"
      closable
      @click:close="deselectAllNodes"
      @click="emitAddNodes"
    >
      <div class="d-flex align-center">
        Add Entities
        <v-badge
          inline
          color="primary"
          :content="selectionCount"
        />
      </div>
    </v-chip>
  </fade-transition>
</template>
<script setup>
import {mdiCheckboxMultipleOutline, mdiPlus} from '@mdi/js';
import {computed, ref, watch} from 'vue';
import {useWorkspaceStore} from '@/pinia/workspace.js';
import FadeTransition from '@/components/common/FadeTransition.vue';
import {WORKSPACE_NODE_TYPE_CLUSTER, WORKSPACE_NODE_TYPE_TRANSACTION} from '@/constants/index.js';

const selectionCount = ref(0);
const addressCount = ref(0);
const transactionCount = ref(0);
const emit = defineEmits(['addNodes',
	'selectAllTransactions',
	'deselectAllTransactions',
	'selectAllAddresses',
	'deselectAllAddresses']);
const workspaceStore = useWorkspaceStore();

defineProps({
	disabled: {type: Boolean, required: false},
	showSelectAllTransactions: {type: Boolean, required: false},
	showSelectAllAddresses: {type: Boolean, required: false},
});

const transactionLabel = computed(() => transactionCount.value ? 'Deselect all Transactions' : 'Select all Transactions');
const addressLabel = computed(() => addressCount.value ? 'Deselect all Addresses' : 'Select all Addresses');

watch(
	() => workspaceStore.workspaceNodes,
	_ => {
		selectionCount.value = workspaceStore.workspaceNodes.size;

		let numAddresses = 0;
		let numTransactions = 0;
		workspaceStore.workspaceNodes.forEach(d => {
			if (d.type === WORKSPACE_NODE_TYPE_CLUSTER) {
				numAddresses += 1;
			} else if (d.type === WORKSPACE_NODE_TYPE_TRANSACTION) {
				numTransactions += 1;
			}
		});

		addressCount.value = numAddresses;
		transactionCount.value = numTransactions;
	},
	{deep: true}, // Deep watch necessary for Set
);

function emitAddNodes() {
	emit('addNodes', [...workspaceStore.workspaceNodes.values()].map(d => d.id));
	workspaceStore.workspaceNodes.clear();
}

function emitSelectAllTransactions() {
	if (transactionCount.value) {
		emit('deselectAllTransactions');
	} else {
		emit('selectAllTransactions');
	}
}

function emitSelectAllAddresses() {
	if (addressCount.value) {
		emit('deselectAllAddresses');
	} else {
		emit('selectAllAddresses');
	}
}

function deselectAllNodes(e) {
	e.stopPropagation();
	workspaceStore.workspaceNodes.clear();
}

</script>

<style scoped>

</style>
