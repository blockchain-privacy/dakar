<template>
  <router-link
    v-slot="{href, navigate}"
    custom
    v-bind="$props"
    :class="$attrs.class"
  >
    <a
      :href="href"
      v-bind="$attrs"
      @click="onLinkClick($event, navigate)"
    >
      <slot />
    </a>
  </router-link>
</template>

<script setup>
import {ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants/index.js';
import {RouterLink} from 'vue-router';
import {useWorkspaceStore} from '@/pinia/workspace.js';
import '@/assets/main.css';

defineOptions({
	inheritAttrs: false,
});

const props = defineProps({
	...RouterLink.props,
});

const workspaceStore = useWorkspaceStore();

// Function

function onLinkClick(e, navigate) {
	if (workspaceStore.getIsWorkspaceActive && (props.to.name === ROUTE_NAME_TRANSACTION_PAGE || props.to.name === ROUTE_NAME_ADDRESS_PAGE)
    && props.to.params?.id) {
		e.preventDefault();
		workspaceStore.setWorkspaceNode({to: props.to, id: props.to.params.id});
		return;
	}

	navigate(e);
}

</script>

<style scoped>

</style>
