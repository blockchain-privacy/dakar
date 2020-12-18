<template>
  <v-container class="fill-height" fluid v-if="data">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <v-card class="elevation-12">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title>
              <v-icon>mdi-cube-outline</v-icon>
              Block {{ data.blockhash }}
            </v-toolbar-title>
          </v-toolbar>
          <v-card-text>
            <v-container>
              <v-row>
                <v-col v-if="data.id">
                  <IconItem icon="mdi-format-list-numbered"
                            title="Block Height" :subtitle="data.id">
                    {{ data.id }}
                  </IconItem>
                </v-col>
                <v-col v-if="data.ts">
                  <IconItem icon="mdi-calendar" title="Timestamp">
                    {{ data.ts != null ? new Date(data.ts).toLocaleString() : "" }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col v-if="data.prevblockhash">
                  <IconItem icon="mdi-format-header-pound" title="Previous Block">
                    <router-link :to="data.prevblockhash">
                      {{ shortenHash(data.prevblockhash) }}
                    </router-link>
                  </IconItem>
                </v-col>
                <v-col v-if="data.nextblockhash">
                  <IconItem icon="mdi-format-header-pound" title="Next Block">
                    <router-link :to="data.nextblockhash">
                      {{ shortenHash(data.nextblockhash) }}
                    </router-link>
                  </IconItem>
                </v-col>
              </v-row>
              <v-divider v-if="data.txhashes"></v-divider>
              <v-row v-if="data.txhashes">
                <v-col v-for="tx in data.txhashes" v-bind:key="tx">
                  <IconItem icon="mdi-transfer" title="Transaction">
                    <router-link :to="tx">{{ shortenHash(tx) }}</router-link>
                  </IconItem>
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
import { shortenHash } from '../utilities';
import { PAGE_TITLE } from '../constants';
import IconItem from './common/IconItem.vue';

export default {
  name: 'BlockLookup',
  components: { IconItem },
  methods: {
    shortenHash,
  },
  computed: {
    data() {
      return this.$store.getters.getSearchResult.payload;
    },
  },
  mounted() {
    document.title = `Block - ${PAGE_TITLE}`;
  },
  updated() {
    let id = ' ';
    if (this.data && this.data.id) {
      id = ` ${this.data.id} `;
    }
    document.title = `Block${id}- ${PAGE_TITLE}`;
  },
};
</script>
