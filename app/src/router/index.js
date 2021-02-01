import Vue from 'vue';
import Router from 'vue-router';
import Misc from '../components/user/Misc.vue';
import Settings from '../components/user/Settings.vue';
import Profile from '../components/user/Profile.vue';
import Administration from '../components/user/Administration.vue';
import Login from '../components/user/Login.vue';
import TxLookup from '../components/data/TxLookup.vue';
import BlockLookup from '../components/data/BlockLookup.vue';
import AddressLookup from '../components/data/AddressLookup.vue';
import NoResults from '../components/NoResults.vue';
import Editor from '../components/heuristic/Editor.vue';
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
      component: Editor,
      meta: { title: 'Heuristic' },
    },
    {
      path: '/login',
      name: Constants.ROUTE_NAME_LOGIN_PAGE,
      component: Login,
      meta: { title: 'Login' },
    },
    {
      path: '/settings',
      component: Settings,
      meta: { title: 'Profile' },
      children: [
        {
          // UserProfile will be rendered inside User's <router-view>
          // when /settings/profile is matched
          path: 'profile',
          name: Constants.ROUTE_NAME_USER_PROFILE_PAGE,
          component: Profile,
        },
        {
          // UserPosts will be rendered inside User's <router-view>
          // when /settings/misc is matched
          path: 'misc',
          name: Constants.ROUTE_NAME_USER_MISC_PAGE,
          component: Misc,
        },
      ],
    },
    {
      path: '/userAdministration',
      name: Constants.ROUTE_NAME_USER_ADMIN_PAGE,
      component: Administration,
      meta: { title: 'User Administration' },
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
