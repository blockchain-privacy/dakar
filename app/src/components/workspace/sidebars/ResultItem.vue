<template>
  <v-list>
    <store-link
      v-for="tx in getLimitedItems"
      :key="tx.txhash"
      :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
             params: { id: tx.txhash }}"
    >
      {{ tx.txhash }}
    </store-link>
    <v-expand-transition>
      <div v-if="showAllOutputs">
        <store-link
          v-for="tx in getResidualItems"
          :key="tx.txhash"
          :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
                 params: { id: tx.txhash }}"
        >
          {{ tx.txhash }}
        </store-link>
      </div>
    </v-expand-transition>
  </v-list>
  <v-btn
    v-if="areItemsLimited"
    variant="text"
    :rounded="false"
    block
    size="small"
    @click="showAllOutputs = !showAllOutputs"
  >
    {{ items.length - maxItems }} additional
    {{ plural('transaction', items.length - maxItems) }}
    <v-icon>{{ showAllOutputs ? mdiChevronUp : mdiChevronDown }}</v-icon>
  </v-btn>
</template>

<script setup>
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants/index.js';
import {mdiChevronDown, mdiChevronUp} from '@mdi/js';
import {plural} from '@/utilities/index.js';
import {computed, ref} from 'vue';
import StoreLink from '@/components/common/WorkspaceLink.vue';

const props = defineProps({
	items: {type: Array, required: true},
	maxItems: {type: Number, required: true},
});

const showAllOutputs = ref(false);

// Computed
const areItemsLimited = computed(() => props.items.length > props.maxItems);
const getLimitedItems = computed(() => {
	if (!props.items) {
		return [];
	}

	return props.items.slice(0, props.maxItems);
});

const getResidualItems = computed(() => {
	if (!props.items) {
		return [];
	}

	if (props.items.length <= props.maxItems) {
		return [];
	}

	if (showAllOutputs.value) {
		return props.items.slice(props.maxItems);
	}

	return [];
});

</script>

<style scoped>

</style>
