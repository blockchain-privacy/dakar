<template>
  <v-chip
    v-tooltip="{'text':tooltipText,'location':'bottom'}"
    v-bind="$attrs"
    rounded
    color="primary"
    :size="size"
    style="cursor: pointer"
    @click="showFingerprintDialog = true"
  >
    <v-icon
      :icon="mdiFingerprint"
      start
    />
    Fingerprint
  </v-chip>
  <fingerprint-transactions-dialog
    v-model="showFingerprintDialog"
    :transaction-hash="transactionHash"
  />
</template>

<script setup>
import {mdiFingerprint} from '@mdi/js';
import {ref} from 'vue';
import FingerprintTransactionsDialog from '@/components/explorer/transaction/FingerprintTransactionsDialog.vue';

defineProps({
	transactionHash: {type: String, required: true},
	size: {type: String, required: false, default: 'default'},
});

const showFingerprintDialog = ref(false);
const tooltipText = 'Search for similar destination transactions';

</script>

<style scoped>

</style>
