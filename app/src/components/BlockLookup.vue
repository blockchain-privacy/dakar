<template>
  <v-container class="fill-height" fluid v-if="data">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <v-card class="elevation-12">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title>
              <v-icon>mdi-file-tree</v-icon>
              Block {{ data.blockhash }}
            </v-toolbar-title>
          </v-toolbar>
          <v-card-text>
            <v-container>
              <v-row>
                <v-col v-if="data.id">
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-format-list-numbered</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>Block Height</v-list-item-title>
                      <v-list-item-subtitle>
                        {{ data.id }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
                <v-col v-if="data.ts">
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi mdi-calendar</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>Timestamp</v-list-item-title>
                      <v-list-item-subtitle>
                        {{ data.ts != null ? new Date(data.ts).toLocaleString() : "" }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
              </v-row>
              <v-row>
                <v-col v-if="data.prevblockhash">
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-format-header-pound</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>Previous Block</v-list-item-title>
                      <v-list-item-subtitle>
                        <router-link :to="data.prevblockhash">{{ shortenHash(data.prevblockhash) }}</router-link>
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
                <v-col>
                  <v-list-item>
                    <v-list-item-avatar v-if="data.nextblockhash">
                      <v-icon>mdi-format-header-pound</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content v-if="data.nextblockhash">
                      <v-list-item-title>Next Block</v-list-item-title>
                      <v-list-item-subtitle>
                        <router-link :to="data.nextblockhash">{{ shortenHash(data.nextblockhash) }}</router-link>
                      </v-list-item-subtitle>
                    </v-list-item-content>
                  </v-list-item>
                </v-col>
              </v-row>
              <v-divider v-if="data.txhashes"></v-divider>
              <v-row v-if="data.txhashes">
                <v-col v-for="tx in data.txhashes" v-bind:key="tx">
                  <v-list-item>
                    <v-list-item-avatar>
                      <v-icon>mdi-transfer</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                      <v-list-item-title>Transaction</v-list-item-title>
                      <v-list-item-subtitle>
                        Hash:
                        <router-link :to="tx">{{ shortenHash(tx) }}</router-link>
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

export default {
  name: 'BlockLookup',
  methods: {
    shortenHash,
  },
  computed: {
    data() {
      return this.$store.getters.getBlockData;
    }
  }
}
</script>
