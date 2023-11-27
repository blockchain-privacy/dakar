<template>
  <v-dialog
    v-model="show"
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
import {computed} from 'vue';

const props = defineProps({
	modelValue: {type: Boolean, required: true},
	txHash: {type: String, required: true},
	privacyType: {type: String, required: true},
	dateTime: {type: Date, required: true},
	inputTxs: {type: Array, required: true},
});

const emit = defineEmits(['update:modelValue']);

const show = computed({
	get() {
		return props.modelValue;
	},
	set(value) {
		emit('update:modelValue', value);
	},
});

</script>

<style scoped>

</style>
