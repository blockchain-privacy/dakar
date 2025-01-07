<template>
  <v-row>
    <v-col>
      <v-select
        v-model="model.order"
        :disabled="loading"
        :loading="loading?'primary':false"
        style="max-width: 300px; min-width: 200px;"
        :items="sortItems"
        item-value="id"
        item-title="text"
        label="Sort"
        @update:model-value="handleSortAndFilter"
      >
        <template #item="item">
          <v-divider v-if="item.item?.raw?.divider !== undefined" />
          <v-list-item
            v-else
            v-bind="item.props"
            :value="item.item.raw.id"
            :title="item.item?.raw?.text"
          />
        </template>
      </v-select>
    </v-col>
    <v-col>
      <v-select
        v-model="model.filter"
        :disabled="loading"
        :loading="loading?'primary':false"
        style="max-width: 300px; min-width: 200px;"
        :items="filterItems"
        item-value="id"
        item-title="text"
        label="Filter"
        multiple
        @update:model-value="handleSortAndFilter"
      >
        <template #selection="{ item }">
          <v-chip size="small">
            <span>{{ item.raw.chip }}</span>
          </v-chip>
        </template>
      </v-select>
    </v-col>
    <v-col v-if="isSortingByInput">
      <v-alert
        type="info"
        variant="text"
      >
        Only spent outputs are shown
      </v-alert>
    </v-col>
  </v-row>
</template>

<script setup>
import {computed, onMounted, ref} from 'vue';

const props = defineProps({
	loading: {type: Boolean, required: false},
	outputCount: {type: Number, required: true},
	inputCount: {type: Number, required: true},
	coinbaseCount: {type: Number, required: true},
});
const model = defineModel({type: Object});
const emit = defineEmits(['update:modelValue', 'change']);

const sortItems = ref([
	{id: 0, text: 'Ascending by output date', disabled: false},
	{id: 1, text: 'Descending by output date', disabled: false},
	{divider: true},
	{id: 2, text: 'Ascending by input date', disabled: false},
	{id: 3, text: 'Descending by input date', disabled: false},
	{divider: true},
	{id: 4, text: 'Ascending by amount', disabled: false},
	{id: 5, text: 'Descending by amount', disabled: false},
]);

const filterItems = ref([
	{
		id: 0, text: 'Only show coinbase outputs', chip: 'Coinbase outputs', disabled: false,
	},
	{
		id: 1, text: 'Only show unspent outputs', chip: 'Unspent outputs', disabled: false,
	},
]);

// Computed

const isSortingByInput = computed(() => model.value.order === 2 || model.value.order === 3);

// Hooks
onMounted(() => {
	updateSortState();
	updateFilterState();
});

// Functions
function handleSortAndFilter() {
	updateSortState();
	updateFilterState();
	emit('change');
}

function updateSortState() {
	let isUnspentFilterSelected = false;

	if (props.inputCount === 0) {
		isUnspentFilterSelected = true;
	} else {
		for (const s of model.value.filter) {
			if (s === 1) {
				isUnspentFilterSelected = true;
				break;
			}
		}
	}

	sortItems.value.forEach(d => {
		if (d.id === 2 || d.id === 3) {
			d.disabled = isUnspentFilterSelected;
		}
	});
}

function updateFilterState() {
	let disableUnspentFilter = false;
	if (props.outputCount - props.inputCount === 0) {
		disableUnspentFilter = true;
	} else {
		const selection = model.value.order;
		if (selection === 2 || selection === 3) {
			disableUnspentFilter = true;
		}
	}

	let disableCoinbaseFilter = false;
	if (props.coinbaseCount === 0 || props.coinbaseCount === props.outputCount) {
		disableCoinbaseFilter = true;
	}

	filterItems.value.forEach(d => {
		if (d.id === 0) {
			d.disabled = disableCoinbaseFilter;
		} else if (d.id === 1) {
			d.disabled = disableUnspentFilter;
		}
	});
}

</script>

<style scoped>

</style>
