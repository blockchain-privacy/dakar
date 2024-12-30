<template>
  <v-dialog
    v-model="model"
    max-width="700px"
  >
    <v-card>
      <v-card-title class="text-h5 text-wrap">
        Similar Destination Transactions
      </v-card-title>
      <v-card-text>
        <fingerprint-transactions :transaction-hash="transactionHash" />
      </v-card-text>
      <v-card-actions>
        <v-btn @click="model = false">
          Close
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import FingerprintTransactions from '@/components/explorer/transaction/FingerprintTransactions.vue';
const model = defineModel({type: Boolean});
defineProps({transactionHash: {type: String, required: true}});
</script>

<style scoped>

</style>
