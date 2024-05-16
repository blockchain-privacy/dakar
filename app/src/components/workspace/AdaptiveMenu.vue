<template>
  <template v-if="$vuetify.display.lgAndUp">
    <template
      v-for="(item, i) in items"
      :key="i"
    >
      <v-scroll-x-reverse-transition>
        <v-btn
          v-if="!item.show || item.show()"
          :disabled="item.disabled && item.disabled()"
          :variant="item.fill?'flat':'text'"
          :to="item.to?item.to:undefined"
          class="my-1"
          @click="item.action?item.action():undefined"
        >
          <v-icon
            v-if="item.icon"
            :icon="item.icon"
            class="me-1"
          />
          {{ isFunction(item.title)?item.title():item.title }}
        </v-btn>
      </v-scroll-x-reverse-transition>
    </template>
    <v-btn-toggle
      v-model="selectionToggle"
      color="primary"
      rounded="0"
      mandatory
      @update:model-value="selectionModeChanged"
    >
      <v-btn :icon="mdiSelect" />
      <v-btn :icon="mdiCursorPointer" />
    </v-btn-toggle>
  </template>
  <v-menu
    v-if="$vuetify.display.mdAndDown"
    location="bottom"
  >
    <template #activator="{ props }">
      <v-btn
        :icon="true"
        variant="text"
        v-bind="props"
        style="outline: 0"
      >
        <v-icon>{{ mdiDotsVertical }}</v-icon>
      </v-btn>
    </template>
    <v-list>
      <template
        v-for="(item, i) in items"
        :key="i"
      >
        <v-list-item
          v-if="!item.show || item.show()"
          :disabled="item.disabled && item.disabled()"
          :to="item.to?item.to:undefined"
          @click="item.action?item.action():undefined"
        >
          <template
            v-if="item.icon"
            #prepend
          >
            <v-icon :icon="item.icon" />
          </template>
          <v-list-item-title>
            {{ isFunction(item.title)?item.title():item.title }}
          </v-list-item-title>
        </v-list-item>
      </template>
      <v-container>
        <v-btn-toggle
          v-model="selectionToggle"
          color="primary"
          mandatory
          @update:model-value="selectionModeChanged"
        >
          <v-btn>
            <v-icon :icon="mdiSelect" />
            Select
          </v-btn>
          <v-btn>
            Drag
            <v-icon :icon="mdiCursorPointer" />
          </v-btn>
        </v-btn-toggle>
      </v-container>
    </v-list>
  </v-menu>
</template>

<script setup>
import {mdiDotsVertical, mdiSelect, mdiCursorPointer} from '@mdi/js';
import {ref} from 'vue';
import {isFunction} from '../../utilities/index.js';

const emit = defineEmits(['isSelectionEnabled']);
defineProps({
	items: {type: Array, required: true},
});

const selectionToggle = ref(1);

function selectionModeChanged(mode) {
	emit('isSelectionEnabled', mode === 0);
}

</script>

<style scoped>

</style>
