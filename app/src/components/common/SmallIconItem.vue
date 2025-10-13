<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div
    v-tooltip="{'text':tooltip, 'location':'top'}"
    class="d-flex align-center"
  >
    <v-icon
      :icon="icon"
      start
      size="large"
      style="opacity: var(--v-medium-emphasis-opacity)"
    />
    <p
      v-if="title"
      class="text-subtitle-1"
    >
      {{ title }}
    </p>
    <slot />
  </div>
</template>

<script setup>
defineProps({
	icon: {type: String, required: true},
	title: {type: String, required: false, default: ''},
	tooltip: {type: String, default: ''},
});
</script>

<style scoped>

</style>
