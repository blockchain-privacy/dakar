import Vue from 'vue';
import Vuex from 'vuex';
import {
  doPost, doGet, handleError, isInvalidTokenMsg,
} from '../utilities';
import * as Constants from '../constants';

import Router from '../router';

Vue.use(Vuex);

function handleGet(context, route, mutation, parameter) {
  return doGet(route, parameter).then((data) => {
    if (isInvalidTokenMsg(data, Router)) return;
    context.commit(mutation, data);
    context.dispatch('resetMsg');
  }).catch((e) => {
    handleError(context, e);
  });
}

function handlePost(context, route, mutation, parameter, body) {
  return doPost(route, parameter, body)
    .then((data) => {
      if (isInvalidTokenMsg(data, Router)) return;
      context.commit(mutation, data);
      context.dispatch('resetMsg');
    })
    .catch((e) => {
      handleError(context, e);
    });
}

function getResetMsgState() {
  return {
    error: null,
    errorActive: false,
    info: null,
    infoActive: false,
    success: null,
    successActive: false,
    warning: null,
    warningActive: false,
  };
}

function getMsg(context) {
  let msgObj = context.state.msg;
  if (msgObj === null) {
    msgObj = getResetMsgState();
  }
  return msgObj;
}

// getInitialState returns the initial state of the store
function getInitialState() {
  return {
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
  };
}

const mutations = {
  SET_MSG(state, payload) {
    state.msg = payload;
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
};

const actions = {
  resetMsg(context) {
    context.commit('SET_MSG', getResetMsgState());
  },
  setErrorMsg(context, payload) {
    const msgObj = getMsg(context);
    msgObj.error = payload;
    msgObj.errorActive = true;
    context.commit('SET_MSG', msgObj);
  },
  setInfoMsg(context, payload) {
    const msgObj = getMsg(context);
    msgObj.info = payload;
    msgObj.infoActive = true;
    context.commit('SET_MSG', msgObj);
  },
  setSuccessMsg(context, payload) {
    const msgObj = getMsg(context);
    msgObj.success = payload;
    msgObj.successActive = true;
    context.commit('SET_MSG', msgObj);
  },
  setWarningMsg(context, payload) {
    const msgObj = getMsg(context);
    msgObj.warning = payload;
    msgObj.warningActive = true;
    context.commit('SET_MSG', msgObj);
  },
  setErrorActive(context, payload) {
    const msgObj = getMsg(context);
    msgObj.errorActive = payload;
    context.commit('SET_MSG', msgObj);
  },
  setInfoActive(context, payload) {
    const msgObj = getMsg(context);
    msgObj.infoActive = payload;
    context.commit('SET_MSG', msgObj);
  },
  setSuccessActive(context, payload) {
    const msgObj = getMsg(context);
    msgObj.successActive = payload;
    context.commit('SET_MSG', msgObj);
  },
  setWarningActive(context, payload) {
    const msgObj = getMsg(context);
    msgObj.warningActive = payload;
    context.commit('SET_MSG', msgObj);
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
    context.commit('SET_ACTIVE_USER', payload);
  },
};

const getters = {
  getErrorMsg: (state) => (state.msg !== null && state.msg.error != null
    ? state.msg.error : null),
  getInfoMsg: (state) => (state.msg !== null && state.msg.info != null
    ? state.msg.info : null),
  getSuccessMsg: (state) => (state.msg !== null && state.msg.success != null
    ? state.msg.success : null),
  getWarningMsg: (state) => (state.msg !== null && state.msg.warning != null
    ? state.msg.warning : null),
  isErrorActive: (state) => (state.msg !== null ? state.msg.errorActive : false),
  isInfoActive: (state) => (state.msg !== null ? state.msg.infoActive : false),
  isSuccessActive: (state) => (state.msg !== null ? state.msg.successActive : false),
  isWarningActive: (state) => (state.msg !== null ? state.msg.warningActive : false),
  getTransactionData: (state) => state.transaction,
  getAddressData: (state) => state.address,
  getBlockData: (state) => state.block,
  getMetaData: (state) => state.meta,
  getHeuristicData: (state) => state.heuristic,
  getHeuristicDetails: (state) => state.heuristicDetails,
  getSearchResultType: (state) => state.searchResultType,
  getUserList: (state) => state.userList,
  getActiveUser: (state) => state.activeUser,
};

const state = getInitialState();

export default new Vuex.Store({
  state,
  mutations,
  actions,
  getters,
});
