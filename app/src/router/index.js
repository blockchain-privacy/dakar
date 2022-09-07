import Vue from 'vue';
import Router from 'vue-router';
import {
  isAdminIdentity, isPrivilegedIdentity, isSessionExpired,
} from '../utilities';
import EntryView from '../components/EntryView.vue';
import ConnectionLookup from '../components/tools/ConnectionLookup.vue';
import Settings from '../components/user/Settings.vue';
import Profile from '../components/user/Profile.vue';
import Administration from '../components/user/Administration.vue';
import Login from '../components/user/Login.vue';
import TxLookup from '../components/data/TxLookup.vue';
import BlockLookup from '../components/data/BlockLookup.vue';
import AddressLookup from '../components/data/address/AddressLookup.vue';
import NoResults from '../components/NoResults.vue';
import Editor from '../components/heuristic/Editor.vue';
import StatusView from '../components/StatusView.vue';
import PageNotFound from '../components/PageNotFound.vue';
import Tools from '../components/tools/Tools.vue';
import ShortestPath from '../components/tools/ShortestPath.vue';
import Heuristics from '../components/tools/Heuristics.vue';
import * as Constants from '../constants';
import Store from '../state';
import HMIView from '../components/cluster/HMIView.vue';
import ClusterOverview from '../components/tools/ClusterOverview.vue';
import Attributions from '../components/tools/attributions/Attributions.vue';
import AddressExclusions from '../components/tools/AddressExclusions.vue';
import Recovery from '../components/user/Recovery.vue';
import Wiki from '../components/wiki/Wiki.vue';

Vue.use(Router);

function getSessionData() {
  return Store.getters.getSession;
}

function isPrivileged() {
  const sessionData = getSessionData();

  return isPrivilegedIdentity(sessionData) || isAdminIdentity(sessionData);
}

function isAdmin() {
  return isAdminIdentity(getSessionData());
}

function checkSession(to, next, fn) {
  const sessionData = getSessionData();

  if (!sessionData) {
    Store.dispatch('setFailedRoute', to);
    next({ name: Constants.ROUTE_NAME_LOGIN_PAGE });
    return;
  }

  // check if token timeout has been reached
  if (isSessionExpired(sessionData)) {
    Store.dispatch('setFailedRoute', to);
    Store.dispatch('setSession', null);
    Store.dispatch('addMessage', { type: 'info', text: 'Your session timed out', temporary: true });
    next({ name: Constants.ROUTE_NAME_LOGIN_PAGE });
    return;
  }

  if ((fn) ? !fn() : false) {
    next({ name: Constants.ROUTE_NAME_ENTRY_PAGE });
    return;
  }

  next();
}

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: Constants.ROUTE_NAME_ENTRY_PAGE,
      component: EntryView,
      meta: { title: 'Entry' },
    },
    {
      path: '/status',
      name: Constants.ROUTE_NAME_STATUS_PAGE,
      component: StatusView,
      meta: { title: 'Status' },
      beforeEnter: (to, from, next) => {
        checkSession(to, next, isPrivileged);
      },
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
      component: Editor,
      meta: { title: 'Heuristic' },
      beforeEnter: (to, from, next) => {
        checkSession(to, next, isPrivileged);
      },
    },
    {
      path: '/hmiLookup/:id',
      name: Constants.ROUTE_NAME_CLUSTER_VIEW_PAGE,
      component: HMIView,
      meta: { title: 'Cluster View' },
      beforeEnter: (to, from, next) => {
        checkSession(to, next, isPrivileged);
      },
    },
    {
      path: '/login',
      name: Constants.ROUTE_NAME_LOGIN_PAGE,
      component: Login,
      meta: { title: 'Login' },
    },
    {
      path: '/wiki/:file',
      name: Constants.ROUTE_NAME_WIKI,
      component: Wiki,
      meta: { title: 'Wiki' },
      beforeEnter: (to, from, next) => {
        checkSession(to, next, isPrivileged);
      },
    },
    {
      path: '/recovery',
      name: Constants.ROUTE_NAME_ACCOUNT_RECOVERY,
      component: Recovery,
      meta: { title: 'Password Reset' },
    },
    {
      path: '/settings',
      component: Settings,
      meta: { title: 'Settings' },
      children: [
        {
          path: 'profile',
          name: Constants.ROUTE_NAME_USER_PROFILE_PAGE,
          component: Profile,
        },
      ],
    },
    {
      path: '/tools',
      component: Tools,
      meta: { title: 'Tools' },
      beforeEnter: (to, from, next) => {
        checkSession(to, next, isPrivileged);
      },
      children: [
        {
          path: 'shortestPath',
          name: Constants.ROUTE_NAME_SHORTEST_PATH_PAGE,
          component: ShortestPath,
        },
        {
          path: 'heuristics',
          name: Constants.ROUTE_NAME_USER_HEURISTIC_PAGE,
          component: Heuristics,
        },
        {
          path: 'connectionLookup',
          name: Constants.ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
          component: ConnectionLookup,
        },
        {
          path: 'clusterOverview',
          name: Constants.ROUTE_NAME_CLUSTER_OVERVIEW,
          component: ClusterOverview,
        },
        {
          path: 'attributions',
          name: Constants.ROUTE_NAME_ATTRIBUTIONS,
          component: Attributions,
        },
        {
          path: 'addressExclusions',
          name: Constants.ROUTE_NAME_ADDRESS_EXCLUSIONS,
          component: AddressExclusions,
        },
      ],
    },
    {
      path: '/userAdministration',
      name: Constants.ROUTE_NAME_USER_ADMIN_PAGE,
      component: Administration,
      meta: { title: 'User Administration' },
      beforeEnter: (to, from, next) => {
        checkSession(to, next, isAdmin);
      },
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
