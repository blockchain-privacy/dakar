<template>
  <v-slide-x-reverse-transition>
    <v-sheet
      v-show="model"
      class="sidebar"
      elevation="4"
      :style="`max-width:min(${maxWidth}, 100vw); min-width:${minWidth}`"
    >
      <v-card-title class="d-flex align-center py-0 mb-1">
        <v-icon class="me-2">
          {{ icon }}
        </v-icon>
        <span class="shorten"> {{ title }}</span>
        <div class="ms-auto">
          <slot
            v-if="titleOneLine || !$vuetify.display.xs"
            name="actions"
          />
          <v-btn
            :icon="true"
            variant="text"
            color="grey"
            @click="model = false"
          >
            <v-icon :icon="mdiCloseCircle" />
          </v-btn>
        </div>
      </v-card-title>
      <v-card-title
        v-if="!titleOneLine && $vuetify.display.xs"
        class="d-flex align-center justify-end mb-1 pt-0"
        style="margin-top: -5px"
      >
        <slot name="actions" />
      </v-card-title>
      <v-divider />
      <slot name="body" />
    </v-sheet>
  </v-slide-x-reverse-transition>
</template>

<script setup>
import {mdiCloseCircle} from '@mdi/js';

defineProps({
	title: {type: String, required: true},
	icon: {type: String, required: true},
	maxWidth: {type: String, required: false, default: '600px'},
	minWidth: {type: String, required: false, default: '300px'},
	titleOneLine: {type: Boolean, required: false, default: true},
});

const model = defineModel({type: Boolean});

</script>

<style scoped>
.sidebar {
  position: absolute;
  top: 0;
  right: 0;
  height: 100%;
  /* Heuristic toolbar a z-index of 1004, therefore set z-index to the same so top shadow is not visible */
  z-index: 1004;
  overflow: auto;
}

.shorten {
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
  margin-right: 2px;
}
</style>
