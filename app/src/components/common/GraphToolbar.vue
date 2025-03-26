<template>
  <div class="d-flex align-center justify-space-between">
    <div
      v-if="name"
      class="d-flex align-center d-inline-block"
    >
      <v-icon
        class="mx-3"
        icon="$graphIcon"
        size="32"
      />
      <p
        v-tooltip="{'text': 'Name of the Workspace', 'location':'top', 'open-delay': 400}"
        class="text-h6 my-2 workspace-name"
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
        :icon="mdiCog"
        class="me-1"
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
        class="me-1"
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
        class="me-1"
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
        class="me-1"
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
          <div class="kbb">
            Control
          </div>
          +
          <div class="kbb">
            k
          </div>
        </div>
      </v-tooltip>
      <v-icon
        :icon="mdiPlus"
        class="me-1"
      />
      Add entities
    </v-btn>
    <v-btn
      v-if="showDeleteButton && selectedItemCount > 0"
      :disabled="deleteDisabled"
      variant="flat"
      class="my-1 me-1"
      @click="emit('deleteSelected')"
    >
      <v-icon
        :icon="mdiDelete"
        class="me-1"
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
        class="me-1"
      />
      Shortest path
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
	mdiSelect, mdiCursorPointer, mdiDelete, mdiCached, mdiImageFilterCenterFocus,
	mdiChartTimelineVariant, mdiCog, mdiFilterPlus, mdiPlus,
} from '@mdi/js';
import {onMounted, onUnmounted, ref} from 'vue';
import ChipFilter from '@/components/explorer/address/ChipFilter.vue';
import SearchDialog from '@/components/common/SearchDialog.vue';

const emit = defineEmits([
	'isSelectionEnabled',
	'rearrange',
	'center',
	'deleteSelected',
	'addEntities',
	'filterChanged',
	'shortestPathLookup',
	'addSelector',
]);

const props = defineProps({
	name: {type: String, required: false, default: ''},
	showSearchButton: {type: Boolean, required: false},
	showDeleteButton: {type: Boolean, required: false},
	showAddSelectorButton: {type: Boolean, required: false},
	selectedItemCount: {type: Number, required: false, default: 0},
	shortestPathEnabled: {type: Boolean, required: false},
	addEntityEnabled: {type: Boolean, required: false},
	deleteDisabled: {type: Boolean, required: false},
	oneLine: {type: Boolean, required: false},
	nodeTypeItems: {type: Array, required: false, default: () => []},
	transactionTypeItems: {type: Array, required: false, default: () => []},
	disableFilter: {type: Boolean, required: false},
});

const selectionToggle = ref(1);
const showFilter = ref(false);
const typeFilters = ref(props.transactionTypeItems.map((_, i) => i));
const nodeFilters = ref(props.nodeTypeItems.map((_, i) => i));
const queryDialogModel = ref(false);

// Hooks
onMounted(() => {
	document.addEventListener('keydown', handleKeyPress, false);
});

onUnmounted(() => {
	document.removeEventListener('keydown', handleKeyPress);
});

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

function handleKeyPress(e) {
	// Don't trigger delete event when editing an <input /> element such as text boxes
	if (e.target instanceof HTMLInputElement || !e.ctrlKey || e.key !== 'k' || !props.showSearchButton) {
		return;
	}

	queryDialogModel.value = true;
	e.preventDefault();
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
