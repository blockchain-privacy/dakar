<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="fill-height">
    <router-view v-slot="{ Component }">
      <fade-transition>
        <component :is="Component" />
      </fade-transition>
    </router-view>
  </div>
</template>

<script setup>
import FadeTransition from '../common/FadeTransition.vue';

</script>

<style scoped>

</style>
