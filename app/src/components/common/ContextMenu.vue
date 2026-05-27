<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-menu
    v-model="model"
    :open-on-hover="false"
    transition="fade-transition"
    :target="[positionX,positionY]"
  >
    <v-list>
      <template
        v-for="(item, index) in menuItems"
        :key="index"
      >
        <v-divider v-if="item.isDivider" />
        <v-list-item
          v-else
          :key="index"
          :disabled="item.disabled && !item.disabled()"
          @click="emitClickEvent(item)"
        >
          <template
            v-if="item.icon"
            #prepend
          >
            <v-icon>{{ item.icon }}</v-icon>
          </template>
          <v-list-item-title>{{ item.title }}</v-list-item-title>
        </v-list-item>
      </template>
    </v-list>
  </v-menu>
</template>

<script setup>

defineProps({
	menuItems: {type: Array, default: () => []},
	positionX: {type: Number, default: 0},
	positionY: {type: Number, default: 0},
});

const model = defineModel({type: Boolean});

// Functions
function emitClickEvent(item) {
	if (item.action) {
		item.action();
	}

	model.value = false;
}

</script>

<style scoped>

</style>
