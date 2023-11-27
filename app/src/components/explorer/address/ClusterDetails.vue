<template>
  <v-row>
    <v-col
      cols="12"
      sm="6"
    >
      <v-list-item :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,params: { id: txHash }}">
        <v-list-item-title>
          Transaction Hash
        </v-list-item-title>
        <v-list-item-subtitle>
          {{ txHash }}
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
  <v-row>
    <v-col
      cols="12"
      sm="6"
    >
      <v-list-item :to="{ name: ROUTE_NAME_BLOCK_PAGE, params: { id: blockHash }}">
        <v-list-item-title>
          Block Hash
        </v-list-item-title>
        <v-list-item-subtitle>
          {{ blockHash }}
        </v-list-item-subtitle>
      </v-list-item>
    </v-col>
    <v-col>
      <v-list-item :to="{ name: ROUTE_NAME_BLOCK_PAGE, params: { id: blockId }}">
        <v-list-item-title>
          Block Id
        </v-list-item-title>
        <v-list-item-subtitle>
          {{ blockId }}
        </v-list-item-subtitle>
      </v-list-item>
    </v-col>
  </v-row>
</template>

<script setup>
import {ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';

defineProps({
	txHash: {type: String, required: true},
	blockHash: {type: String, required: true},
	blockId: {type: Number, required: true},
	timestamp: {type: String, required: true},
});

</script>

<style scoped>

</style>
