import Vue from 'vue';
import Vuex from 'vuex';
import {
  doPost, doGet, handleError, setLocalUser, setLocalSettings,
} from '../utilities';
import * as Constants from '../constants';

import Router from '../router';

Vue.use(Vuex);

function handleGet(context, route, mutation, parameter) {
  return doGet(route, Router, parameter).then((data) => {
    context.commit(mutation, data);
    context.dispatch('resetMessages');
  }).catch((e) => {
    handleError(context, e);
  });
}

function handlePost(context, route, mutation, parameter, body) {
  return doPost(route, Router, body, parameter)
    .then((data) => {
      context.commit(mutation, data);
      context.dispatch('resetMessages');
    })
    .catch((e) => {
      handleError(context, e);
    });
}

// getInitialState returns the initial state of the store
function getInitialState() {
  return {
    messages: [],
    msg: null,
    transaction: null,
    searchResultType: null,
    address: null,
    block: null,
    meta: null,
    heuristic: null,
    heuristicDetails: new Map(),
    userList: null,
    activeUser: null,
    settings: null,
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
  UPDATE_META_DATA(state, payload) {
    state.meta = payload;
  },
  UPDATE_HEURISTIC_DATA(state, payload) {
    state.heuristic = payload;
  },
  SET_HEURISTIC_DATA(state, payload) {
    mutations.UPDATE_HEURISTIC_DATA(state, payload);
  },
  SET_HEURISTIC_DETAILS(state, payload) {
    state.heuristicDetails = payload;
  },
  ADD_HEURISTIC_DETAILS(state, payload) {
    state.heuristicDetails.set(payload.uid, payload);
  },
  SET_SEARCH_RESULT_TYPE(state, payload) {
    state.searchResultType = payload;
  },
  UPDATE_SEARCH_RESULT(state, payload) {
    state.searchResultType = payload.type;
    switch (payload.type) {
      case Constants.RESPONSE_TYPE_TRANSACTION:
        state.transaction = payload.payload;
        break;
      case Constants.RESPONSE_TYPE_BLOCK:
        state.block = payload.payload;
        break;
      case Constants.RESPONSE_TYPE_ADDRESS:
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
  UPDATE_USER_LIST(state, payload) {
    state.userList = payload;
  },
  SET_ACTIVE_USER(state, payload) {
    state.activeUser = payload;
  },
  SET_SETTINGS(state, payload) {
    state.settings = payload;
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
    return handleGet(context, Constants.ROUTE_BLOCK, 'UPDATE_BLOCK_DATA', payload);
  },
  updateTransactionData(context, payload) {
    return handleGet(context, Constants.ROUTE_TRANSACTION, 'UPDATE_TRANSACTION_DATA', payload);
  },
  updateAddressData(context, payload) {
    return handleGet(context, Constants.ROUTE_ADDRESS, 'UPDATE_ADDRESS_DATA', payload);
  },
  updateSearchResult(context, payload) {
    return handleGet(context, Constants.ROUTE_SEARCH, 'UPDATE_SEARCH_RESULT', payload);
  },
  updateMetaData(context, payload) {
    return handleGet(context, Constants.ROUTE_META, 'UPDATE_META_DATA', payload);
  },
  updateHeuristicData(context, payload) {
    return handleGet(context, Constants.ROUTE_HEURISTICS, 'UPDATE_HEURISTIC_DATA', payload);
  },
  updateHeuristicDetails(context, payload) {
    return handlePost(context, Constants.ROUTE_HEURISTIC_DETAILS, 'ADD_HEURISTIC_DETAILS',
      payload.parameter, payload.body);
  },
  updateUserList(context) {
    return handleGet(context, Constants.ROUTE_USER_LIST, 'UPDATE_USER_LIST');
  },
  resetHeuristicDetails(context) {
    context.commit('SET_HEURISTIC_DETAILS', new Map());
  },
  setHeuristicDetails(context, payload) {
    context.commit('SET_HEURISTIC_DETAILS', payload);
  },
  setHeuristicData(context, payload) {
    context.commit('SET_HEURISTIC_DATA', payload);
  },
  setSearchResultType(context, payload) {
    context.commit('SET_SEARCH_RESULT_TYPE', payload);
  },
  setUserList(context, payload) {
    context.commit('UPDATE_USER_LIST', payload);
  },
  setActiveUser(context, payload) {
    setLocalUser(payload);
    context.commit('SET_ACTIVE_USER', payload);
  },
  setSettings(context, payload) {
    setLocalSettings(payload);
    context.commit('SET_SETTINGS', payload);
  },
};

const getters = {
  getMessages: (state) => state.messages,
  getTransactionData: (state) => state.transaction,
  getAddressData: (state) => state.address,
  getBlockData: (state) => state.block,
  getMetaData: (state) => state.meta,
  getHeuristicData: (state) => state.heuristic,
  getHeuristicDetails: (state) => state.heuristicDetails,
  getSearchResultType: (state) => state.searchResultType,
  getUserList: (state) => state.userList,
  getActiveUser: (state) => state.activeUser,
  getSettings: (state) => state.settings,
};

const state = getInitialState();

export default new Vuex.Store({
  state,
  mutations,
  actions,
  getters,
});
