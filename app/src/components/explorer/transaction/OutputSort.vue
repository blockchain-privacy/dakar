<template>
  <v-card
    min-width="250px"
    max-width="400px"
  >
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
        label="Filter"
        mandatory
        @update:model-value="handleModelUpdate"
      />
    </v-card-text>
  </v-card>
</template>

<script setup>
import SortSelect from '@/components/common/SortSelect.vue';
import {ref} from 'vue';
import ChipFilter from '@/components/explorer/address/ChipFilter.vue';

const props = defineProps({
	transactionTypes: {type: Array, required: true},
});

const sortItems = ['Amount', 'Time'];
const sortValue = ref(sortItems[1]); // Sort by time by default
const sortDescending = ref(false); // Sort by ascending by default

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
