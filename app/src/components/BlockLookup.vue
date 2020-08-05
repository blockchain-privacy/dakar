<template>
  <v-layout row v-if="data">
    <v-flex xs12 sm8 offset-sm2>
      <v-card>
        <v-card-title>
          <v-icon>mdi-bank-transfer</v-icon>
          Block
        </v-card-title>
        <v-list two-line subheader>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon>mdi-format-header-pound</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Hash</v-list-item-title>
              <v-list-item-subtitle>
                {{ data.hash }}
              </v-list-item-subtitle>
            </v-list-item-content>
            <v-list-item-avatar>
              <v-icon>mdi mdi-calendar-month-outline </v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Timestamp</v-list-item-title>
              <v-list-item-subtitle>
                {{ new Date(data.ts).toLocaleString() }}
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon>mdi-pound</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>ID</v-list-item-title>
              <v-list-item-subtitle>
                {{ data.id }}
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon>mdi-format-header-pound</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Previous Block</v-list-item-title>
              <v-list-item-subtitle>
                <router-link :to="data.prevblockhash">{{ data.prevblockhash }}</router-link>
              </v-list-item-subtitle>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>Next Block</v-list-item-title>
              <v-list-item-subtitle>
                <router-link :to="data.nextblockhash">{{ data.nextblockhash }}</router-link>
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-list-item v-for="tx in data.txhashes" v-bind:key="tx">
            <v-list-item-avatar>
              <v-icon></v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Transaction</v-list-item-title>
              <v-list-item-subtitle>
                Hash:
                <router-link :to="tx">{{ tx }}</router-link>
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
  name: 'BlockLookup',
  computed: {
    data() {
      return this.$store.getters.getBlockData;
    }
  }
}
</script>
