<template>
  <div class="d-flex align-center">
    <v-icon
      class="mx-3"
      icon="$graphIcon"
      size="32"
    />
    <p class="me-3 text-h6 workspace-name hidden-sm-and-down">
      {{ name }}
    </p>
    <v-text-field
      v-model="graphQuery"
      class="noOutline flex-grow-1"
      style="min-width:220px; max-width:400px"
      :hide-details="true"
      variant="outlined"
      density="compact"
      color="primary"
      :single-line="true"
      label="Add entities"
      :disabled="!addEntityEnabled"
      :append-inner-icon="mdiMagnify"
      @click:append-inner="onAddEntity"
      @keydown.enter="onAddEntity"
    />
    <div class="hidden-md-and-down">
      <v-btn
        variant="text"
        class="my-1"
        @click="onRearrange"
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
        @click="onCenter"
      >
        <v-icon
          :icon="mdiImageFilterCenterFocus"
          class="me-1"
        />
        Center
      </v-btn>
      <v-btn
        variant="text"
        class="my-1"
        :to="{name: ROUTE_NAME_WORKSPACES_PAGE}"
      >
        <v-icon
          :icon="mdiOpenInNew"
          class="me-1"
        />
        Workspaces
      </v-btn>
      <v-scroll-x-reverse-transition>
        <v-btn
          v-if="selectedItemCount > 0"
          variant="flat"
          class="my-1"
          :disabled="!deleteEnabled"
          @click="onDeleteSelected"
        >
          <v-icon
            :icon="mdiDelete"
            class="me-1"
          />
          {{ `Delete ${selectedItemCount} ${plural('node', selectedItemCount)}` }}
        </v-btn>
      </v-scroll-x-reverse-transition>
    </div>
    <v-spacer />
    <v-btn-toggle
      v-model="selectionToggle"
      color="primary"
      class="ms-2"
      rounded="0"
      mandatory
      @update:model-value="selectionModeChanged"
    >
      <v-btn :icon="mdiSelect" />
      <v-btn :icon="mdiCursorPointer" />
    </v-btn-toggle>
  </div>
  <div class="hidden-lg-and-up">
    <div class="d-flex justify-center flex-wrap">
      <v-btn
        variant="text"
        class="my-1"
        @click="onRearrange"
      >
        <v-icon
          :icon="mdiCached"
          class="me-1"
        />
        <template v-if="$vuetify.display.smAndUp">
          Rearrange
        </template>
      </v-btn>
      <v-btn
        variant="text"
        class="my-1"
        @click="onCenter"
      >
        <v-icon
          :icon="mdiImageFilterCenterFocus"
          class="me-1"
        />
        <template v-if="$vuetify.display.smAndUp">
          Center
        </template>
      </v-btn>
      <v-btn
        variant="text"
        class="my-1"
        :to="{name: ROUTE_NAME_WORKSPACES_PAGE}"
      >
        <v-icon
          :icon="mdiOpenInNew"
          class="me-1"
        />
        <template v-if="$vuetify.display.smAndUp">
          Workspaces
        </template>
      </v-btn>
      <v-scroll-x-reverse-transition>
        <v-btn
          v-if="selectedItemCount > 0"
          variant="flat"
          class="my-1"
          :disabled="!deleteEnabled"
          @click="onDeleteSelected"
        >
          <v-icon
            :icon="mdiDelete"
            class="me-1"
          />
          {{ selectedItemCount }}
        </v-btn>
      </v-scroll-x-reverse-transition>
    </div>
  </div>
</template>

<script setup>
import {
	mdiSelect, mdiCursorPointer, mdiDelete, mdiCached,
	mdiImageFilterCenterFocus, mdiOpenInNew, mdiMagnify,
} from '@mdi/js';
import {ref} from 'vue';
import {ROUTE_NAME_WORKSPACES_PAGE} from '@/constants';
import {plural} from '@/utilities/index.js';

const emit = defineEmits(['isSelectionEnabled', 'rearrange', 'center', 'deleteSelected', 'addEntity']);
defineProps({
	name: {type: String, required: true},
	selectedItemCount: {type: Number, required: true},
	deleteEnabled: {type: Boolean, required: true},
	addEntityEnabled: {type: Boolean, required: true},
});

const selectionToggle = ref(1);
const graphQuery = ref('');

// Functions

function selectionModeChanged(mode) {
	emit('isSelectionEnabled', mode === 0);
}

function onRearrange() {
	emit('rearrange');
}

function onCenter() {
	emit('center');
}

function onDeleteSelected() {
	emit('deleteSelected');
}

function onAddEntity() {
	emit('addEntity', graphQuery.value);
	graphQuery.value = '';
}

</script>

<style scoped>

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
