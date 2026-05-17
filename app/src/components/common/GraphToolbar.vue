<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="d-flex align-center justify-space-between">
    <div
      v-if="name"
      class="d-flex align-center d-inline-block"
    >
      <div class="position-relative">
        <v-icon
          class="mx-3"
          icon="$graphIcon"
          size="32"
        />
        <v-icon
          class="position-absolute"
          :icon="BLOCKCHAIN_ATTRIBUTES[route.params.blockchainMode].icon"
          :color="BLOCKCHAIN_ATTRIBUTES[route.params.blockchainMode].color"
          size="20"
          style="left: calc(100% - 20px); bottom: -5px"
        />
      </div>
      <p
        v-tooltip="{'text': 'Name of the Workspace', 'location':'top', 'open-delay': 400}"
        class="text-title-large my-2 workspace-name"
      >
        {{ name }}
      </p>
    </div>
    <v-btn-toggle
      v-if="!oneLine"
      v-model="selectionToggle"
      color="primary"
      class="ms-2"
      rounded="0"
      mandatory
      @update:model-value="onSelectionModeChanged"
    >
      <v-btn
        v-tooltip="{'text': 'Select', 'location':'top', 'open-delay': 400}"
        :icon="mdiSelect"
      />
      <v-btn
        v-tooltip="{'text': 'Drag', 'location':'top', 'open-delay': 400}"
        :icon="mdiCursorPointer"
      />
    </v-btn-toggle>
  </div>
  <div class="d-flex justify-center align-center flex-wrap">
    <v-btn
      v-if="!disableFilter"
      variant="text"
      class="my-1"
      :active="showFilter"
      @click="showFilter = !showFilter"
    >
      <v-icon
        start
        :icon="mdiCog"
      />
      Filter Nodes
    </v-btn>
    <v-btn
      variant="text"
      class="my-1"
      @click="emit('rearrange')"
    >
      <v-icon
        :icon="mdiCached"
        start
      />
      Rearrange
    </v-btn>
    <v-btn
      variant="text"
      class="my-1"
      @click="emit('center')"
    >
      <v-icon
        :icon="mdiImageFilterCenterFocus"
        start
      />
      Center
    </v-btn>
  </div>
  <div class="d-flex justify-center align-center flex-wrap">
    <v-btn
      v-if="showAddSelectorButton"
      variant="text"
      class="my-1"
      @click="onAddSelector"
    >
      <v-icon
        :icon="mdiFilterPlus"
        start
      />
      Add Property Selector
    </v-btn>
    <v-btn
      v-if="showSearchButton"
      variant="text"
      :disabled="!addEntityEnabled"
      @click="queryDialogModel = true"
    >
      <v-tooltip
        activator="parent"
        location="top"
        open-delay="400"
      >
        <div class="d-flex align-center">
          <v-hotkey keys="cmd+k" />
        </div>
      </v-tooltip>
      <v-icon
        :icon="mdiPlus"
        start
      />
      Add Entities
    </v-btn>
    <v-btn
      v-if="showDeleteButton && selectedItemCount > 0"
      v-tooltip="{'text': 'Delete Nodes', 'location':'top', 'open-delay': 400}"
      :disabled="deleteDisabled"
      variant="flat"
      class="my-1 me-1"
      @click="emit('deleteSelected')"
    >
      <v-icon
        :icon="mdiDelete"
        start
      />
      {{ selectedItemCount }}
    </v-btn>
    <v-btn
      v-if="shortestPathEnabled"
      variant="flat"
      class="my-1 me-1"
      @click="emit('shortestPathLookup')"
    >
      <v-icon
        :icon="mdiChartTimelineVariant"
        start
      />
      Shortest path
    </v-btn>
    <v-btn
      v-if="heuristicBatchEnabled && selectedItemCount > 0"
      v-tooltip="{'text': 'Add Multiple CoinJoin Heuristics', 'location':'top', 'open-delay': 400}"
      variant="flat"
      class="my-1 me-1"
      @click="emit('heuristicBatch')"
    >
      <v-icon
        :icon="blenderPlus"
        start
      />
      {{ selectedItemCount }}
    </v-btn>
    <v-btn-toggle
      v-if="oneLine"
      v-model="selectionToggle"
      color="primary"
      class="ms-2"
      rounded="0"
      mandatory
      @update:model-value="onSelectionModeChanged"
    >
      <v-btn :icon="mdiSelect" />
      <v-btn :icon="mdiCursorPointer" />
    </v-btn-toggle>
  </div>
  <v-expand-transition>
    <div v-if="!disableFilter && showFilter">
      <div class="d-flex justify-center">
        <chip-filter
          v-model="nodeFilters"
          style="max-width: 420px"
          mandatory
          label="Node Types"
          :items="nodeTypeItems"
          @changed="onFilterChanged"
        />
      </div>
      <div class="d-flex justify-center">
        <chip-filter
          v-model="typeFilters"
          label="Transaction Types"
          :items="transactionTypeItems"
          @changed="onFilterChanged"
        />
      </div>
    </div>
  </v-expand-transition>
  <search-dialog
    v-if="showSearchButton"
    v-model="queryDialogModel"
    :add-entity-enabled="addEntityEnabled"
    @add-entities="onAddEntities"
  />
