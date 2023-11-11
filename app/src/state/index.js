import {createStore} from 'vuex';
import {
	setLocalSettings, getLocalSettings, getLocalSession, setLocalSession,
} from '@/utilities';
import {
	RESPONSE_TYPE_TRANSACTION, RESPONSE_TYPE_BLOCK, RESPONSE_TYPE_ADDRESS,
} from '@/constants';

let msgCounter = 1;
// GetInitialState returns the initial state of the store
function getInitialState() {
	return {
		messages: new Map(),
		transaction: null,
		searchResultType: null,
		address: null,
		block: null,
		// Ory kratos session
		session: null,
		settings: null,
		// FailedRoute is filled with the route which the user wanted
		// to access but did for some reason (e.g. invalid credentials) fail
		failedRoute: null,
		// PushFromUserInput is true if a data route
		// navigated to by using router.push() instead of browser navigation
		pushFromUserInput: false,
	};
}

const mutations = {
	ADD_MESSAGE(state, payload) {
		state.messages.set(msgCounter, payload);
		msgCounter += 1;
	},
	RESET_MESSAGES(state) {
		state.messages = new Map();
	},
	REMOVE_MESSAGE(state, msgKey) {
		state.messages.delete(msgKey);
	},
	FILTER_MESSAGES(state, category) {
		for (const [key, value] of state.messages) {
			if (value.category === category) {
				state.messages.delete(key);
			}
		}
	},
	SET_TRANSACTION_DATA(state, payload) {
		state.transaction = payload;
	},
	SET_ADDRESS_DATA(state, payload) {
		state.address = payload;
	},
	SET_BLOCK_DATA(state, payload) {
		state.block = payload;
	},
	UPDATE_SEARCH_RESULT(state, payload) {
		state.searchResultType = payload.type;
		switch (payload.type) {
			case RESPONSE_TYPE_TRANSACTION:
				state.transaction = payload.payload;
				break;
			case RESPONSE_TYPE_BLOCK:
				state.block = payload.payload;
				break;
			case RESPONSE_TYPE_ADDRESS:
				state.address = payload.payload;
				break;
			default:

				state = getInitialState();
		}
	},
	UPDATE_BLOCK_DATA(state, payload) {
		state.searchResultType = payload.type;
		state.block = payload.payload;
	},
	UPDATE_TRANSACTION_DATA(state, payload) {
		state.searchResultType = payload.type;
		state.transaction = payload.payload;
	},
	UPDATE_ADDRESS_DATA(state, payload) {
		state.searchResultType = payload.type;
		state.address = payload.payload;
	},
	// Ory kratos session
	SET_SESSION(state, payload) {
		state.session = payload;
	},
	SET_SETTINGS(state, payload) {
		state.settings = payload;
	},
	SET_FAILED_ROUTE(state, payload) {
		state.failedRoute = payload;
	},
	SET_PUSH_FROM_USER_INPUT(state, payload) {
		state.pushFromUserInput = payload;
	},
};

const actions = {
	addMessage(context, payload) {
		if (!payload.text || payload.text.toString().trim() === '') {
			return;
		}

		// Convert potential error object to string
		payload.text = payload.text.toString();

		context.commit('ADD_MESSAGE', payload);
	},
	removeMessage(context, payload) {
		context.commit('REMOVE_MESSAGE', payload);
	},
	filterMessages(context, payload) {
		context.commit('FILTER_MESSAGES', payload);
	},
	resetMessages(context) {
		context.commit('RESET_MESSAGES');
	},
	setTransactionData(context, payload) {
		context.commit('SET_TRANSACTION_DATA', payload);
	},
	setAddressData(context, payload) {
		context.commit('SET_ADDRESS_DATA', payload);
	},
	setBlockData(context, payload) {
		context.commit('SET_BLOCK_DATA', payload);
	},
	updateBlockData(context, payload) {
		context.commit('UPDATE_BLOCK_DATA', payload);
	},
	updateTransactionData(context, payload) {
		context.commit('UPDATE_TRANSACTION_DATA', payload);
	},
	updateAddressData(context, payload) {
		context.commit('UPDATE_ADDRESS_DATA', payload);
	},
	updateSearchResult(context, payload) {
		context.commit('UPDATE_SEARCH_RESULT', payload);
	},
	setSession(context, payload) {
		setLocalSession(payload);
		context.commit('SET_SESSION', payload);
	},
	setSettings(context, payload) {
		setLocalSettings(payload);
		context.commit('SET_SETTINGS', payload);
	},
	setFailedRoute(context, payload) {
		context.commit('SET_FAILED_ROUTE', payload);
	},
	setPushFromUserInput(context, payload) {
		context.commit('SET_PUSH_FROM_USER_INPUT', payload);
	},
};

const getters = {
	getMessages: state => state.messages,
	getTransactionData: state => state.transaction,
	getAddressData: state => state.address,
	getBlockData: state => state.block,
	getSearchResultType: state => state.searchResultType,
	getSession: state => state.session,
	getSettings: state => state.settings,
	getFailedRoute: state => state.failedRoute,
	getPushFromUserInput: state => state.pushFromUserInput,
};

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

let state = getInitialState();
state = insertLocalSettingsData(state);
state = insertLocalSessionData(state);

const store = createStore({
	state() {
		return state;
	},
	mutations,
	actions,
	getters,
});

export default store;
