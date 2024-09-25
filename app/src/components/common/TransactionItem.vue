<template>
  <v-card>
    <icon-title
      :title="`Transaction ${tx.txhash}`"
      :icon="mdiTransfer"
      :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: tx.txhash }}"
    >
      <privacy-chip
        v-if="tx.txtype"
        class="mx-3"
        :transaction-type="tx.txtype"
      />
    </icon-title>
    <v-card-text>
      <v-list>
        <v-list-item :to="{ name: ROUTE_NAME_BLOCK_PAGE, params: { id: tx.bhash }}">
          <v-list-item-title>
            Block Hash
          </v-list-item-title>
          <v-list-item-subtitle class="shorten">
            {{ tx.bhash }}
          </v-list-item-subtitle>
        </v-list-item>
        <v-list-item :to="{ name: ROUTE_NAME_BLOCK_PAGE, params: { id: tx.bid }}">
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

defineProps({tx: {type: Object, required: true}});

</script>
<style scoped>

</style>
