<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-menu
    v-model="menuModel"
    :close-on-content-click="false"
    eager
  >
    <template #activator="activator">
      <v-btn
        variant="text"
        icon
        v-bind="activator.props"
      >
        <v-badge
          v-if="modified"
          color="success"
          dot
          location="bottom right"
        >
          <v-icon :icon="mdiFilter" />
        </v-badge>
        <v-icon
          v-else
          :icon="mdiFilter"
        />
      </v-btn>
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
import {computed, ref} from 'vue';
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

// Computed
const modified = computed(() => sortDescending.value || sortValue.value.value !== sortItems[1].value
	|| chipFilterModel.value.length < props.transactionTypes.length);

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
