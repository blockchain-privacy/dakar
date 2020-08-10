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
                <v-col v-for="o in data.addr_outputs" v-bind:key="o.input_transaction + o.output_transaction">
                  <v-sheet min-height="50" class="fill-height" color="transparent">
                    <v-lazy min-height="90" transition="fade-transition" :options="{threshold: 1}">
                      <v-list-item>
                        <v-list-item-avatar>
                          <v-icon>mdi-currency-usd-circle-outline</v-icon>
                        </v-list-item-avatar>
                        <v-list-item-content>
                          <v-list-item-title>Output</v-list-item-title>
                          <v-list-item-subtitle>Amount: {{ o.amount }}</v-list-item-subtitle>
                          <v-list-item-subtitle v-if="o.iscoinbase" :data="o.iscoinbase">Coinbase:
                            {{ o.iscoinbase }}
                          </v-list-item-subtitle>
                          <v-list-item-subtitle>Output Transaction:
                            <router-link :to="o.output_transaction">
                              {{ shortenHash(o.output_transaction) }}
                            </router-link>
                          </v-list-item-subtitle>
                          <v-list-item-subtitle v-if="o.input_transaction">Input transaction:
                            <router-link :to="o.input_transaction">
                              {{ shortenHash(o.input_transaction) }}
                            </router-link>
                          </v-list-item-subtitle>
                        </v-list-item-content>
                      </v-list-item>
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

export default {
  name: 'AddressLookup',
  methods: {
    shortenHash,
  },
  computed: {
    data() {
      return this.$store.getters.getAddressData;
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
