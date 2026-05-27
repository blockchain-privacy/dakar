<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div>
    <cluster-list
      v-for="item in authPerMode"
      :key="item.mode"
      :title="item.title"
      :blockchain-mode="item.mode"
    />
  </div>
</template>
<script setup>
import {computed} from 'vue';
import {storeToRefs} from 'pinia';
import {BLOCKCHAIN_ATTRIBUTES} from '@/constants/index.js';
import {isAdminIdentity, isPrivilegedIdentity} from '@/utilities/index.js';
import {useLocalStore} from '@/pinia/local.js';
import ClusterList from '@/components/tools/clusters/ClusterList.vue';

const {session} = storeToRefs(useLocalStore());

// Computed
const authPerMode = computed(() => Object.values(BLOCKCHAIN_ATTRIBUTES).filter(m => isPrivilegedIdentity(session.value, m.mode)
	|| isAdminIdentity(session.value, m.mode)));
</script>

<style scoped>

</style>
