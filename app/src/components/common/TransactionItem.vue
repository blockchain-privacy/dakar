<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-card>
    <icon-title
      :title="`Transaction ${tx.txhash}`"
      :icon="mdiTransfer"
      :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: tx.txhash, blockchainMode: route.params.blockchainMode }}"
    >
      <privacy-chip
        v-if="tx.txtype"
        class="mx-3"
        :transaction-type="tx.txtype"
      />
    </icon-title>
    <v-card-text>
      <v-list>
        <v-list-item :to="{ name: ROUTE_NAME_BLOCK_PAGE, params: { id: tx.bhash, blockchainMode: route.params.blockchainMode }}">
          <v-list-item-title>
            Block Hash
          </v-list-item-title>
          <v-list-item-subtitle class="shorten">
            {{ tx.bhash }}
          </v-list-item-subtitle>
        </v-list-item>
        <v-list-item :to="{ name: ROUTE_NAME_BLOCK_PAGE, params: { id: tx.bid, blockchainMode: route.params.blockchainMode }}">
          <v-list-item-title>
            Block ID
          </v-list-item-title>
          <v-list-item-subtitle>
            {{ tx.bid }}
          </v-list-item-subtitle>
        </v-list-item>
      </v-list>
    </v-card-text>
  </v-card>
</template>

<script setup>
import PrivacyChip from '@/components/common/PrivacyChip.vue';
import {ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import IconTitle from '@/components/common/IconTitle.vue';
import {mdiTransfer} from '@mdi/js';
import {useRoute} from 'vue-router';

defineProps({tx: {type: Object, required: true}});
const route = useRoute();

</script>
<style scoped>

</style>