</template>

<script setup>
import {
	mdiSelect,
	mdiCursorPointer,
	mdiDelete,
	mdiCached,
	mdiImageFilterCenterFocus,
	mdiChartTimelineVariant,
	mdiCog,
	mdiFilterPlus,
	mdiPlus,
} from '@mdi/js';
import {ref} from 'vue';
import {useRoute} from 'vue-router';
import {useHotkey} from 'vuetify';
import ChipFilter from '@/components/explorer/address/ChipFilter.vue';
import SearchDialog from '@/components/common/SearchDialog.vue';
import {BLOCKCHAIN_ATTRIBUTES} from '@/constants/index.js';
import {blenderPlus} from '@/customIcons/index.js';

const route = useRoute();

const emit = defineEmits([
	'isSelectionEnabled',
	'rearrange',
	'center',
	'deleteSelected',
	'addEntities',
	'filterChanged',
	'shortestPathLookup',
	'heuristicBatch',
	'addSelector',
]);

const props = defineProps({
	name: {type: String, required: false, default: ''},
	showSearchButton: {type: Boolean, required: false},
	showDeleteButton: {type: Boolean, required: false},
	showAddSelectorButton: {type: Boolean, required: false},
	selectedItemCount: {type: Number, required: false, default: 0},
	shortestPathEnabled: {type: Boolean, required: false},
	heuristicBatchEnabled: {type: Boolean, required: false},
	addEntityEnabled: {type: Boolean, required: false},
	deleteDisabled: {type: Boolean, required: false},
	oneLine: {type: Boolean, required: false},
	nodeTypeItems: {type: Array, required: false, default: () => []},
	transactionTypeItems: {type: Array, required: false, default: () => []},
	disableFilter: {type: Boolean, required: false},
});

useHotkey('cmd+k', openSearchDialog);

const selectionToggle = ref(1);
const showFilter = ref(false);
const typeFilters = ref(props.transactionTypeItems.map((_, i) => i));
const nodeFilters = ref(props.nodeTypeItems.map((_, i) => i));
const queryDialogModel = ref(false);

// Functions
function onSelectionModeChanged(mode) {
	emit('isSelectionEnabled', mode === 0);
}

async function onAddEntities(entities) {
	emit('addEntities', entities);
}

function onFilterChanged() {
	emit(
		'filterChanged',
		nodeFilters.value.map(d => props.nodeTypeItems[d].text),
		typeFilters.value.map(d => props.transactionTypeItems[d].text),
	);
}

function onAddSelector() {
	emit('addSelector');
}

function openSearchDialog() {
	if (!props.showSearchButton) {
		return;
	}

	queryDialogModel.value = true;
}

</script>

<style scoped>

.workspace-name {
  max-width: 200px;
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}

</style>
