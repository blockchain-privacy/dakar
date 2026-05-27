<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div>
    <div
      class="d-flex align-center ma-2"
      style="white-space: nowrap"
    >
      <v-icon
        v-if="icon"
        start
        size="x-large"
        :icon="icon"
      />
      <div
        v-if="slots.title"
        class="text-title-large shorten"
      >
        <slot name="title" />
      </div>
      <router-link
        v-else-if="to"
        class="shorten text-title-large"
        style="color: inherit;"
        :to="to"
      >
        {{ title }}
      </router-link>
      <span
        v-else
        class="shorten text-title-large"
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
	icon: {type: String, required: false, default: ''},
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
