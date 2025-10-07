<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-hover v-slot="{ hover }">
    <v-card
      :class="{'on-hover': hover, 'mx-auto px-2': true}"
      rounded
      flat
      :to="to"
      max-width="90"
    >
      <div class="d-flex justify-center">
        <v-icon
          size="37"
          :color="color"
          :icon="icon"
        />
      </div>
      <div class="d-flex justify-center pt-1">
        <div class="text-subtitle-2 title-text">
          {{ title }}
        </div>
      </div>
    </v-card>
  </v-hover>
</template>

<script setup>

defineProps({
	title: {type: String, required: true},
	to: {type: Object, required: true},
	icon: {type: String, required: true},
	color: {type: String, default: 'primary'},
});

</script>

<style scoped>

.title-text {
  text-align: center;
  line-height: 1.3;
}

.on-hover::before {
  opacity: 0.08;
  border-radius: 4px;
}

</style>
