<template>
  <v-list>
    <v-list-item
      v-for="tx in getLimitedItems"
      :key="tx.txhash"
      :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
             params: { id: tx.txhash }}"
    >
      <v-list-item-title>
        {{ tx.txhash }}
        <div v-if="tx.destinationCount">
          Destinations: {{ tx.destinationCount }}
        </div>
      </v-list-item-title>
    </v-list-item>
    <v-expand-transition>
      <div v-if="showAllOutputs">
        <v-list-item
          v-for="tx in getResidualItems"
          :key="tx.txhash"
          :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
                 params: { id: tx.txhash }}"
        >
          <v-list-item-title>
            {{ tx.txhash }}
            <div v-if="tx.destinationCount">
              Destinations: {{ tx.destinationCount }}
            </div>
          </v-list-item-title>
        </v-list-item>
      </div>
    </v-expand-transition>
  </v-list>
  <v-btn
    v-if="areItemsLimited"
    variant="text"
    :rounded="false"
    :block="true"
    size="small"
    @click="showAllOutputs = !showAllOutputs"
  >
    {{ items.length - maxItems }} additional
    {{ plural('transaction', items.length - maxItems) }}
    <v-icon>{{ showAllOutputs ? mdiChevronUp : mdiChevronDown }}</v-icon>
  </v-btn>
</template>

<script setup>
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import {mdiChevronDown, mdiChevronUp} from '@mdi/js';
import {plural} from '@/utilities';
import {computed, ref} from 'vue';

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
