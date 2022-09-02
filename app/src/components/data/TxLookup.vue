<template>
  <v-container fluid v-if="data">
    <v-row align="center" justify="center">
      <!-- duplicate transaction hashes can exist -> loop through all results
      (e.g. d5d27987d2a3dfc724e359870c6644b40e497bdc0589a033220fe15429d88599 in Bitcoin) -->
      <v-col cols="12" sm="12" md="12" lg="10" xl="8" v-for="tx in data" :key="tx.txhash+tx.bid">
        <Transaction :tx="tx" :show-heuristic-editor-link="showHeuristicEditor" show-details/>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import Transaction from './Transaction.vue';
import { PAGE_TITLE } from '../../constants';
import { isAdminIdentity, isPrivilegedIdentity } from '../../utilities';

export default {
  name: 'TxLookup',
  components: { Transaction },
  computed: {
    data() {
      return this.$store.getters.getTransactionData;
    },
    session() {
      return this.$store.getters.getSession;
    },
    showHeuristicEditor() {
      return isPrivilegedIdentity(this.session) || isAdminIdentity(this.session);
    },
  },
  methods: {
    setPageTitle() {
      let h = ' ';
      if (this.data && this.data[0].txhash) {
        h = ` ${this.data[0].txhash} `;
      }
      document.title = `Transaction${h}- ${PAGE_TITLE}`;
    },
  },
  mounted() {
    this.setPageTitle();
  },
  updated() {
    this.setPageTitle();
  },
};
</script>
