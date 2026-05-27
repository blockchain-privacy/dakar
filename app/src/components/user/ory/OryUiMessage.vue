<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-alert
    :type="message.type"
    class="mt-2"
    variant="text"
  >
    {{ message.text }}
  </v-alert>
</template>

<script setup>

defineProps({message: {type: Object, required: true}});

</script>

<style scoped>

</style>
