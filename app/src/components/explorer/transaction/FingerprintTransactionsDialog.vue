<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="700px"
  >
    <v-card>
      <v-card-title class="text-wrap">
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
