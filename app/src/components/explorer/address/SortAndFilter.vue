<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-row>
    <v-col>
      <sort-select
        v-model:sort="sort"
        v-model:direction="direction"
        :items="sortItems"
        style="max-width: 300px; min-width: 200px;"
        @update:sort="handleSortAndFilter"
        @update:direction="handleSortAndFilter"
      />
    </v-col>
    <v-col>
      <v-select
        v-model="filter"
        :items="filterItems"
        label="Filter"
        multiple
        style="max-width: 300px; min-width: 200px;"
        @update:model-value="handleSortAndFilter"
      >
        <template #selection="{ item }">
          <v-chip size="small">
            <span>{{ item.chip }}</span>
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
import SortSelect from '@/components/common/SortSelect.vue';

const props = defineProps({
	outputCount: {type: Number, required: true},
	inputCount: {type: Number, required: true},
});

const sort = defineModel('sort', {type: Object});
const direction = defineModel('direction', {type: Boolean});
const filter = defineModel('filter', {type: Array});

const sortItems = ref([
	{value: 0, title: 'Output date'},
	{value: 1, title: 'Input date'},
	{value: 2, title: 'Amount'},
]);

const filterItems = ref([
	{value: 0, title: 'Only show coinbase outputs', chip: 'Coinbase outputs'},
	{value: 1, title: 'Only show unspent outputs', chip: 'Unspent outputs'},
]);

// Computed
const isSortingByInput = computed(() => sort.value.value === 1);

// Hooks
onMounted(() => {
	updateSortState();
	updateFilterState();
});

// Functions
function handleSortAndFilter() {
	updateSortState();
	updateFilterState();
}

function updateSortState() {
	let isUnspentFilterSelected = false;

	if (props.inputCount === 0) {
		isUnspentFilterSelected = true;
	} else {
		for (const s of filter.value) {
			if (s === 1) {
				isUnspentFilterSelected = true;
				break;
			}
		}
	}

	sortItems.value.forEach(d => {
		if (d.value === 1) {
			d.props ??= {};
			d.props.disabled = isUnspentFilterSelected;
		}
	});
}

function updateFilterState() {
	let disableUnspentFilter = false;
	if (props.outputCount - props.inputCount === 0 || sort.value.value === 1) {
		disableUnspentFilter = true;
	}

	let disableCoinbaseFilter = false;
	if (props.coinbaseCount === props.outputCount) {
		disableCoinbaseFilter = true;
	}

	filterItems.value.forEach(d => {
		if (d.value === 0) {
			d.props ??= {};
			d.props.disabled = disableCoinbaseFilter;
		} else if (d.value === 1) {
			d.props ??= {};
			d.props.disabled = disableUnspentFilter;
		}
	});
}

</script>

<style scoped>

</style>
