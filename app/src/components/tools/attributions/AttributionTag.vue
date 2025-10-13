<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-chip
    v-tooltip="{'text': tooltipText, 'location':'bottom'}"
    label
    class="overflow-x-auto"
    :color="color"
    rounded
  >
    {{ attribution.tag }}
  </v-chip>
</template>

<script setup>
import {computed} from 'vue';
const props = defineProps({attribution: {type: Object, required: true}});
const tooltipText = computed(() => props.attribution.isPublic
	? 'This is a public attribution'
	: 'This is a private attribution. It is only visible to you.');
const color = computed(() => props.attribution.isPublic ? 'primary' : null);
</script>

<style scoped>

</style>
