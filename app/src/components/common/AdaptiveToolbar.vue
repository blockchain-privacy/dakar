<template>
  <div class="d-flex align-center">
    <template v-if="name && !$vuetify.display.xs">
      <v-icon
        class="mx-3"
        icon="$graphIcon"
        size="32"
      />
      <p class="me-3 text-h6 workspace-name">
        {{ name }}
      </p>
    </template>
    <v-text-field
      v-if="showSearchField"
      v-model="graphQuery"
      class="noOutline"
      hide-details
      variant="outlined"
      density="compact"
      color="primary"
      single-line
      label="Add entities"
      :disabled="!addEntityEnabled"
      :append-inner-icon="mdiMagnify"
      @click:append-inner="onAddEntity"
      @keydown.enter="onAddEntity"
    />
    <v-btn
      v-if="!disableFilter"
      variant="text"
      :icon="mdiCog"
      :active="showFilter"
      @click="showFilter = !showFilter"
    />
    <v-btn-toggle
      v-if="!oneLine"
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
  <div class="d-flex justify-center flex-wrap">
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
      v-if="showDeleteButton && selectedItemCount > 0"
      variant="flat"
      class="my-1 me-1"
      :disabled="!deleteEnabled"
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
          v-model="privacyFilters"
          label="Transaction Types"
          :items="privacyTypeItems"
          @changed="onFilterChanged"
        />
      </div>
    </div>
  </v-expand-transition>
</template>

<script setup>
import {
	mdiSelect, mdiCursorPointer, mdiDelete, mdiCached, mdiImageFilterCenterFocus,
	mdiMagnify, mdiChartTimelineVariant, mdiCog, mdiFilterPlus,
} from '@mdi/js';
import {ref} from 'vue';
import ChipFilter from '@/components/explorer/address/ChipFilter.vue';

const emit = defineEmits([
	'isSelectionEnabled',
	'rearrange',
	'center',
	'deleteSelected',
	'addEntity',
	'filterChanged',
	'shortestPathLookup',
	'addSelector',
]);

const props = defineProps({
	name: {type: String, required: false, default: ''},
	showSearchField: {type: Boolean, required: false, default: true},
	selectedItemCount: {type: Number, required: false, default: 0},
	deleteEnabled: {type: Boolean, required: false, default: true},
	shortestPathEnabled: {type: Boolean, required: false, default: false},
	addEntityEnabled: {type: Boolean, required: false, default: true},
	oneLine: {type: Boolean, required: false, default: false},
	showDeleteButton: {type: Boolean, required: false, default: true},
	showAddSelectorButton: {type: Boolean, required: false, default: true},
	nodeTypeItems: {type: Array, required: false, default: () => []},
	privacyTypeItems: {type: Array, required: false, default: () => []},
	disableFilter: {type: Boolean, required: false, default: false},
});

const selectionToggle = ref(1);
const graphQuery = ref('');
const showFilter = ref(false);
const privacyFilters = ref(props.privacyTypeItems.map((_, i) => i));
const nodeFilters = ref(props.nodeTypeItems.map((_, i) => i));

// Functions
function onSelectionModeChanged(mode) {
	emit('isSelectionEnabled', mode === 0);
}

function onAddEntity() {
	emit('addEntity', graphQuery.value);
	graphQuery.value = '';
}

function onFilterChanged() {
	emit(
		'filterChanged',
		nodeFilters.value.map(d => props.nodeTypeItems[d].text),
		privacyFilters.value.map(d => props.privacyTypeItems[d].text),
	);
}

function onAddSelector() {
	emit('addSelector');
}

</script>

<style scoped>

.workspace-name {
  max-width: 220px;
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}

/* remove outline from text-field variant 'outlined'.
 This can also be achieved by using variant 'plain',
 but then the label text is not centered */
.noOutline :deep(.v-field__outline__start) {
  border-width: 0 0 0 0 !important;
}

.noOutline :deep(.v-field__outline__end) {
  border-width: 0 0 0 0 !important;
}
</style>
