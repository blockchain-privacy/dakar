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
          <store-link
            class="shorten"
            disable-select
            :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,params: { id: txHash }}"
          >
            {{ txHash }}
          </store-link>
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
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import StoreLink from '@/components/common/WorkspaceLink.vue';

defineProps({
	txHash: {type: String, required: true},
	timestamp: {type: String, required: true},
});

</script>

<style scoped>

</style>
