<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-sheet
    style="width:15px; height:15px"
    rounded
    :color="color?color:'black'"
  />
</template>

<script setup>

defineProps({color: {type: String, required: true}});

</script>
<style scoped>

</style>
