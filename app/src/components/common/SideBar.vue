<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-slide-x-reverse-transition>
    <v-sheet
      v-show="model"
      ref="sidebarBody"
      class="sidebar"
      elevation="4"
      :style="sheetStyle"
    >
      <!-- need z-index so slot is not above sticky header -->
      <div
        class="position-sticky top-0"
        style="z-index: 10; background-color: rgb(var(--v-theme-surface))"
      >
        <v-card-title class="d-flex align-center py-0 mb-1">
          <v-icon
            start
            :icon="icon"
          />
          <span class="shorten"> {{ title }}</span>
          <div class="ms-auto text-title-medium">
            <slot
              v-if="titleOneLine"
              name="actions"
            />
            <v-btn
              v-if="!disableFullScreen"
              icon
              variant="text"
              color="grey"
              @click="isFullScreen = !isFullScreen"
            >
              <v-icon :icon="isFullScreen?mdiFullscreenExit:mdiFullscreen" />
            </v-btn>
            <v-btn
              icon
              variant="text"
              color="grey"
              @click="model = false"
            >
              <v-icon :icon="mdiCloseCircle" />
            </v-btn>
          </div>
        </v-card-title>
        <v-card-title
          v-if="!titleOneLine"
          class="d-flex align-center justify-start mb-1 pt-0"
          style="margin-top: -5px"
        >
          <div class="overflow-auto">
            <slot name="actions" />
          </div>
        </v-card-title>
        <v-card-title
          v-if="slots.secondaryActions"
          class="d-flex align-center justify-start mb-1 pt-0"
          style="margin-top: -5px"
        >
          <div class="overflow-auto">
            <slot name="secondaryActions" />
          </div>
        </v-card-title>
        <v-divider />
      </div>
      <div :class="{'increaseMargin': expandMargins}">
        <slot name="body" />
      </div>
    </v-sheet>
  </v-slide-x-reverse-transition>
</template>

<script setup>
import {mdiCloseCircle, mdiFullscreen, mdiFullscreenExit} from '@mdi/js';
import {
	computed,
	onMounted,
	onUnmounted,
	ref,
	useSlots,
	useTemplateRef,
} from 'vue';

const props = defineProps({
	title: {type: String, required: true},
	icon: {type: String, required: true},
	maxWidth: {type: String, required: false, default: '600px'},
	minWidth: {type: String, required: false, default: '300px'},
	titleOneLine: {type: Boolean, required: false},
	disableFullScreen: {type: Boolean, required: false},
});

const model = defineModel({type: Boolean});
const slots = useSlots();
const isFullScreen = ref(false);
const expandMargins = ref(false);
const sidebarBody = useTemplateRef('sidebarBody');

let resizeObserver;

// Hooks
onMounted(() => {
	window.addEventListener('keydown', keyListener);
	resizeObserver = new ResizeObserver(entries => {
		// Need to use ResizeObserver instead of CSS container query,
		// because that would also apply size containment which restricts
		// the sidebar from growing (connection sidebar and shortest path sidebar).
		const {width} = entries[0].contentRect;
		expandMargins.value = width >= 1545;
	});
	resizeObserver.observe(sidebarBody.value.$el);
});

onUnmounted(() => {
	window.removeEventListener('keydown', keyListener);
	resizeObserver.disconnect();
});

// Computed
const sheetStyle = computed(() => {
	const ret = {
		'min-width': `${props.minWidth}`,
	};

	if (isFullScreen.value) {
		ret.width = '100vw';
	} else {
		ret['max-width'] = `min(${props.maxWidth}, 100vw)`;
	}

	return ret;
});

// Functions
function keyListener(event) {
	if (event.key === 'Escape') {
		model.value = false;
	}
}

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

.increaseMargin {
  margin-left: 200px;
  margin-right: 200px;
}

</style>
