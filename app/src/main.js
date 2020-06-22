import Vue from 'vue'
import App from './App.vue'
import vuetify from './plugins/vuetify';
import Vuex from 'vuex';
import router from './router'

Vue.config.productionTip = false

// vuex store
Vue.use(Vuex);
const state = {
    msg: null,
    transaction: null,
    address: null
}

function getMsg(context) {
    let msgObj = context.state.msg;
    if (msgObj === null) {
        msgObj = {};
    }
    return msgObj;
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
    }
}

const actions = {
    resetMsg(context) {
        context.commit('SET_MSG', null);
    },
    setErrorMsg(context, payload) {
        const msgObj = getMsg(context);
        msgObj.error = payload;
        context.commit('SET_MSG', msgObj);
    },
    setInfoMsg(context, payload) {
        const msgObj = getMsg(context);
        msgObj.info = payload;
        context.commit('SET_MSG', msgObj);
    },
    setSuccessMsg(context, payload) {
        const msgObj = getMsg(context);
        msgObj.success = payload;
        context.commit('SET_MSG', msgObj);
    },
    setWarningMsg(context, payload) {
        const msgObj = getMsg(context);
        msgObj.warning = payload;
        context.commit('SET_MSG', msgObj);
    },
    setTransactionData(context, payload) {
        context.commit('SET_TRANSACTION_DATA', payload);
    },
    setAddressData(context, payload) {
        context.commit('SET_ADDRESS_DATA', payload);
    }
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
    getTransactionData: state => state.transaction,
    getAddressData: state => state.address,
}

const store = new Vuex.Store({
    state,
    mutations,
    actions,
    getters
})

new Vue({
    vuetify,
    store,
    router,
    render: h => h(App)
}).$mount('#app')
