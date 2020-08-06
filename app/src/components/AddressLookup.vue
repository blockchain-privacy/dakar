<template>
  <v-layout row v-if="data">
    <v-flex xs12 sm8 offset-sm2>
      <v-card>
        <v-card-title>
          <v-icon>mdi-bank-transfer</v-icon>
          Address
        </v-card-title>
        <v-list two-line subheader>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon>mdi-format-header-pound</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Hash</v-list-item-title>
              <v-list-item-subtitle>{{ data.addresshash }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-list-item v-for="o in data.addr_outputs" v-bind:key="o.input_transaction + o.output_transaction">
            <v-list-item-avatar>
              <v-icon></v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Output</v-list-item-title>
              <v-list-item-subtitle>Amount: {{ o.amount }}</v-list-item-subtitle>
              <v-list-item-subtitle v-if="o.iscoinbase" :data="o.iscoinbase">Coinbase:
                {{ o.iscoinbase }}
              </v-list-item-subtitle>
              <v-list-item-subtitle>Output Transaction:
                <router-link :to="o.output_transaction">{{ o.output_transaction }}</router-link>
              </v-list-item-subtitle>
              <v-list-item-subtitle v-if="o.input_transaction">Input transaction:
                <router-link :to="o.input_transaction">
                  {{ o.input_transaction }}
                </router-link>
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
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
  name: 'AddressLookup',
  computed: {
    data() {
      return this.$store.getters.getAddressData;
    }
  }
}
</script>
