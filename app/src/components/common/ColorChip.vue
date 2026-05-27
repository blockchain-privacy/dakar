<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-chip rounded>
    <template #prepend>
      <color-sheet
        :color="color"
        class="me-2"
      />
    </template>
    {{ capitalize(title) }}
  </v-chip>
</template>

<script setup>
import {capitalize} from '@/utilities/index.js';
import ColorSheet from '@/components/common/ColorSheet.vue';

defineProps({
	color: {type: String, required: true},
	title: {type: String, required: true},
});
</script>

<style scoped>

</style>
