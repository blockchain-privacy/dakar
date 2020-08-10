import Vue from 'vue'
import Router from 'vue-router'
import SearchView from '../components/SearchView';
import EntryView from '../components/EntryView';
import PageNotFound from '../components/PageNotFound';
import * as Constants from '../constants';

Vue.use(Router)

export default new Router({
    mode: 'history',
    routes: [
        {
            path: '/',
            name: Constants.ROUTE_NAME_ENTRY_PAGE,
            component: EntryView
        },
        {
            path: '/search/:id',
            name: Constants.ROUTE_NAME_SEARCH_PAGE,
            component: SearchView
        },
        {
            path: '*',
            name: Constants.ROUTE_NAME_404_PAGE,
            component: PageNotFound
        },

    ]
})
