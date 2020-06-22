import Vue from 'vue'
import Router from 'vue-router'
import SearchView from "../components/SearchView";
import EntryView from "../components/EntryView";

Vue.use(Router)
const PageNotFound = { template: '<div>404 Page not found</div>' }
export default new Router({
    mode: 'history',
    routes: [
        {
            path: '/',
            name: 'Entry Page',
            component: EntryView
        },
        {
            path: '/search/:id',
            name: 'Search Page',
            component: SearchView
        },
        {
            path: '*',
            name: 'Page not found',
            component: PageNotFound
        },

    ]
})
