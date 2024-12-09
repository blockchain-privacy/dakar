<template>
  <v-list-item>
    <template #prepend>
      <v-icon
        v-if="!isColor"
        size="xx-large"
      >
        {{ icon }}
      </v-icon>
      <v-icon
        v-if="isColor"
        :class="{ 'green-icon': !isRed, 'red-icon': isRed }"
        size="xx-large"
      >
        {{ icon }}
      </v-icon>
    </template>
    <div>
      <v-list-item-title>
        {{ title }}
        <v-icon
          v-if="tooltip"
          v-tooltip="{'text':tooltip, 'location':'right'}"
          size="x-small"
          :icon="mdiHelpCircleOutline"
        />
      </v-list-item-title>
      <v-list-item-title class="text-subtitle-2">
        <slot />
      </v-list-item-title>
    </div>
  </v-list-item>
</template>

<script setup>
import {mdiHelpCircleOutline} from '@mdi/js';

defineProps({
	title: {type: String, required: true},
	icon: {type: String, required: true},
	tooltip: {type: String, default: ''},
	isColor: {type: Boolean},
	isRed: {type: Boolean},
});

</script>

<style scoped>
:deep(.green-icon .v-icon__svg) {
  fill: green;
}

:deep(.red-icon .v-icon__svg) {
  fill: red;
}

</style>
