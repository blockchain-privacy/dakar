import {defineStore} from 'pinia';
import {
	getLocalSession, getLocalSettings, setLocalSession, setLocalSettings,
} from '@/utilities';
import {BLOCKCHAIN_DASH} from '@/constants/index.js';

// InsertLocalData inserts session and settings data, which is
// stored in LocalStorage, into the store. This is not done
// in App.vue so settings data is available in the
// route guards even on page load.
function insertLocalData(state) {
	const localSettings = getLocalSettings();
	if (localSettings !== null) {
		// Explictly set values, so new settings are merged with old localstorage settings
		if (localSettings.dark !== undefined) {
			state.settings.dark = localSettings.dark;
		}

		if (localSettings.blockchainMode !== undefined) {
			state.settings.blockchainMode = localSettings.blockchainMode;
		}
	}

	const localSession = getLocalSession();
	if (localSession !== null) {
		state.session = localSession;
	}

	return state;
}

const initialState = {
	// Ory kratos session
	session: null,
	settings: {
		// Set dark to be not initialized, so initial value can be set from media query
		dark: null,
		blockchainMode: BLOCKCHAIN_DASH,
	},
};

export const useLocalStore = defineStore('local', {
	state: () => insertLocalData(initialState),
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
