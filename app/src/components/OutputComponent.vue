<template>

  <v-card outlined max-width="470px">
    <v-row align="center" justify="center">
      <v-col class="text--accent-1" style="text-align: center">
        {{ convertAmount(address.amount) }}
        DASH
      </v-col>
      <v-col>
        <v-card
            class="mx-auto"
            max-width="400px"
            flat>
          <v-list-item>
            <v-list-item-avatar style="margin: 0">
              <v-icon>mdi-bank-transfer-in</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-tooltip bottom>
                <template v-slot:activator="{ on }">
                  <router-link :to="{ name: transactionRoute,
                        params: { id: address.output_transaction }}" class="tx">
                    <div v-on="on">{{ shortenHash(address.output_transaction) }}</div>
                  </router-link>
                </template>
                <span>{{ new Date(address.output_ts).toLocaleString() }}</span>
              </v-tooltip>
            </v-list-item-content>
          </v-list-item>
        </v-card>
        <v-row v-if="address.input_transaction">
          <v-col style="text-align: center; padding: 0">
            <v-icon>mdi-arrow-down</v-icon>
          </v-col>
        </v-row>
        <v-card v-if="address.input_transaction"
                class="mx-auto"
                max-width="400px"
                flat>
          <v-list-item>
            <v-list-item-avatar style="margin: 0">
              <v-icon>mdi-bank-transfer-out</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-tooltip bottom>
                <template v-slot:activator="{ on }">
                  <router-link small :to="{ name: transactionRoute,
                        params: { id: address.input_transaction }}" class="tx">
                    <div v-on="on"> {{ shortenHash(address.input_transaction) }}</div>
                  </router-link>
                </template>
                <span>{{ new Date(address.input_ts).toLocaleString() }}</span>
              </v-tooltip>
            </v-list-item-content>
          </v-list-item>
        </v-card>
      </v-col>
    </v-row>
  </v-card>

</template>

<script>

import { shortenHash, convertAmount } from '../utilities';
import { ROUTE_NAME_TRANSACTION_PAGE } from '../constants';

export default {
  name: 'OutputComponent',
  props: {
    address: { type: Object, required: true },
  },
  data() {
    return {
      transactionRoute: ROUTE_NAME_TRANSACTION_PAGE,
    };
  },
  methods: {
    shortenHash,
    convertAmount,
  },
};
</script>

<style>
.tx {
  font-size: 0.75rem !important;
  text-transform: uppercase;
  font-family: "Roboto", sans-serif !important;
}
</style>
