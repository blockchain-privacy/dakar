<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
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
      <v-card-text>
        <div class="d-flex justify-center">
          <v-icon
            :icon="mdiCheckCircle"
            color="success"
            size="100"
          />
        </div>
        <h5 class="text-body-large text-center mb-2">
          Authentication successful. You can close this page now.
        </h5>
      </v-card-text>
    </v-card>
  </div>
</template>

<script setup>
import {onMounted} from 'vue';
import {mdiCheckCircle} from '@mdi/js';
import {PAGE_TITLE} from '@/constants';

// Hooks
onMounted(async () => {
	document.title = `OAuth Success - ${PAGE_TITLE}`;
});

// Functions

</script>

<style scoped>

</style>
