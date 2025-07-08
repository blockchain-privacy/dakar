<template>
  <v-menu
    v-model="menuModel"
    :close-on-content-click="false"
    eager
  >
    <template #activator="activator">
      <v-btn
        :icon="mdiFilter"
        variant="text"
        v-bind="activator.props"
      />
    </template>
    <v-card width="300px">
      <v-card-text>
        <sort-select
          v-model:sort="sortValue"
          v-model:direction="sortDescending"
          :items="sortItems"
          @update:sort="handleModelUpdate"
          @update:direction="handleModelUpdate"
        />
        <chip-filter
          v-model="chipFilterModel"
          class="mt-2"
          :items="transactionTypes"
          label="Filter by type"
          mandatory
          @update:model-value="handleModelUpdate"
        />
      </v-card-text>
      <v-card-actions>
        <v-btn
          class="ms-auto"
          @click="menuModel = false"
        >
          Close
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>

<script setup>
import SortSelect from '@/components/common/SortSelect.vue';
import {ref} from 'vue';
import ChipFilter from '@/components/explorer/address/ChipFilter.vue';
import {mdiFilter} from '@mdi/js';

const props = defineProps({
	transactionTypes: {type: Array, required: true},
});

const sortItems = [{value: 'amount', title: 'Amount'}, {value: 'time', title: 'Time'}, {value: 'txtype', title: 'Transaction type'}];
const sortValue = ref(sortItems[1]); // Sort by time by default
const sortDescending = ref(false); // Sort by ascending by default
const menuModel = ref(false);

const chipFilterModel = ref([...props.transactionTypes.keys()]);
const model = defineModel({type: Object});

// Functions
function handleModelUpdate() {
	model.value = {
		sortValue: sortValue.value,
		sortDescending: sortDescending.value,
		// Map index to string keys
		filter: chipFilterModel.value.map(v => props.transactionTypes[v].text),
	};
}
</script>

<style scoped>

</style>
