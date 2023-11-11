<template>
  <v-container :fluid="true">
    <v-row
      v-if="data"
      align="center"
      justify="center"
    >
      <!-- duplicate transaction hashes can exist -> loop through all results
      (e.g. d5d27987d2a3dfc724e359870c6644b40e497bdc0589a033220fe15429d88599 in Bitcoin) -->
      <v-col
        v-for="tx in data"
        :key="tx.txhash+tx.bid"
        cols="12"
        sm="12"
        md="12"
        lg="10"
        xl="8"
      >
        <Transaction
          :tx="tx"
          :show-heuristic-editor-link="isAtLeastPrivileged"
          :show-fingerprint-link="isAtLeastPrivileged"
          show-details
        />
      </v-col>
    </v-row>
    <v-skeleton-loader
      v-else
      class="mx-auto"
      type="list-item-three-line, list-item-three-line, list-item-three-line"
    />
  </v-container>
</template>

<script>
import Transaction from './Transaction.vue';
import {PAGE_TITLE} from '@/constants';
import {isAdminIdentity, isPrivilegedIdentity} from '@/utilities';

export default {
	name: 'TransactionPage',
	components: {Transaction},
	computed: {
		data() {
			return this.$store.getters.getTransactionData;
		},
		session() {
			return this.$store.getters.getSession;
		},
		isAtLeastPrivileged() {
			return isPrivilegedIdentity(this.session) || isAdminIdentity(this.session);
		},
	},
	watch: {
		data() {
			this.setPageTitle();
		},
	},
	mounted() {
		this.setPageTitle();
	},
	updated() {
		this.setPageTitle();
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
};
</script>
