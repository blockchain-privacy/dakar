import {defineStore} from 'pinia';
import {RESPONSE_TYPE_ADDRESS, RESPONSE_TYPE_BLOCK, RESPONSE_TYPE_TRANSACTION} from '@/constants';

export const useExplorerStore = defineStore('explorer', {
	state: () => ({
		transaction: null,
		searchResultType: null,
		address: null,
		block: null,
	}),
	getters: {
		getTransaction: state => state.transaction,
		getAddress: state => state.address,
		getBlock: state => state.block,
		getSearchResultType: state => state.searchResultType,
	},
	actions: {
		setTransaction(payload) {
			this.transaction = payload;
		},
		setAddress(payload) {
			this.address = payload;
		},
		setBlock(payload) {
			this.block = payload;
		},
		updateBlock(payload) {
			this.searchResultType = payload.type;
			this.block = payload.payload;
		},
		updateTransaction(payload) {
			this.searchResultType = payload.type;
			this.transaction = payload.payload;
		},
		updateAddress(payload) {
			this.searchResultType = payload.type;
			this.address = payload.payload;
		},
		updateSearchResult(payload) {
			this.searchResultType = payload.type;
			switch (payload.type) {
				case RESPONSE_TYPE_TRANSACTION:
					this.transaction = payload.payload;
					break;
				case RESPONSE_TYPE_BLOCK:
					this.block = payload.payload;
					break;
				case RESPONSE_TYPE_ADDRESS:
					this.address = payload.payload;
					break;
				default:
					// Something went wrong, let's reset state
					this.searchResultType = null;
					this.transaction = null;
					this.block = null;
					this.address = null;
			}
		},
	},
});
