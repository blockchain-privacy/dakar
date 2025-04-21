import {defineStore} from 'pinia';
import {deleteLocalstorageData, getLocalstorageData, setLocalstorageData} from '@/utilities';
import {
	LOCALSTORAGE_FIELD_SEARCH_HISTORY,
	LOCALSTORAGE_FIELD_SESSION,
	LOCALSTORAGE_FIELD_SETTINGS,
} from '@/constants/index.js';

// InsertLocalData inserts session and settings data, which is
// stored in LocalStorage, into the store. This is not done
// in App.vue so settings data is available in the
// route guards even on page load.
function insertLocalData(state) {
	const localSettings = getLocalstorageData(LOCALSTORAGE_FIELD_SETTINGS);
	if (localSettings !== null) {
		// Explictly set values, so new settings are merged with old localstorage settings
		if (localSettings.dark !== undefined) {
			state.settings.dark = localSettings.dark;
		}

		if (localSettings.hideBitcoinAlert !== undefined) {
			state.settings.hideBitcoinAlert = localSettings.hideBitcoinAlert;
		}
	}

	const localSession = getLocalstorageData(LOCALSTORAGE_FIELD_SESSION);
	if (localSession !== null) {
		state.session = localSession;
	}

	const localSearchHistory = getLocalstorageData(LOCALSTORAGE_FIELD_SEARCH_HISTORY);
	if (localSearchHistory !== null) {
		state.searchHistory = localSearchHistory;
	}

	return state;
}

const initialState = {
	// Ory kratos session
	session: null,
	settings: {
		// Set dark to be not initialized, so initial value can be set from media query
		dark: null,
		hideBitcoinAlert: false,
	},
	searchHistory: [],
};

export const useLocalStore = defineStore('local', {
	state: () => insertLocalData(initialState),
	getters: {
		getSession: state => state.session,
		getSettings: state => state.settings,
		getSearchHistory: state => state.searchHistory,
	},
	actions: {
		setSession(payload) {
			setLocalstorageData(LOCALSTORAGE_FIELD_SESSION, payload);
			this.session = payload;
		},
		deleteSession() {
			deleteLocalstorageData(LOCALSTORAGE_FIELD_SETTINGS);
			this.session = null;
		},
		setSettings(payload) {
			setLocalstorageData(LOCALSTORAGE_FIELD_SETTINGS, payload);
			this.settings = payload;
		},
		addSearchHistoryItem(item) {
			if (!item) {
				return;
			}

			// Remove the item if it already exist and add it to the first position
			const items = this.searchHistory.filter(i => i.title !== item.title);
			items.unshift(item);

			if (items.length > 5) {
				// Remove last element if the array is too large
				items.pop();
			}

			setLocalstorageData(LOCALSTORAGE_FIELD_SEARCH_HISTORY, items);
			this.searchHistory = items;
		},
	},
});
