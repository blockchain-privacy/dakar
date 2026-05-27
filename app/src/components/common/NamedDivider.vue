<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div :class="`my-${verticalMargin}`">
    <div class="d-flex align-center">
      <v-sheet
        color="grey-lighten-2"
        height="1"
        width="100%"
      />
      <div
        v-if="title"
        :class="`mx-2 ${titleClass}`"
        style="white-space: nowrap;"
      >
        {{ title }}
      </div>
      <v-sheet
        v-if="title"
        color="grey-lighten-2"
        height="1"
        width="100%"
      />
    </div>
  </div>
</template>

<script setup>

defineProps({
	title: {type: String, required: false, default: ''},
	verticalMargin: {type: Number, required: false, default: 8},
	titleClass: {type: String, required: false, default: ''},
});

</script>

<style scoped>

</style>
