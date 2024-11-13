<template>
  <v-dialog
    v-model="model"
    max-width="700px"
  >
    <v-card>
      <v-card-title class="text-h5">
        Transaction
        <workspace-link
          class="shorten"
          disable-select
          :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: txHash, blockchainMode: getSettings.blockchainMode }}"
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
                class="shorten"
                disable-select
                :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: t.txhash, blockchainMode: getSettings.blockchainMode }}"
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
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';
const {getSettings} = storeToRefs(useLocalStore());

defineProps({
	txHash: {type: String, required: true},
	transactionType: {type: String, required: true},
	dateTime: {type: Date, required: true},
	inputTxs: {type: Array, required: true},
});

const model = defineModel({type: Boolean});
</script>

<style scoped>

</style>
