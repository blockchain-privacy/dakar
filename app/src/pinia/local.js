// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {defineStore} from 'pinia';
import {
	LOCALSTORAGE_FIELD_SEARCH_HISTORY,
	LOCALSTORAGE_FIELD_SESSION,
	LOCALSTORAGE_FIELD_SETTINGS,
	LOCALSTORAGE_FIELD_UPDATE_ID,
	UPDATE_ID,
} from '@/constants/index.js';

function setLocalstorageData(key, settingsData) {
	localStorage.setItem(key, JSON.stringify(settingsData));
}

function getLocalstorageData(key) {
	let localStorageData = localStorage.getItem(key);
	if (localStorageData !== null) {
		localStorageData = JSON.parse(localStorageData);
	}

	return localStorageData;
}

function deleteLocalstorageData(key) {
	localStorage.removeItem(key);
}

// InsertLocalData inserts session and settings data, which is
// stored in LocalStorage, into the store. This is not done
// in App.vue so settings data is available in the
// route guards even on page load.
function insertLocalData(state) {
	const localSettings = getLocalstorageData(LOCALSTORAGE_FIELD_SETTINGS);
	if (localSettings !== null && localSettings.theme !== undefined) {
		// Explicitly set values, so new settings are merged with old localstorage settings
		state.settings.theme = localSettings.theme;
	}

	const localSession = getLocalstorageData(LOCALSTORAGE_FIELD_SESSION);
	if (localSession !== null) {
		state.session = localSession;
	}

	const localSearchHistory = getLocalstorageData(LOCALSTORAGE_FIELD_SEARCH_HISTORY);
	if (localSearchHistory !== null) {
		state.searchHistory = localSearchHistory;
	}

	const updateID = getLocalstorageData(LOCALSTORAGE_FIELD_UPDATE_ID);
	if (updateID !== null) {
		state.lastUpdateID = updateID;
	}

	return state;
}

const initialState = {
	// Ory kratos session
	session: null,
	settings: {
		theme: 'system',
	},
	searchHistory: [],
	lastUpdateID: 0,
};

export const useLocalStore = defineStore('local', {
	state: () => insertLocalData(initialState),
	getters: {
		getSession: state => state.session,
		getSettings: state => state.settings,
		getSearchHistory: state => state.searchHistory,
		getLastUpdateID: state => state.lastUpdateID,
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
		setLastUpdateID(payload) {
			setLocalstorageData(LOCALSTORAGE_FIELD_UPDATE_ID, payload);
			this.lastUpdateID = payload;
		},
		wasUpdateViewed() {
			return this.lastUpdateID === UPDATE_ID;
		},
		addSearchHistoryItem(item) {
			if (!item) {
				return;
			}

			// Remove the item if it already exists and add it to the first position
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
