import Vue from 'vue';
import Vuex from 'vuex';
import {
  setLocalUser, setLocalSettings, getLocalUser, getLocalSettings,
} from '../utilities';
import {
  RESPONSE_TYPE_TRANSACTION, RESPONSE_TYPE_BLOCK, RESPONSE_TYPE_ADDRESS,
} from '../constants';

Vue.use(Vuex);

// getInitialState returns the initial state of the store
function getInitialState() {
  return {
    messages: [],
    msg: null,
    transaction: null,
    searchResultType: null,
    address: null,
    block: null,
    activeUser: null,
    settings: null,
    // failedRoute is filled with the route which the user wanted
    // to access but did for some reason (e.g. invalid credentials) fail
    failedRoute: null,
  };
}

const mutations = {
  ADD_MESSAGE(state, payload) {
    state.messages.push(payload);
  },
  RESET_MESSAGES(state) {
    state.messages = [];
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
  SET_SEARCH_RESULT_TYPE(state, payload) {
    state.searchResultType = payload;
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
        // eslint-disable-next-line no-param-reassign
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
  SET_ACTIVE_USER(state, payload) {
    state.activeUser = payload;
  },
  SET_SETTINGS(state, payload) {
    state.settings = payload;
  },
  SET_FAILED_ROUTE(state, payload) {
    state.failedRoute = payload;
  },
};

const actions = {
  addMessage(context, payload) {
    if (!payload.text || payload.text.toString().trim() === '') return;

    context.commit('ADD_MESSAGE', payload);
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
  setSearchResultType(context, payload) {
    context.commit('SET_SEARCH_RESULT_TYPE', payload);
  },
  setActiveUser(context, payload) {
    setLocalUser(payload);
    context.commit('SET_ACTIVE_USER', payload);
  },
  setSettings(context, payload) {
    setLocalSettings(payload);
    context.commit('SET_SETTINGS', payload);
  },
  setFailedRoute(context, payload) {
    context.commit('SET_FAILED_ROUTE', payload);
  },
};

const getters = {
  getMessages: (state) => state.messages,
  getTransactionData: (state) => state.transaction,
  getAddressData: (state) => state.address,
  getBlockData: (state) => state.block,
  getSearchResultType: (state) => state.searchResultType,
  getActiveUser: (state) => state.activeUser,
  getSettings: (state) => state.settings,
  getFailedRoute: (state) => state.failedRoute,
};

// insertLocalUserData inserts user data, which is
// stored in LocalStorage, into the store. This is not done
// in App.vue on purpose so user data is available in the
// route guards even on page load.
function insertLocalUserData(state) {
  const localStorageUserData = getLocalUser();
  if (localStorageUserData !== null) {
    state.activeUser = localStorageUserData;
  }
  return state;
}
// insertLocalSettingsData inserts settings data, which is
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

let state = getInitialState();
state = insertLocalUserData(state);
state = insertLocalSettingsData(state);

export default new Vuex.Store({
  state,
  mutations,
  actions,
  getters,
});
