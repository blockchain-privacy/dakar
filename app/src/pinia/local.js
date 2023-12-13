import {defineStore} from 'pinia';
import {getLocalSession, getLocalSettings, setLocalSession, setLocalSettings} from '@/utilities';

// InsertLocalSettingsData inserts settings data, which is
// stored in LocalStorage, into the store. This is not done
// in App.vue on purpose so settings data is available in the
// route guards even on page load.
function insertLocalSettingsData(state) {
	const localStorageSettingsData = getLocalSettings();
	if (localStorageSettingsData !== null) {
		state.settings = localStorageSettingsData;
	}

	return state;
}

// InsertLocalSessionData inserts session data, which is
// stored in LocalStorage, into the store. This is not done
// in App.vue on purpose so settings data is available in the
// route guards even on page load.
function insertLocalSessionData(state) {
	const localStorageSessionData = getLocalSession();
	if (localStorageSessionData !== null) {
		state.session = localStorageSessionData;
	}

	return state;
}

let initialState = {
	// Ory kratos session
	session: null,
	settings: null,
};
initialState = insertLocalSettingsData(initialState);
initialState = insertLocalSessionData(initialState);

export const useLocalStore = defineStore('local', {
	state: () => initialState,
	getters: {
		getSession: state => state.session,
		getSettings: state => state.settings,
	},
	actions: {
		setSession(payload) {
			setLocalSession(payload);
			this.session = payload;
		},
		setSettings(payload) {
			setLocalSettings(payload);
			this.settings = payload;
		},
	},
});
