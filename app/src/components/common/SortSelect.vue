<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="d-flex align-center">
    <v-select
      v-model="sort"
      :items="items"
      label="Sort by"
      hide-details
      return-object
      :variant="variant"
    />
    <v-btn
      :icon="direction?mdiSortDescending:mdiSortAscending"
      variant="text"
      class="ms-1"
      @click="direction = !direction"
    />
  </div>
</template>

<script setup>
import {mdiSortAscending, mdiSortDescending} from '@mdi/js';

const sort = defineModel('sort', {type: Object});
const direction = defineModel('direction', {type: Boolean});
defineProps({
	items: {type: Array, required: true},
	variant: {type: String, required: false, default: 'outlined'},
});

</script>

<style scoped>

</style>
