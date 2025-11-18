<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div>
    <attribution-tabs
      v-for="item in authPerMode"
      :key="item.mode"
      :title="item.title"
      :blockchain-mode="item.mode"
    />
  </div>
</template>

<script setup>
import {BLOCKCHAIN_ATTRIBUTES} from '@/constants/index.js';
import {computed} from 'vue';
import {isAdminIdentity, isPrivilegedIdentity} from '@/utilities/index.js';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';
import AttributionTabs from '@/components/tools/attributions/AttributionTabs.vue';

const {session} = storeToRefs(useLocalStore());

// Computed
const authPerMode = computed(() => Object.values(BLOCKCHAIN_ATTRIBUTES).filter(m => isPrivilegedIdentity(session.value, m.mode)
	|| isAdminIdentity(session.value, m.mode)));
</script>

<style scoped>

</style>
