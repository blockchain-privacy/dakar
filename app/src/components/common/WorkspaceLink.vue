<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div
    class="me-1 align-center d-flex shorten"
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
    <!-- The order of props is important. v-bind has to come before 'custom',
    because 'custom' needs to overwrite 'custom' the props passed in v-bind.
    See more here: https://www.vueframework.com/guide/migration/v-bind.html -->
    <router-link
      v-slot="{href, navigate}"
      v-bind="$props"
      custom
      :class="$attrs.class"
      :style="$attrs.style"
    >
      <a
        class="shorten"
        v-bind="$attrs"
        :href="href"
        @click="onLinkClick($event, navigate)"
      >
        <slot />
      </a>
    </router-link>
  </div>
</template>

<script setup>
import {RouterLink} from 'vue-router';
import {
	computed,
	onMounted,
	ref,
	watch,
} from 'vue';
import {
	ROUTE_NAME_ADDRESS_PAGE,
	ROUTE_NAME_TRANSACTION_PAGE,
	WORKSPACE_NODE_TYPE_TRANSACTION,
	WORKSPACE_NODE_TYPE_CLUSTER,
} from '@/constants/index.js';
import {useWorkspaceStore} from '@/pinia/workspace.js';
import '@/assets/main.css';

defineOptions({
	inheritAttrs: false,
});

const props = defineProps({
	...RouterLink.props,
	disableSelect: {type: Boolean, required: false},
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
