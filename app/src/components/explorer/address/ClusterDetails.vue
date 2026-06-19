<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-row>
    <v-col
      cols="12"
      sm="6"
    >
      <v-list-item>
        <v-list-item-title>
          Transaction
        </v-list-item-title>
        <v-list-item-subtitle>
          <workspace-link
            disable-select
            :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: txHash, blockchainMode: route.params.blockchainMode }}"
          >
            {{ txHash }}
          </workspace-link>
        </v-list-item-subtitle>
      </v-list-item>
    </v-col>
    <v-col>
      <v-list-item>
        <v-list-item-title>
          Timestamp
        </v-list-item-title>
        <v-list-item-subtitle>
          {{ new Date(timestamp).toLocaleString() }}
        </v-list-item-subtitle>
      </v-list-item>
    </v-col>
  </v-row>
</template>

<script setup>
import {useRoute} from 'vue-router';
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';

defineProps({
	txHash: {type: String, required: true},
	timestamp: {type: String, required: true},
});

const route = useRoute();

</script>

<style scoped>

</style>
