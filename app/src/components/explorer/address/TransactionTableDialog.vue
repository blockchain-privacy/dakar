<template>
  <v-dialog
    v-model="model"
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
            <workspace-link
              class="shorten"
              disable-select
              :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: item.txhash, blockchainMode: getSettings.blockchainMode }}"
              @clicked="model = false"
            >
              {{ item.txhash }}
            </workspace-link>
          </template>
          <template #item.dateTime="{ item }">
            <span>{{ item.dateTime.toLocaleString() }}</span>
          </template>
          <template #item.txtype="{ item }">
            <span>{{ capitalize(item.txtype) }}</span>
          </template>
        </v-data-table>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';
import {capitalize} from '../../../utilities/index.js';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';

const {getSettings} = storeToRefs(useLocalStore());

defineProps({
	transactions: {type: Array, required: true},
	headers: {type: Array, required: true},
	startDate: {type: String, required: true},
	endDate: {type: String, required: true},
});

const model = defineModel({type: Boolean});

</script>

<style scoped>

</style>
