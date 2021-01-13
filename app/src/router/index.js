import Vue from 'vue';
import Router from 'vue-router';
import Login from '../components/Login.vue';
import TxLookup from '../components/TxLookup.vue';
import BlockLookup from '../components/BlockLookup.vue';
import AddressLookup from '../components/AddressLookup.vue';
import NoResults from '../components/NoResults.vue';
import HeuristicEditor from '../components/HeuristicEditor.vue';
import EntryView from '../components/EntryView.vue';
import PageNotFound from '../components/PageNotFound.vue';
import * as Constants from '../constants';

Vue.use(Router);

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: Constants.ROUTE_NAME_ENTRY_PAGE,
      component: EntryView,
      meta: { title: 'Status' },
    },
    {
      path: '/block/:id',
      name: Constants.ROUTE_NAME_BLOCK_PAGE,
      component: BlockLookup,
      meta: { title: 'Block' },
    },
    {
      path: '/tx/:id',
      name: Constants.ROUTE_NAME_TRANSACTION_PAGE,
      component: TxLookup,
      meta: { title: 'Transaction' },
    },
    {
      path: '/address/:id',
      name: Constants.ROUTE_NAME_ADDRESS_PAGE,
      component: AddressLookup,
      meta: { title: 'Address' },
    },
    {
      path: '/heuristic/:id',
      name: Constants.ROUTE_NAME_HEURISTIC_PAGE,
      component: HeuristicEditor,
      meta: { title: 'Heuristic' },
    },
    {
      path: '/login',
      name: Constants.ROUTE_NAME_LOGIN_PAGE,
      component: Login,
      meta: { title: 'Login' },
    },
    {
      path: '/noresults',
      name: Constants.ROUTE_NAME_NO_RESULTS,
      component: NoResults,
      meta: { title: 'No results found' },
    },
    {
      path: '*',
      name: Constants.ROUTE_NAME_404_PAGE,
      component: PageNotFound,
      meta: { title: 'Page not found' },
    },
  ],
});
