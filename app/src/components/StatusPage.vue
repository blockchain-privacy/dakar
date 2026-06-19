<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div>
    <status
      v-for="item in authPerMode"
      :key="item.mode"
      :title="item.title"
      :blockchain-mode="item.mode"
    />
  </div>
</template>

<script setup>
import {computed, onMounted} from 'vue';
import {storeToRefs} from 'pinia';
import {BLOCKCHAIN_ATTRIBUTES, PAGE_TITLE} from '@/constants/index.js';
import {isAdminIdentity, isPrivilegedIdentity} from '@/utilities/index.js';
import {useLocalStore} from '@/pinia/local.js';
import Status from '@/components/Status.vue';

const {session} = storeToRefs(useLocalStore());

// Computed
const authPerMode = computed(() => Object.values(BLOCKCHAIN_ATTRIBUTES).filter(m => isPrivilegedIdentity(session.value, m.mode)
	|| isAdminIdentity(session.value, m.mode)));

// Hooks
onMounted(() => {
	document.title = `Status - ${PAGE_TITLE}`;
});

</script>

<style scoped>

</style>
