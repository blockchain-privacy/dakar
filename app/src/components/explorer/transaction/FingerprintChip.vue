<template>
  <v-chip
    :id="`btn_find_similar_transactions_${transactionHash}`"
    :rounded="true"
    color="primary"
    :size="size"
    style="cursor: pointer"
    v-bind="$attrs"
    @click="showFingerprintDialog = true"
  >
    <v-icon
      :icon="mdiFingerprint"
      class="me-2"
    />
    Fingerprint
  </v-chip>
  <fingerprint-transactions-dialog
    v-model="showFingerprintDialog"
    :transaction-hash="transactionHash"
  />
  <v-tooltip
    location="bottom"
    :activator="`#btn_find_similar_transactions_${transactionHash}`"
  >
    <span>Search for similar destination transactions</span>
  </v-tooltip>
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

</script>

<style scoped>

</style>
