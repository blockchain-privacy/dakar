<template>
  <v-container class="fill-height" fluid v-if="data">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <v-card class="elevation-12">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title>
              <v-icon>mdi-card-bulleted-outline</v-icon>
              Address {{ data.addresshash }}
            </v-toolbar-title>
          </v-toolbar>
          <v-card-text>
            <v-container>
              <v-row>
                <v-col>
                  <IconItem icon="mdi-scale-balance" title="Balance">
                    {{ this.amounts.received - this.amounts.spent }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <IconItem icon="mdi-bank-transfer-in" title="Total amount received">
                    {{ this.amounts.received }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem icon="mdi-bank-transfer-out" title="Total amount spent">
                    {{ this.amounts.spent }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-divider></v-divider>
              <v-row>
                <v-col v-for="o in data.addr_outputs" v-bind:key="o.input_transaction + o.output_transaction">
                  <v-sheet min-height="50" class="fill-height" color="transparent">
                    <v-lazy min-height="90" transition="fade-transition" :options="{threshold: 1}">
                      <IconItem icon="mdi-currency-usd-circle-outline" title="Output">
                        Amount: {{ o.amount }}
                        <br v-if="o.iscoinbase"/>
                        {{ o.iscoinbase ? 'Coinbase: ' + o.iscoinbase : '' }}
                        <br/>
                        Output Transaction:
                        <router-link :to="o.output_transaction">
                          {{ shortenHash(o.output_transaction) }}
                        </router-link>
                        <br v-if="o.input_transaction"/>
                        {{ o.input_transaction ? 'Input transaction:' : '' }}
                        <router-link :to="o.input_transaction" v-if="o.input_transaction">
                          {{ shortenHash(o.input_transaction) }}
                        </router-link>
                      </IconItem>
                    </v-lazy>
                  </v-sheet>
                </v-col>
              </v-row>
            </v-container>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import {shortenHash} from "@/utilities";
import {PAGE_TITLE} from "@/constants";
import IconItem from "@/components/common/IconItem";

export default {
  name: 'AddressLookup',
  components: {IconItem},
  methods: {
    shortenHash,
    calculateAmountReceived: function (outputs) {
      return outputs
          .map(e => parseFloat(e.amount))
          .reduce((sum, e) => sum + e, 0);
    },
    calculateAmountSpent: function (outputs) {
      return outputs
          .filter(e => e.input_transaction !== '')
          .map(e => parseFloat(e.amount))
          .reduce((sum, e) => sum + e, 0);
    }
  },
  computed: {
    data() {
      return this.$store.getters.getAddressData;
    },
    amounts() {
      return {
        received: this.calculateAmountReceived(this.data.addr_outputs),
        spent: this.calculateAmountSpent(this.data.addr_outputs),
      };
    }
  },
  mounted() {
    document.title = `Address - ${PAGE_TITLE}`;
  },
  updated() {
    let h = ' ';
    if (this.data && this.data.addresshash) {
      h = ` ${this.data.addresshash} `
    }
    document.title = `Address${h}- ${PAGE_TITLE}`;
  },
}
</script>
