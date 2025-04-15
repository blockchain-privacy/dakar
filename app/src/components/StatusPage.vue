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
import {BLOCKCHAIN_ATTRIBUTES} from '@/constants/index.js';
import {computed} from 'vue';
import {isAdminIdentity, isPrivilegedIdentity} from '@/utilities/index.js';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';
import Status from '@/components/Status.vue';

const {session} = storeToRefs(useLocalStore());

// Computed
const authPerMode = computed(() => Object.values(BLOCKCHAIN_ATTRIBUTES).filter(m => isPrivilegedIdentity(session.value, m.mode)
	|| isAdminIdentity(session.value, m.mode)));

</script>

<style scoped>

</style>
