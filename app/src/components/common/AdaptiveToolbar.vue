<template>
  <div class="d-flex align-center">
    <template v-if="name">
      <v-icon
        class="mx-3"
        icon="$graphIcon"
        size="32"
      />
      <p class="me-3 text-h6 workspace-name hidden-sm-and-down">
        {{ name }}
      </p>
    </template>
    <v-text-field
      v-if="showSearchField"
      v-model="graphQuery"
      class="noOutline flex-grow-1"
      :style="adaptive && $vuetify.display.xs?'min-width:100px':'min-width:200px'"
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
    <div :class="{'hidden-md-and-down':adaptive}">
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
        v-if="showWorkspacesButton"
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
          v-if="showDeleteButton && selectedItemCount > 0"
          variant="flat"
          class="my-1"
          :disabled="!deleteEnabled"
          @click="emit('deleteSelected')"
        >
          <v-icon
            :icon="mdiDelete"
            class="me-1"
          />
          {{ `Delete ${selectedItemCount} ${plural('node', selectedItemCount)}` }}
        </v-btn>
      </v-scroll-x-reverse-transition>
    </div>
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
  <div
    v-if="adaptive"
    class="hidden-lg-and-up"
  >
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
        v-if="showWorkspacesButton"
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
import {ROUTE_NAME_WORKSPACES_PAGE} from '@/constants/index.js';
import {plural} from '@/utilities/index.js';

const emit = defineEmits(['isSelectionEnabled', 'rearrange', 'center', 'deleteSelected', 'addEntity']);
defineProps({
	name: {type: String, required: false, default: ''},
	showSearchField: {type: Boolean, required: false, default: true},
	selectedItemCount: {type: Number, required: false, default: 0},
	deleteEnabled: {type: Boolean, required: false, default: true},
	addEntityEnabled: {type: Boolean, required: false, default: true},
	adaptive: {type: Boolean, required: false, default: true},
	showDeleteButton: {type: Boolean, required: false, default: true},
	showWorkspacesButton: {type: Boolean, required: false, default: true},
});

const selectionToggle = ref(1);
const graphQuery = ref('');

// Functions
function selectionModeChanged(mode) {
	emit('isSelectionEnabled', mode === 0);
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
