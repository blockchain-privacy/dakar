<template>
  <v-dialog
    v-model="show"
    max-width="1200px"
  >
    <v-card class="pb-2">
      <v-card-title>
        <div class="text-h5 text-wrap">
          Privacy Transactions from {{ startDate }} to {{ endDate }}
        </div>
      </v-card-title>
      <v-card-text>
        <v-data-table
          v-if="transactions.length > 0"
          :headers="headers"
          :items="transactions"
        >
          <template #item.txhash="{ item }">
            <router-link :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: item.txhash }}">
              {{ item.txhash }}
            </router-link>
          </template>
          <template #item.dateTime="{ item }">
            <span>{{ item.dateTime.toLocaleString() }}</span>
          </template>
        </v-data-table>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import {computed} from 'vue';

const props = defineProps({
	modelValue: {type: Boolean, required: true},
	transactions: {type: Array, required: true},
	headers: {type: Array, required: true},
	startDate: {type: String, required: true},
	endDate: {type: String, required: true},
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
