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
      v-if="showSearchField"
      variant="text"
      :disabled="!addEntityEnabled"
      @click="queryDialogModel = true"
    >
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
  <v-dialog
    v-model="queryDialogModel"
    max-width="600"
  >
    <v-card title="Add entities">
      <v-form
        id="queryForm"
        ref="queryForm"
        validate-on="submit"
      >
        <v-card-text>
          <p class="text-subtitle-1">
            Add one or multiple entities. Separate multiple entities by any special character.
          </p>
          <v-text-field
            v-model="graphQuery"
            class="mt-4"
            autofocus
            variant="outlined"
            density="compact"
            color="primary"
            :rules="inputRules"
            label="Add a transactions or address clusters"
            :disabled="!addEntityEnabled"
            :append-inner-icon="mdiMagnify"
            @click:append-inner="onAddEntity"
            @keydown.enter="onAddEntity"
          />
          <v-expand-transition>
            <div
              v-if="queryItemCount > 1"
              class="d-flex justify-center"
            >
              <v-btn
                variant="text"
                @click="showDetectedEntities = !showDetectedEntities"
              >
                {{ showDetectedEntities?'Hide':'Show' }} detected entities
              </v-btn>
            </div>
          </v-expand-transition>
          <v-expand-transition>
            <div v-if="queryItemCount > 1 && showDetectedEntities">
              <v-list
                v-for="entity in detectedEntities"
                :key="entity"
                density="compact"
              >
                <v-list-item>
                  {{ entity }}
                </v-list-item>
              </v-list>
            </div>
          </v-expand-transition>
        </v-card-text>
        <v-card-actions>
          <v-btn
            class="ml-auto"
            text="Cancel"
            @click="queryDialogModel = false"
          />
          <v-btn
            :disabled="queryItemCount === 0"
            @click="onAddEntity"
          >
            Add {{ queryItemCount > 0?queryItemCount:'' }} {{ pluralIrregular('entity','entities', queryItemCount) }}
          </v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {
	mdiSelect, mdiCursorPointer, mdiDelete, mdiCached, mdiImageFilterCenterFocus,
	mdiMagnify, mdiChartTimelineVariant, mdiCog, mdiFilterPlus, mdiPlus,
} from '@mdi/js';
import {ref, computed} from 'vue';
import ChipFilter from '@/components/explorer/address/ChipFilter.vue';
import {extractEntities, pluralIrregular} from '@/utilities/index.js';

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
	showSearchField: {type: Boolean, required: false},
	selectedItemCount: {type: Number, required: false, default: 0},
	shortestPathEnabled: {type: Boolean, required: false},
	addEntityEnabled: {type: Boolean, required: false},
	deleteDisabled: {type: Boolean, required: false},
	oneLine: {type: Boolean, required: false},
	showDeleteButton: {type: Boolean, required: false},
	showAddSelectorButton: {type: Boolean, required: false},
	nodeTypeItems: {type: Array, required: false, default: () => []},
	transactionTypeItems: {type: Array, required: false, default: () => []},
	disableFilter: {type: Boolean, required: false},
});

const selectionToggle = ref(1);
const graphQuery = ref('');
const showFilter = ref(false);
const typeFilters = ref(props.transactionTypeItems.map((_, i) => i));
const nodeFilters = ref(props.nodeTypeItems.map((_, i) => i));
const queryDialogModel = ref(false);
const queryForm = ref(null);
const showDetectedEntities = ref(false);

const inputRules = [
	q => extractEntities(q).length > 0 || 'query contains no valid entities',
];

// Computed
const detectedEntities = computed(() => extractEntities(graphQuery.value));

const queryItemCount = computed(() => detectedEntities.value.length);

// Functions
function onSelectionModeChanged(mode) {
	emit('isSelectionEnabled', mode === 0);
}

async function onAddEntity() {
	const {valid} = await queryForm.value.validate();
	if (!valid) {
		return;
	}

	queryDialogModel.value = false;
	emit('addEntities', extractEntities(graphQuery.value));
	graphQuery.value = '';
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

</script>

<style scoped>

.workspace-name {
  max-width: 200px;
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
