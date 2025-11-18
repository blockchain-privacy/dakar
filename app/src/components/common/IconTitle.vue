<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div>
    <div
      class="d-flex align-center text-h6 ma-2"
      style="white-space: nowrap"
    >
      <v-icon
        start
        :icon="icon"
      />
      <template v-if="slots.title">
        <slot name="title" />
      </template>
      <router-link
        v-else-if="to"
        class="shorten"
        style="color: inherit;"
        :to="to"
      >
        {{ title }}
      </router-link>
      <span
        v-else
        class="shorten"
      > {{ title }}</span>
      <div
        v-if="oneLine || !$vuetify.display.xs"
        class="ms-auto"
      >
        <slot />
      </div>
    </div>
    <div
      v-if="!oneLine && $vuetify.display.xs"
      class="d-flex align-center justify-end"
    >
      <slot />
    </div>
  </div>
</template>

<script setup>

import {useSlots} from 'vue';

defineProps({
	title: {type: String, required: false, default: ''},
	icon: {type: String, required: true},
	to: {type: Object, required: false, default: null},
	oneLine: {type: Boolean, required: false},
});

const slots = useSlots();

</script>

<style scoped>
.shorten {
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}
</style>
