<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div
    class="d-flex align-center justify-center"
    style="height: 100%; width:100%"
  >
    <v-card
      max-width="600px"
      style="flex:1"
    >
      <div class="pa-5">
        <div class="d-flex justify-center">
          <v-icon
            :icon="mdiCheckCircle"
            color="success"
            size="100"
          />
        </div>
        <h3 class="text-h5 font-weight-bold text-center mb-2">
          Authentication successful. You can close this page now.
        </h3>
      </div>
    </v-card>
  </div>
</template>

<script setup>
import {PAGE_TITLE} from '@/constants';
import {onMounted} from 'vue';
import {mdiCheckCircle} from '@mdi/js';

// Hooks
onMounted(async () => {
	document.title = `OAuth Success - ${PAGE_TITLE}`;
});

// Functions

</script>

<style scoped>

</style>
