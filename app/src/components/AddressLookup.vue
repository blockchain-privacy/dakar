<template>
  <v-container class="fill-height" fluid v-if="data">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <v-card class="elevation-12">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title>
              <v-icon>mdi-card-bulleted-outline</v-icon>
              Address {{ this.data.addresshash }}
            </v-toolbar-title>
          </v-toolbar>
          <v-card-text>
            <v-container>
              <v-row>
                <v-col>
                  <IconItem icon="mdi-scale-balance" title="Balance">
                    {{ convertAmount(this.amounts.received - this.amounts.spent) }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <IconItem icon="mdi-bank-transfer-in" title="Total amount received">
                    {{ convertAmount(this.amounts.received) }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem icon="mdi-bank-transfer-out" title="Total amount spent">
                    {{ convertAmount(this.amounts.spent) }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-divider></v-divider>
              <v-row>
                <v-col>
                  <v-select
                      style="max-width: 300px;"
                      v-model="combobox.selected"
                      :items="combobox.items"
                      label="Sort addresses by"
                      v-on:change="sortItems"
                  ></v-select>
                </v-col>
              </v-row>
              <v-sheet min-height="50" class="fill-height" color="transparent">
                <v-lazy min-height="90" transition="fade-transition" :options="{threshold: 0.7}">
                  <v-row>

                    <v-col v-for="o in outputData"
                           v-bind:key="o.input_transaction + o.output_transaction + o.amount">
                      <OutputComponent :address="o"/>
                    </v-col>
                  </v-row>
                </v-lazy>
              </v-sheet>
            </v-container>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import OutputComponent from './OutputComponent.vue';
import { convertAmount } from '../utilities';
import { PAGE_TITLE, ROUTE_NAME_TRANSACTION_PAGE } from '../constants';
import IconItem from './common/IconItem.vue';

export default {
  name: 'AddressLookup',
  components: { IconItem, OutputComponent },
  methods: {
    sortItems() {
      switch (this.combobox.selected) {
        case 'ascending by input date':
          this.outputData = this.outputData.sort((a, b) => (a.inputDate < b.inputDate ? -1 : 1));
          break;
        case 'descending by input date':
          this.outputData = this.outputData.sort((a, b) => (a.inputDate > b.inputDate ? -1 : 1));
          break;
        case 'ascending by output date':
          this.outputData = this.outputData.sort((a, b) => (a.outputDate < b.outputDate ? -1 : 1));
          break;
        case 'descending by output date':
          this.outputData = this.outputData.sort((a, b) => (a.outputDate > b.outputDate ? -1 : 1));
          break;
        default:
      }
    },
    convertAmount,
    calculateAmountReceived(outputs) {
      if (outputs === undefined) return 0;
      return outputs
        .map((e) => parseInt(e.amount, 10))
        .reduce((sum, e) => sum + e, 0);
    },
    calculateAmountSpent(outputs) {
      if (outputs === undefined) return 0;
      return outputs
        .filter((e) => e.input_transaction !== '')
        .map((e) => parseInt(e.amount, 10))
        .reduce((sum, e) => sum + e, 0);
    },
  },
  data() {
    return {
      combobox: {
        selected: [],
        items: ['ascending by input date',
          'descending by input date',
          'ascending by output date',
          'descending by output date'],
      },
      outputData: [],
      transactionRoute: ROUTE_NAME_TRANSACTION_PAGE,
    };
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
    },
  },
  mounted() {
    document.title = `Address - ${PAGE_TITLE}`;
  },
  updated() {
    let h = ' ';
    if (this.data && this.outputData.length === 0 && this.data.addresshash) {
      h = ` ${this.data.addresshash} `;

      this.outputData = this.data.addr_outputs.map((d) => {
        d.inputDate = new Date(d.input_ts);
        d.outputDate = new Date(d.output_ts);
        return d;
      });
    }
    document.title = `Address${h}- ${PAGE_TITLE}`;
  },
};
</script>
