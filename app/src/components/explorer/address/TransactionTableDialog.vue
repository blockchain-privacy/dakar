<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="1200px"
  >
    <v-card class="pb-2">
      <v-card-title>
        <div class="text-wrap">
          Transactions from {{ startDate }} to {{ endDate }}
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
              disable-select
              :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: item.txhash, blockchainMode: route.params.blockchainMode }}"
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
import {useRoute} from 'vue-router';
import {capitalize} from '../../../utilities/index.js';
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';

defineProps({
	transactions: {type: Array, required: true},
	headers: {type: Array, required: true},
	startDate: {type: String, required: true},
	endDate: {type: String, required: true},
});

const route = useRoute();
const model = defineModel({type: Boolean});

</script>

<style scoped>

</style>
