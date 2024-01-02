import {createRouter, createWebHistory} from 'vue-router';
import {isAdminIdentity, isPrivilegedIdentity} from '@/utilities';
import EntryPage from '../components/EntryPage.vue';
import ConnectionLookupPage from '../components/tools/ConnectionLookupPage.vue';
import SettingsPage from '../components/user/SettingsPage.vue';
import ProfilePage from '../components/user/ProfilePage.vue';
import AdministrationPage from '../components/user/AdministrationPage.vue';
import LoginPage from '../components/user/LoginPage.vue';
import TransactionPage from '../components/explorer/transaction/TransactionPage.vue';
import BlockPage from '../components/explorer/BlockPage.vue';
import AddressPage from '../components/explorer/address/AddressPage.vue';
import WorkspaceEditorPage from '../components/workspace/WorkspaceEditorPage.vue';
import StatusPage from '../components/StatusPage.vue';
import ToolsPage from '../components/tools/ToolsPage.vue';
import ShortestPathPage from '../components/tools/ShortestPathPage.vue';
import WorkspacePage from '@/components/tools/WorkspacePage.vue';
import * as Constants from '../constants';
import ClusterPage from '../components/tools/clusters/ClusterPage.vue';
import AttributionsPage from '../components/tools/attributions/AttributionsPage.vue';
import AddressExclusionsPage from '../components/tools/addressExclusions/AddressExclusionsPage.vue';
import RecoveryPage from '../components/user/RecoveryPage.vue';
import WikiPage from '../components/wiki/WikiPage.vue';
import TextLoaderPage from '../components/TextLoaderPage.vue';
import ErrorPage from '@/components/ErrorPage.vue';
import {useLocalStore} from '@/pinia/local';
import {useNavStore} from '@/pinia/nav';
import {useMsgStore} from '@/pinia/msg';

let msgStore = null;
let navStore = null;
let localStore = null;

// Call this right after the pinia store was created and added to the vue instance
export function setupStore() {
	msgStore = useMsgStore();
	navStore = useNavStore();
	localStore = useLocalStore();
}

function isPrivileged() {
	return isPrivilegedIdentity(localStore.getSession) || isAdminIdentity(localStore.getSession);
}

function isAdmin() {
	return isAdminIdentity(localStore.getSession);
}

function checkSession(to, next, fn) {
	if (!localStore.getSession) {
		navStore.setFailedRoute(to);
		next({name: Constants.ROUTE_NAME_LOGIN_PAGE});
		return;
	}

	if ((fn) ? !fn() : false) {
		next({name: Constants.ROUTE_NAME_ENTRY_PAGE});
		return;
	}

	next();
}

