<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div
    class="fill-height"
    style="padding: 12px 10px 0 10px"
  >
    <v-row class="fill-height">
      <v-col class="fill-height">
        <router-view v-slot="{ Component }">
          <fade-transition>
            <component :is="Component" />
          </fade-transition>
        </router-view>
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import FadeTransition from '../common/FadeTransition.vue';

</script>

<style scoped>

</style>
