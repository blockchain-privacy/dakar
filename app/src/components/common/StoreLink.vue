<template>
  <div class="d-flex float-left me-1 align-center">
    <v-checkbox
      v-if="isWorkspaceMode"
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
import {ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants/index.js';
import {RouterLink} from 'vue-router';
import {useWorkspaceStore} from '@/pinia/workspace.js';
import '@/assets/main.css';
import {
	computed, ref, watch,
} from 'vue';

defineOptions({
	inheritAttrs: false,
});

const props = defineProps({
	...RouterLink.props,
});

const workspaceStore = useWorkspaceStore();
const checkBoxModel = ref(false);

// Computed
const isWorkspaceMode = computed(() => workspaceStore.getIsWorkspaceActive
  && (props.to.name === ROUTE_NAME_TRANSACTION_PAGE || props.to.name === ROUTE_NAME_ADDRESS_PAGE)
  && props.to.params?.id);

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

// Functions
function onLinkClick(e, navigate) {
	if (isWorkspaceMode.value) {
		e.preventDefault();
		workspaceStore.setWorkspaceNode({to: props.to, id: props.to.params.id});
		return;
	}

	navigate(e);
}

function checkBoxChanged(val) {
	if (isWorkspaceMode.value) {
		if (val) {
			workspaceStore.addNodeToSet(props.to.params.id);
		} else {
			workspaceStore.removeNodeFromSet(props.to.params.id);
		}
	}
}

</script>

<style scoped>

</style>