export const router = createRouter({
	history: createWebHistory(),
	routes: [
		{
			path: '/',
			name: Constants.ROUTE_NAME_ENTRY_PAGE,
			component: EntryPage,
			meta: {title: 'Entry'},
		},
		{
			path: '/status',
			name: Constants.ROUTE_NAME_STATUS_PAGE,
			component: StatusPage,
			meta: {title: 'Status'},
			async beforeEnter(to, from, next) {
				checkSession(to, next, isPrivileged);
			},
		},
		{
			path: '/block/:id',
			name: Constants.ROUTE_NAME_BLOCK_PAGE,
			component: BlockPage,
			meta: {title: 'Block'},
		},
		{
			path: '/tx/:id',
			name: Constants.ROUTE_NAME_TRANSACTION_PAGE,
			component: TransactionPage,
			meta: {title: 'Transaction'},
		},
		{
			path: '/address/:id',
			name: Constants.ROUTE_NAME_ADDRESS_PAGE,
			component: AddressPage,
			meta: {title: 'Address'},
		},
		{
			path: '/workspace/:id',
			name: Constants.ROUTE_NAME_WORKSPACE_PAGE,
			component: WorkspaceEditorPage,
			meta: {title: 'Workspace'},
			async beforeEnter(to, from, next) {
				checkSession(to, next, isPrivileged);
			},
		},
		{
			path: '/login',
			name: Constants.ROUTE_NAME_LOGIN_PAGE,
			component: LoginPage,
			meta: {title: 'Login'},
		},
		{
			// Wiki root page
			path: '/wiki',
			name: Constants.ROUTE_NAME_WIKI_ROOT,
			component: WikiPage,
			meta: {title: 'Wiki'},
			async beforeEnter(to, from, next) {
				checkSession(to, next, isPrivileged);
			},
		},
		{
			// Wiki content page
			// allow additional slashes in path
			path: '/wiki/:file(.*)',
			name: Constants.ROUTE_NAME_WIKI,
			component: WikiPage,
			meta: {title: 'Wiki'},
			async beforeEnter(to, from, next) {
				checkSession(to, next, isPrivileged);
			},
		},
		{
			path: '/recovery',
			name: Constants.ROUTE_NAME_ACCOUNT_RECOVERY,
			component: RecoveryPage,
			meta: {title: 'Account Recovery'},
		},
		{
			path: '/settings',
			component: SettingsPage,
			meta: {title: 'Settings'},
			children: [
				{
					path: 'profile/:tabName?',
					name: Constants.ROUTE_NAME_USER_PROFILE_PAGE,
					component: ProfilePage,
				},
			],
		},
		{
			path: '/tools',
			component: ToolsPage,
			meta: {title: 'Tools'},
			async beforeEnter(to, from, next) {
				checkSession(to, next, isPrivileged);
			},
			children: [
				{
					path: 'shortestPath',
					name: Constants.ROUTE_NAME_SHORTEST_PATH_PAGE,
					component: ShortestPathPage,
				},
				{
					path: 'workspaces',
					name: Constants.ROUTE_NAME_WORKSPACES_PAGE,
					component: WorkspacePage,
				},
				{
					path: 'connectionLookup',
					name: Constants.ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
					component: ConnectionLookupPage,
				},
				{
					path: 'clusterOverview',
					name: Constants.ROUTE_NAME_CLUSTER_OVERVIEW,
					component: ClusterPage,
				},
				{
					path: 'attributions',
					name: Constants.ROUTE_NAME_ATTRIBUTIONS,
					component: AttributionsPage,
				},
				{
					path: 'addressExclusions',
					name: Constants.ROUTE_NAME_ADDRESS_EXCLUSIONS,
					component: AddressExclusionsPage,
				},
			],
		},
		{
			path: '/userAdministration',
			name: Constants.ROUTE_NAME_USER_ADMIN_PAGE,
			component: AdministrationPage,
			meta: {title: 'User Administration'},
			async beforeEnter(to, from, next) {
				checkSession(to, next, isAdmin);
			},
		},
		{
			path: '/about',
			name: Constants.ROUTE_NAME_ABOUT,
			component: TextLoaderPage,
			props: {pageTitle: 'About', url: 'about.html'},
			meta: {title: 'About'},
		},
		{
			path: '/privacy',
			name: Constants.ROUTE_NAME_PRIVACY,
			component: TextLoaderPage,
			props: {pageTitle: 'Privacy Policy', url: 'privacy_policy.html'},
			meta: {title: 'Privacy Policy'},
		},
		{
			path: '/termsOfUse',
			name: Constants.ROUTE_NAME_TERMS_OF_USE,
			component: TextLoaderPage,
			props: {pageTitle: 'Terms of Use', url: 'terms_of_use.html'},
			meta: {title: 'Terms of Use'},
		},
		{
			path: '/noResults',
			name: Constants.ROUTE_NAME_NO_RESULTS,
			component: ErrorPage,
			meta: {title: 'No results found!'},
			props: {
				default: true,
				title: 'No results found!',
				description: 'Your search query did not return any results. Either navigate back or click below to get back to the entry page.',
				imageSource: '/src/assets/no_results.webp',
			},
		},
		{
			path: '/error',
			name: Constants.ROUTE_NAME_ERROR,
			component: ErrorPage,
			meta: {title: 'Error'},
			props: {
				default: true,
				title: 'Error',
				description: '',
				imageSource: '/src/assets/bugs.webp',
			},
		},
		{
			path: '/:catchAll(.*)',
			name: Constants.ROUTE_NAME_404_PAGE,
			component: ErrorPage,
			meta: {title: '404 - Page not found!'},
			props: {
				default: true,
				title: '404 - Page not found!',
				description: 'The requested page does not exist. Either navigate back or click below to get back to the entry page.',
				imageSource: '/src/assets/bugs.webp',
			},
		},
	],
});

router.beforeEach((to, from) => {
	if (from && to.name !== from.name) {
		// Clear all notifications belonging to the previous page
		msgStore.filterMessages(from.name);
	}

	return true;
});

