<template>
  <div
    class="d-flex float-left me-1 align-center"
    style="max-width: 100%"
  >
    <v-checkbox
      v-if="!disableSelect && isWorkspaceMode"
      v-model="checkBoxModel"
      hide-details
      density="compact"
      class="flex-shrink-0"
      @update:model-value="checkBoxChanged"
    />
    <router-link
      v-slot="{href, navigate}"
      custom
      v-bind="$props"
      :class="$attrs.class"
      :style="$attrs.style"
    >
      <a
        :href="href"
        v-bind="$attrs"
        @click="onLinkClick($event, navigate)"
      >
        <slot />
      </a>
    </router-link>
  </div>
</template>

<script setup>
import {
	ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_TRANSACTION_PAGE, WORKSPACE_NODE_TYPE_TRANSACTION, WORKSPACE_NODE_TYPE_CLUSTER,
} from '@/constants/index.js';
import {RouterLink} from 'vue-router';
import {useWorkspaceStore} from '@/pinia/workspace.js';
import '@/assets/main.css';
import {
	computed, onMounted, ref, watch,
} from 'vue';

defineOptions({
	inheritAttrs: false,
});

const props = defineProps({
	...RouterLink.props,
	disableSelect: {type: Boolean, required: false, default: false},
});
const emit = defineEmits(['clicked']);

const workspaceStore = useWorkspaceStore();
const checkBoxModel = ref(false);

// Computed
const isWorkspaceMode = computed(() => workspaceStore.getIsWorkspaceActive
  && (props.to.name === ROUTE_NAME_TRANSACTION_PAGE || props.to.name === ROUTE_NAME_ADDRESS_PAGE)
  && Boolean(props.to.params?.id));

// Watchers
// keep state of checkbox in sync with store

watch(
	() => workspaceStore.workspaceNodes,
	_ => {
		if (isWorkspaceMode.value) {
			checkBoxModel.value = workspaceStore.workspaceNodes.has(props.to.params.id);
		}
	},
	{deep: true}, // Deep watch necessary for Set
);

// Hooks

onMounted(() => {
	checkBoxModel.value = workspaceStore.workspaceNodes.has(props.to.params.id);
});

// Functions
function onLinkClick(e, navigate) {
	if (isWorkspaceMode.value) {
		e.preventDefault();
		workspaceStore.setWorkspaceNode({to: props.to, id: props.to.params.id});
		emit('clicked');
		return;
	}

	navigate(e);
}

function checkBoxChanged(val) {
	if (val) {
		let type = WORKSPACE_NODE_TYPE_TRANSACTION;
		if (props.to.name === ROUTE_NAME_ADDRESS_PAGE) {
			type = WORKSPACE_NODE_TYPE_CLUSTER;
		}

		workspaceStore.addNodeToMap({id: props.to.params.id, type});
	} else {
		workspaceStore.removeNodeFromMap(props.to.params.id);
	}
}

</script>

<style scoped>

</style>
