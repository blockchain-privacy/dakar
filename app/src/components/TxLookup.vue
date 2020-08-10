<template>
  <v-container class="fill-height" fluid v-if="data">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <v-card class="elevation-12">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title>
              <v-icon>mdi-transfer</v-icon>
              Transaction {{ data.txhash }}
            </v-toolbar-title>
          </v-toolbar>
          <v-card-text>
            <v-container>
              <v-row>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-calendar</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>Timestamp</v-list-item-title>
                      <v-list-item-subtitle>{{ new Date(data.bts).toLocaleString() }}</v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-pound</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>Block Id</v-list-item-title>
                      <v-list-item-subtitle>
                        <router-link :to="data.bid.toString()"> {{ data.bid }}</router-link>
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-format-header-pound</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>Block</v-list-item-title>
                      <v-list-item-subtitle>
                        <router-link :to="data.bhash">{{ shortenHash(data.bhash) }}</router-link>
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-pound</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>Number of outputs</v-list-item-title>
                      <v-list-item-subtitle v-if="!data.outputs">0</v-list-item-subtitle>
                      <v-list-item-subtitle v-if="data.outputs">{{ data.outputs.length }}</v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-pound</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>Number of inputs</v-list-item-title>
                      <v-list-item-subtitle v-if="!data.inputs">0</v-list-item-subtitle>
                      <v-list-item-subtitle v-if="data.inputs">{{ data.inputs.length }}</v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
              </v-row>
              <v-divider v-if="outputs"></v-divider>
              <v-row v-if="outputs">
                <v-col v-for="i in outputs" v-bind:key="i.addresshash">
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-currency-usd-circle-outline</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>Output</v-list-item-title>
                      <v-list-item-subtitle>
                        Address hash:
                        <router-link :to="i.addresshash">{{ i.addresshash }}</router-link>
                        <br>
                        Amount: {{ i.amount }}<br>
                        Spent: {{ i.inputindex != null }}<br>
                        Index: {{ i.outputindex }}<br>
                        Coinbase: {{ i.iscoinbase }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
              </v-row>
              <v-divider v-if="inputs"></v-divider>
              <v-row v-if="inputs">
                <v-col v-for="i in inputs" v-bind:key="i.addresshash">
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-currency-usd-circle</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>Input</v-list-item-title>
                      <v-list-item-subtitle>
                        Hash:
                        <router-link :to="i.addresshash">{{ i.addresshash }}</router-link>
                        <br>
                        Amount: {{ i.amount }}<br>
                        Index: {{ i.inputindex }}<br>
                        Coinbase: {{ i.iscoinbase }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
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
  name: 'TxLookup',
  computed: {
    data() {
      return this.$store.getters.getTransactionData;
    },
    inputs() {
      return this.sortByInput(this.data.inputs);
    },
    outputs() {
      return this.sortByOutput(this.data.outputs);
    }
  },
  methods: {
    shortenHash,
    sortByOutput(outputs) {
      if (outputs == null) return null;
      return outputs.sort((a, b) => {
        return a.outputindex > b.outputindex
      })
    },
    sortByInput(inputs) {
      if (inputs == null) return null;
      return inputs.sort((a, b) => {
        return a.inputindex > b.inputindex
      })
    }
  },
  mounted() {
    document.title = `Transaction - ${PAGE_TITLE}`;
  },
  updated() {
    let h = ' ';
    if (this.data.txhash) {
      h = ` ${this.data.txhash} `
    }
    document.title = `Transaction${h}- ${PAGE_TITLE}`;
  },
}
</script>
