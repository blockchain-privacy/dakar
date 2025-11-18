<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="d-flex align-center">
    <v-select
      v-model="sort"
      :items="items"
      label="Sort by"
      hide-details
      return-object
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
});

</script>

<style scoped>

</style>
