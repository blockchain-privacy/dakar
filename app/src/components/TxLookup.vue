<template>
  <v-layout row v-if="data">
    <v-flex xs12 sm8 offset-sm2>
      <v-card>
        <v-card-title>
          <v-icon>mdi-bank-transfer</v-icon>
          Transaction
        </v-card-title>
        <v-list two-line subheader>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon>mdi-format-header-pound</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Hash</v-list-item-title>
              <v-list-item-subtitle>{{ data.txhash }}</v-list-item-subtitle>
            </v-list-item-content>
            <v-list-item-avatar>
              <v-icon>mdi-calendar</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Timestamp</v-list-item-title>
              <v-list-item-subtitle>{{ new Date(data.bts).toLocaleString() }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon>mdi-pound</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Block Id</v-list-item-title>
              <v-list-item-subtitle>{{ data.bid }}</v-list-item-subtitle>
            </v-list-item-content>
            <v-list-item-avatar>
              <v-icon>mdi-format-header-pound</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Block</v-list-item-title>
              <v-list-item-subtitle>
                <router-link :to="data.bhash">{{ data.bhash }}</router-link>
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon></v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Number of outputs</v-list-item-title>
              <v-list-item-subtitle v-if="!data.outputs">0</v-list-item-subtitle>
              <v-list-item-subtitle v-if="data.outputs">{{ data.outputs.length }}</v-list-item-subtitle>
            </v-list-item-content>
            <v-list-item-avatar>
              <v-icon></v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Number of inputs</v-list-item-title>
              <v-list-item-subtitle v-if="!data.inputs">0</v-list-item-subtitle>
              <v-list-item-subtitle v-if="data.inputs">{{ data.inputs.length }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <div v-if="outputs">
            <v-list-item v-for="i in outputs" v-bind:key="i.addresshash">
              <v-list-item-avatar>
                <v-icon></v-icon>
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
          </div>
          <div v-if="inputs">
            <v-list-item v-for="i in inputs" v-bind:key="i.addresshash">
              <v-list-item-avatar>
                <v-icon></v-icon>
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
          </div>

        </v-list>
      </v-card>
      <v-card><!-- TODO hack, to leave space at the bottom, such that content is not hidden by the footer.-->
        <div style="height: 40pt">
        </div>
      </v-card>
    </v-flex>
  </v-layout>
</template>

<script>

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
  }
}
</script>
