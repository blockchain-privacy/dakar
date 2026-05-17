<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-list-item>
    <template #prepend>
      <v-icon
        v-if="isColor"
        :icon="icon"
        :class="{ 'green-icon': !isRed, 'red-icon': isRed }"
        size="xx-large"
      />
      <v-icon
        v-else
        size="xx-large"
        :icon="icon"
      />
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
      <v-list-item-title class="text-title-small">
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
