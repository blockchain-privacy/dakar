<template>
  <v-text-field
    v-model="query"
    style="margin: 23px 0 0 0;min-width:220px"
    variant="outlined"
    density="compact"
    color="primary"
    single-line
    label="Search for blocks, transactions and addresses"
    :append-inner-icon="icon.mdiMagnify"
    :rules="[isValidQuery]"
    @click:append-inner="handleInput(query)"
    @keydown.enter="handleInput(query)"
  />
</template>

<script>
import {
	mdiMagnify,
} from '@mdi/js';
import * as Constants from '@/constants';
import {handleError, isValidQuery, isValidQueryInput} from '@/utilities';

export default {
	name: 'QueryInput',
	data() {
		return {
			// Query is not managed by the vuex store
			// as it only needs to be accessed by this component
			query: '',
			lastQuery: '',
			icon: {
				mdiMagnify,
			},
		};
	},
	computed: {
		searchResultType: {
			get() {
				return this.$store.getters.getSearchResultType;
			},
		},
		isPushFromUserInput: {
			async set(value) {
				await this.$store.dispatch('setPushFromUserInput', value);
			},
			get() {
				return this.$store.getters.getPushFromUserInput;
			},
		},
	},
	watch: {
		$route() {
			this.lastQuery = '';
			this.newRouting(this);
		},
	},
	created() {
		this.newRouting(this);
	},
	methods: {
		isValidQuery,
		newRouting() {
			const {id} = this.$route.params;
			const pushFromUserInput = this.isPushFromUserInput;

			if (pushFromUserInput) {
				this.isPushFromUserInput = false;
			}

			if (pushFromUserInput || !id
        || !(this.$route.name === Constants.ROUTE_NAME_BLOCK_PAGE
          || this.$route.name === Constants.ROUTE_NAME_ADDRESS_PAGE
          || this.$route.name === Constants.ROUTE_NAME_TRANSACTION_PAGE)) {
				return;
			}

			switch (this.$route.name) {
				case Constants.ROUTE_NAME_TRANSACTION_PAGE:
					this.handleQuery(id, Constants.RESPONSE_TYPE_TRANSACTION);
					break;
				case Constants.ROUTE_NAME_BLOCK_PAGE:
					this.handleQuery(id, Constants.RESPONSE_TYPE_BLOCK);
					break;
				case Constants.ROUTE_NAME_ADDRESS_PAGE:
					this.handleQuery(id, Constants.RESPONSE_TYPE_ADDRESS);
					break;
				default:
					this.handleQuery(id);
			}
		},
		async handleInput(q) {
			// Template string in case it is a number
			const query = `${q}`.trim();
			// Update route only when input is from user and query is different
			// ignore whitespace and empty queries
			// Get data for route
			if (query.length === 0 || query === this.lastQuery || !isValidQueryInput(query) || !await this.handleQuery(query)) {
				this.setWarningMessage('Input was not valid');
				return;
			}

			// Route to corresponding page
			switch (this.searchResultType) {
				case Constants.RESPONSE_EMPTY:
					await this.$router.push({name: Constants.ROUTE_NAME_NO_RESULTS});
					break;
				case Constants.RESPONSE_TYPE_ADDRESS:
					this.isPushFromUserInput = true;
					await this.$router.push({name: Constants.ROUTE_NAME_ADDRESS_PAGE, params: {id: query}});
					break;
				case Constants.RESPONSE_TYPE_BLOCK:
					this.isPushFromUserInput = true;
					await this.$router.push({name: Constants.ROUTE_NAME_BLOCK_PAGE, params: {id: query}});
					break;
				case Constants.RESPONSE_TYPE_TRANSACTION:
					this.isPushFromUserInput = true;
					await this.$router.push({name: Constants.ROUTE_NAME_TRANSACTION_PAGE, params: {id: query}});
					break;
				default:
					await this.$router.push({name: Constants.ROUTE_NAME_NO_RESULTS});
					break;
			}
		},
		async storeResult(promise, action) {
			try {
				const response = await promise;
				this.$store.dispatch(action, response);
			} catch (e) {
				handleError(this, e);
			}
		},
		async handleQuery(q, type) {
			this.query = '';
			// Template string in case it is a number
			const query = `${q}`.trim();

			if (this.lastQuery !== '' && this.lastQuery === query) {
				return false;
			}

			this.lastQuery = query;

			if (!isValidQueryInput(query)) {
				this.setWarningMessage('Input was not valid');
				return false;
			}

			await this.$store.dispatch('resetMessages');
			await this.$store.dispatch('setBlockData', null);
			await this.$store.dispatch('setTransactionData', null);
			await this.$store.dispatch('setAddressData', null);

			switch (type) {
				case Constants.RESPONSE_TYPE_TRANSACTION:
					await this.storeResult(this.dakar.data.txHashGet({hash: query}), 'updateTransactionData');
					break;
				case Constants.RESPONSE_TYPE_BLOCK:
					await this.storeResult(this.dakar.data.blkHashGet({hash: query}), 'updateBlockData');
					break;
				case Constants.RESPONSE_TYPE_ADDRESS:
					await this.storeResult(this.dakar.data.addressHashGet({hash: query}), 'updateAddressData');
					break;
				default:
					await this.storeResult(this.dakar.data.searchQueryGet({query}), 'updateSearchResult');
			}

			return true;
		},
		setWarningMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'warning', temporary: true, category: this.$route.name});
		},
	},
};
</script>

<style scoped>

</style>
