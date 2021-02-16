<template>
  <v-card outlined max-width="470px">
    <v-container>
      <v-row align="center" justify="center">
        <v-col class="text--accent-1" style="text-align: center">
          {{ convertAmount(address.amount) }}
          {{ this.coinUnit }}
          <v-icon class="itemIcon" :id="`a${index}`" v-if="address.is_coinbase">
            {{ icon.mdiPickaxe }}
          </v-icon>
          <v-tooltip bottom :activator="`#a${index}`" v-if="address.is_coinbase">
            <span>Coinbase</span>
          </v-tooltip>
        </v-col>
        <v-col>
          <v-card
              class="mx-auto"
              max-width="400px"
              flat>
            <v-list-item>
              <v-list-item-avatar style="margin: 0">
                <v-icon class="itemIcon">{{ icon.mdiBankTransferIn }}</v-icon>
              </v-list-item-avatar>
              <v-list-item-content>
                <router-link :to="{ name: transactionRoute,
                        params: { id: address.output_transaction }}" class="tx">
                  <div :id="`b${index}`">{{ shortenHash(address.output_transaction) }}</div>
                </router-link>
              </v-list-item-content>
              <v-tooltip bottom :activator="`#b${index}`">
                <span>{{ new Date(address.output_ts).toLocaleString() }}</span>
              </v-tooltip>
            </v-list-item>
          </v-card>
          <v-container>
            <v-row v-if="address.input_transaction">
              <v-col style="text-align: center; padding: 0">
                <v-icon class="itemIcon">{{ icon.mdiArrowDown }}</v-icon>
              </v-col>
            </v-row>
          </v-container>
          <v-card v-if="address.input_transaction"
                  class="mx-auto"
                  max-width="400px"
                  flat>
            <v-list-item>
              <v-list-item-avatar style="margin: 0">
                <v-icon class="itemIcon">{{ icon.mdiBankTransferOut }}</v-icon>
              </v-list-item-avatar>
              <v-list-item-content>
                <router-link :to="{ name: transactionRoute,
                        params: { id: address.input_transaction }}" class="tx">
                  <div :id="`c${index}`">{{ shortenHash(address.input_transaction) }}</div>
                </router-link>
              </v-list-item-content>
              <v-tooltip bottom :activator="`#c${index}`">
                <span>{{ new Date(address.input_ts).toLocaleString() }}</span>
              </v-tooltip>
            </v-list-item>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </v-card>
</template>

<script>
import {
  mdiPickaxe, mdiBankTransferIn, mdiArrowDown, mdiBankTransferOut,
} from '@mdi/js';
import { shortenHash, convertAmount } from '../../utilities';
import { ROUTE_NAME_TRANSACTION_PAGE, COIN_UNIT } from '../../constants';

export default {
  name: 'OutputComponent',
  props: {
    address: { type: Object, required: true },
    index: { type: Number, required: true },
  },
  data() {
    return {
      transactionRoute: ROUTE_NAME_TRANSACTION_PAGE,
      coinUnit: COIN_UNIT,
      icon: {
        mdiPickaxe, mdiBankTransferIn, mdiArrowDown, mdiBankTransferOut,
      },
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

.itemIcon {
  max-width: 32px;
}
</style>
