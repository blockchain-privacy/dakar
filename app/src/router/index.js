import Vue from 'vue';
import Router from 'vue-router';
import HeuristicEditor from '../components/HeuristicEditor.vue';
import SearchView from '../components/SearchView.vue';
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
      path: '/search/:id',
      name: Constants.ROUTE_NAME_SEARCH_PAGE,
      component: SearchView,
      meta: { title: 'Search' },
    },
    {
      path: '/heuristic/:id',
      name: Constants.ROUTE_NAME_HEURISTIC_PAGE,
      component: HeuristicEditor,
      meta: { title: 'Heuristic' },
    },
    {
      path: '*',
      name: Constants.ROUTE_NAME_404_PAGE,
      component: PageNotFound,
      meta: { title: 'Page not found' },
    },
  ],
});
