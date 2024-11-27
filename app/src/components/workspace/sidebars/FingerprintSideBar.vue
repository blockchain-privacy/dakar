<template>
  <side-bar
    v-model="model"
    :title="`Fingerprint ${transactionHash}`"
    :icon="mdiFingerprint"
    max-width="648px"
    title-one-line
  >
    <template #actions />
    <template #body>
      <fingerprint-transactions
        v-if="transactionHash"
        :transaction-hash="transactionHash"
      />
    </template>
  </side-bar>
</template>

<script setup>
import SideBar from '@/components/common/SideBar.vue';
import {mdiFingerprint} from '@mdi/js';
import FingerprintTransactions from '@/components/explorer/transaction/FingerprintTransactions.vue';

const model = defineModel({type: Boolean});

defineProps({transactionHash: {type: String, required: true}});

</script>

<style scoped>
.textBorder {
  border-style: solid;
  border-radius: 5px;
  border-width: 1px;
}
</style>
