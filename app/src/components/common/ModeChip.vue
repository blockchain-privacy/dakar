<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-chip
    :prepend-icon="mode.icon"
    :text="mode.title"
    :color="mode.color"
    rounded
  />
</template>

<script setup>
import {computed} from 'vue';
import {BLOCKCHAIN_ATTRIBUTES} from '@/constants/index.js';

const props = defineProps({blockchainMode: {type: String, required: true}});

// Computed
const mode = computed(() => BLOCKCHAIN_ATTRIBUTES[props.blockchainMode]);

</script>

<style scoped>

</style>
