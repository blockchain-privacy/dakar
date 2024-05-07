<template>
  <v-chip
    v-if="querySelectionCount"
    :rounded="true"
    class="me-2"
    color="green"
    variant="tonal"
    :prepend-icon="mdiPlus"
    @click="emitAddNodes"
  >
    Add {{ querySelectionCount }} {{ plural('element',querySelectionCount) }}
  </v-chip>
</template>
<script setup>
import {mdiPlus} from '@mdi/js';
import {plural} from '@/utilities/index.js';
import {ref, watch} from 'vue';
import {useWorkspaceStore} from '@/pinia/workspace.js';

const querySelectionCount = ref(0);
const emit = defineEmits(['addNodes']);
const workspaceStore = useWorkspaceStore();

watch(
	() => workspaceStore.workspaceNodes,
	_ => {
		querySelectionCount.value = workspaceStore.workspaceNodes.size;
	},
	{deep: true}, // Deep watch necessary for Set
);

function emitAddNodes() {
	emit('addNodes', [...workspaceStore.workspaceNodes]);
	workspaceStore.workspaceNodes.clear();
}

</script>

<style scoped>

</style>
