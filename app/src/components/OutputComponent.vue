<template>
  <v-sheet min-height="50" class="fill-height" color="transparent">
    <v-lazy min-height="90" transition="fade-transition" :options="{threshold: 0.7}">
      <v-card flat max-width="470px">
        <v-row align="center" justify="center">
          <v-col class="text-caption" style="text-align: center">
            {{ convertAmount(address.amount) }}
            DASH
          </v-col>
          <v-col>
            <v-card
                class="mx-auto"
                max-width="400px"
                outlined>
              <v-list-item>
                <v-list-item-avatar>
                  <v-icon>mdi-bank-transfer-in</v-icon>
                </v-list-item-avatar>
                <v-list-item-content>
                  <router-link :to="{ name: transactionRoute,
                        params: { id: address.output_transaction }}" class="tx">
                    {{ shortenHash(address.output_transaction) }}
                  </router-link>
                </v-list-item-content>
              </v-list-item>
            </v-card>
            <v-row v-if="address.input_transaction">
              <v-col style="text-align: center">
                <v-icon>mdi-arrow-down</v-icon>
              </v-col>
            </v-row>
            <v-card v-if="address.input_transaction"
                    class="mx-auto"
                    max-width="400px"
                    outlined>
              <v-list-item>
                <v-list-item-avatar>
                  <v-icon>mdi-bank-transfer-out</v-icon>
                </v-list-item-avatar>
                <v-list-item-content>
                  <router-link :to="{ name: transactionRoute,
                        params: { id: address.input_transaction }}" class="tx">
                    {{ shortenHash(address.input_transaction) }}
                  </router-link>
                </v-list-item-content>
              </v-list-item>
            </v-card>
          </v-col>
        </v-row>
      </v-card>
    </v-lazy>
  </v-sheet>
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
