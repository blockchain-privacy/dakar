<template>
  <v-dialog
    v-model="model"
    max-width="700px"
  >
    <v-card class="pb-2">
      <v-card-title>
        <div class="text-h5 text-wrap">
          Transaction
          <router-link
            class="ml-1"
            :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: txHash }}"
          >
            {{ txHash }}
          </router-link>
        </div>
      </v-card-title>
      <v-card-text>
        <p class="text-subtitle-1">
          Privacy Type: {{ privacyType }}
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
              :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: t.txhash }}"
            >
              <v-list-item-title>{{ t.txhash }}</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-expand-transition>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';

defineProps({
	txHash: {type: String, required: true},
	privacyType: {type: String, required: true},
	dateTime: {type: Date, required: true},
	inputTxs: {type: Array, required: true},
});

const model = defineModel({type: Boolean});

</script>

<style scoped>

</style>
