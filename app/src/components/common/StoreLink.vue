<template>
  <router-link
    :to="to"
    @click="routing"
  />
  <router-link
    v-slot="{href, navigate}"
    :to="to"
    :custom="true"
    :target="target"
  >
    <a
      :href="href"
      @click="onLinkClick($event, navigate)"
    >
      <slot />
    </a>
  </router-link>
</template>
<script setup>

import {ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants/index.js';

const props = defineProps({
	to: {type: Object, required: true},
	target: {type: String, required: false, default: undefined},
});
import {useWorkspaceStore} from '@/pinia/workspace.js';

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
