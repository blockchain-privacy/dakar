<template>
  <v-dialog
    v-model="model"
    max-width="700px"
  >
    <v-card>
      <v-card-title class="text-h5">
        Transaction
        <store-link
          class="shorten"
          disable-select
          :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: txHash }}"
          @clicked="model = false"
        >
          {{ txHash }}
        </store-link>
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
            >
              <store-link
                class="shorten"
                disable-select
                :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: t.txhash }}"
                @clicked="model = false"
              >
                {{ t.txhash }}
              </store-link>
            </v-list-item>
          </v-list>
        </v-expand-transition>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import StoreLink from '@/components/common/StoreLink.vue';

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
