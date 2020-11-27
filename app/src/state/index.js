import Vue from 'vue';
import Vuex from 'vuex';
import * as Constants from '../constants'

Vue.use(Vuex);
const state = {
    msg: null,
    transaction: null,
    address: null,
    block: null,
    meta: null,
    heuristic: null,
}

function getMsg(context) {
    let msgObj = context.state.msg;
    if (msgObj === null) {
        msgObj = getResetMsgState();
    }
    return msgObj;
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
}

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
    updateMetaData(context, payload) {
        console.log("Fetching meta data");
        return doUpdate(context, Constants.ROUTE_META, 'UPDATE_META_DATA', payload);
    },
    updateHeuristicData(context, payload) {
        console.log("Fetching heuristic data");
        return doUpdate(context, Constants.ROUTE_HEURISTICS, 'UPDATE_HEURISTIC_DATA', payload);
    },
    setHeuristicData(context, payload) {
        context.commit('SET_HEURISTIC_DATA', payload);
    },
}

function doUpdate(context, route, mutation, parameter) {

    return fetch(route + parameter)
        .then(response => {
            if (!response.ok) throw new Error(response.status + " " + response.statusText)
            return response
        })
        .then(response => response.json())
        .then(data => {
            context.commit(mutation, data);
            context.dispatch('resetMsg');
        })
        .catch(e => {
            let errMsg;
            if (e.message === '500 Internal Server Error') {
                errMsg = 'Server ist not reachable';
            } else {
                errMsg = `Error getting data: ${e}`
            }

            context.dispatch('setErrorMsg', errMsg);
        });
}

const getters = {
    getErrorMsg: state => state.msg !== null && state.msg.error != null ?
        state.msg.error : null,
    getInfoMsg: state => state.msg !== null && state.msg.info != null ?
        state.msg.info : null,
    getSuccessMsg: state => state.msg !== null && state.msg.success != null ?
        state.msg.success : null,
    getWarningMsg: state => state.msg !== null && state.msg.warning != null ?
        state.msg.warning : null,
    isErrorActive: state => state.msg !== null ? state.msg.errorActive : false,
    isInfoActive: state => state.msg !== null ? state.msg.infoActive : false,
    isSuccessActive: state => state.msg !== null ? state.msg.successActive : false,
    isWarningActive: state => state.msg !== null ? state.msg.warningActive : false,
    getTransactionData: state => state.transaction,
    getAddressData: state => state.address,
    getBlockData: state => state.block,
    getMetaData: state => state.meta,
    getHeuristicData: state => state.heuristic,
}

export default new Vuex.Store({
    state,
    mutations,
    actions,
    getters
})

