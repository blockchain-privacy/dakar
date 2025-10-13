<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-container
    fluid
    class="content"
  >
    <v-row align="center">
      <v-col>
        <!-- set width and height so transition looks better -->
        <div
          class="d-flex justify-center mb-2 mx-auto"
          style="width: 105px; height: 187px"
        >
          <v-img
            :src="DakarAnimatedImg"
            max-width="105px"
            transition="fade-transition"
          />
        </div>
        <p class="text-h2 text-center">
          {{ APPLICATION_NAME }}
        </p>
        <div class="d-flex justify-center">
          <query-input
            class="my-6"
            style="max-width: 600px"
          />
        </div>
        <p class="text-h6 text-center">
          Blockchain Analytics
        </p>
      </v-col>
    </v-row>
    <!-- footer -->
    <v-container
      v-if="false"
      class="align-self-end"
    >
      <v-row justify="center">
        <v-col
          md="2"
          class="text-center mx-1"
        >
          <v-btn
            :to="{name: ROUTE_NAME_ABOUT}"
            variant="plain"
            size="small"
          >
            About
          </v-btn>
        </v-col>
        <v-col
          md="2"
          class="text-center mx-1"
        >
          <v-btn
            :to="{name: ROUTE_NAME_TERMS_OF_USE}"
            variant="plain"
            size="small"
          >
            Terms of Use
          </v-btn>
        </v-col>
        <v-col
          md="2"
          class="text-center mx-1"
        >
          <v-btn
            :to="{name: ROUTE_NAME_PRIVACY}"
            variant="plain"
            size="small"
          >
            Privacy Policy
          </v-btn>
        </v-col>
      </v-row>
    </v-container>
  </v-container>
</template>

<script setup>
import {
	APPLICATION_NAME, ROUTE_NAME_ABOUT,
	ROUTE_NAME_TERMS_OF_USE, ROUTE_NAME_PRIVACY,
} from '@/constants';
import {onMounted} from 'vue';
import DakarAnimatedImg from '@/assets/dakar_animated.svg?url';
import QueryInput from '@/components/common/QueryInput.vue';

// Hooks
onMounted(() => {
	document.title = APPLICATION_NAME;
});

</script>

<style scoped>

.content {
  display: flex;
  flex-direction: column;
  height: 100%
}

</style>
