<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="700px"
  >
    <v-card>
      <v-card-title class="text-h5">
        Transaction
        <workspace-link
          disable-select
          :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: txHash, blockchainMode: route.params.blockchainMode }}"
          @clicked="model = false"
        >
          {{ txHash }}
        </workspace-link>
      </v-card-title>
      <v-card-text>
        <p class="text-subtitle-1">
          Transaction Type: {{ transactionType }}
        </p>
        <p class="text-subtitle-1">
          Timestamp: {{ dateTime.toLocaleString() }}
        </p>
        <p
          v-if="inputTxs && inputTxs.length > 0"
          class="text-subtitle-1"
        >
          Input Transactions:
        </p>
        <v-expand-transition>
          <v-list v-if="inputTxs">
            <v-list-item
              v-for="(t) in inputTxs"
              :key="t.txhash"
            >
              <workspace-link
                disable-select
                :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: t.txhash, blockchainMode: route.params.blockchainMode }}"
                @clicked="model = false"
              >
                {{ t.txhash }}
              </workspace-link>
            </v-list-item>
          </v-list>
        </v-expand-transition>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';
import {useRoute} from 'vue-router';

defineProps({
	txHash: {type: String, required: true},
	transactionType: {type: String, required: true},
	dateTime: {type: Date, required: true},
	inputTxs: {type: Array, required: true},
});

const model = defineModel({type: Boolean});
const route = useRoute();
</script>

<style scoped>

</style>
