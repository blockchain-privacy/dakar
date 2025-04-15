import {defineStore} from 'pinia';
import {
	deleteLocalSession,
	getLocalSession, getLocalSettings, setLocalSession, setLocalSettings,
} from '@/utilities';

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

		if (localSettings.hideBitcoinAlert !== undefined) {
			state.settings.hideBitcoinAlert = localSettings.hideBitcoinAlert;
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
		hideBitcoinAlert: false,
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
		deleteSession() {
			deleteLocalSession();
			this.session = null;
		},
		setSettings(payload) {
			setLocalSettings(payload);
			this.settings = payload;
		},
	},
});
