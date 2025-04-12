import {defineStore} from 'pinia';

export const useExplorerStore = defineStore('explorer', {
	state: () => ({
		highlightWasabi2Denominations: false,
	}),
	getters: {
		getHighlightWasabi2Denominations: state => state.highlightWasabi2Denominations,
	},
	actions: {
		setHighlightWasabi2Denominations(payload) {
			this.highlightWasabi2Denominations = payload;
		},
	},
});
