// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {createRouter, createWebHistory} from 'vue-router';
import {isAnyAdminIdentity, isAnyPrivilegedIdentity} from '@/utilities';
import EntryPage from '../components/EntryPage.vue';
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
import WorkspacePage from '@/components/tools/workspaces/WorkspacePage.vue';
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
import NoResultsImg from '@/assets/no_results.webp';
import BugsImg from '@/assets/bugs.webp';

let msgStore = null;
let navStore = null;
let localStore = null;

// Call this right after the pinia store was created and added to the vue instance
export function setupStore() {
	msgStore = useMsgStore();
	navStore = useNavStore();
	localStore = useLocalStore();
}

function isAdmin() {
	return isAnyAdminIdentity(localStore.getSession);
}

function isPrivileged() {
	return isAnyPrivilegedIdentity(localStore.getSession)
		|| isAdmin();
}

function checkSession(to, fn) {
	if (!localStore.getSession) {
		navStore.setFailedRoute(to);
		return {name: Constants.ROUTE_NAME_LOGIN_PAGE};
	}

	if ((fn) ? !fn() : false) {
		return {name: Constants.ROUTE_NAME_ENTRY_PAGE};
	}

	return null;
}

export const router = createRouter({
	history: createWebHistory(),
	routes: [
		{
			path: '/',
			name: Constants.ROUTE_NAME_ENTRY_PAGE,
			component: EntryPage,
		},
		{
			path: '/status/',
			name: Constants.ROUTE_NAME_STATUS_PAGE,
			component: StatusPage,
			meta: {limitToRole: 'privileged'},
		},
		{
			path: '/block/:blockchainMode/:id',
			name: Constants.ROUTE_NAME_BLOCK_PAGE,
			component: BlockPage,
		},
		{
			path: '/tx/:blockchainMode/:id',
			name: Constants.ROUTE_NAME_TRANSACTION_PAGE,
			component: TransactionPage,
		},
		{
			path: '/address/:blockchainMode/:id',
			name: Constants.ROUTE_NAME_ADDRESS_PAGE,
			component: AddressPage,
		},
		{
			path: '/workspace/:blockchainMode/:id',
			name: Constants.ROUTE_NAME_WORKSPACE_PAGE,
			component: WorkspaceEditorPage,
			meta: {limitToRole: 'privileged'},
		},
		{
			path: '/login',
			name: Constants.ROUTE_NAME_LOGIN_PAGE,
			component: LoginPage,
		},
		{
			// Wiki root page
			path: '/wiki',
			name: Constants.ROUTE_NAME_WIKI_ROOT,
			component: WikiPage,
			meta: {limitToRole: 'privileged'},
		},
		{
			// Wiki content page
			// allow additional slashes in path
			path: '/wiki/:file(.*)',
			name: Constants.ROUTE_NAME_WIKI,
			component: WikiPage,
			meta: {limitToRole: 'privileged'},
		},
		{
			path: '/recovery',
			name: Constants.ROUTE_NAME_ACCOUNT_RECOVERY,
			component: RecoveryPage,
		},
		{
			path: '/settings',
			component: SettingsPage,
			children: [
				{
					path: 'profile/:tabName?',
					name: Constants.ROUTE_NAME_USER_PROFILE_PAGE,
					component: ProfilePage,
				},
			],
		},
		{
			path: '/tools/',
			component: ToolsPage,
			meta: {limitToRole: 'privileged'},
			children: [
				{
					path: 'workspaces',
					name: Constants.ROUTE_NAME_WORKSPACES_PAGE,
					component: WorkspacePage,
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
			meta: {limitToRole: 'admin'},
		},
		{
			path: '/about',
			name: Constants.ROUTE_NAME_ABOUT,
			component: TextLoaderPage,
			props: {pageTitle: 'About', url: 'about.html'},
		},
		{
			path: '/privacy',
			name: Constants.ROUTE_NAME_PRIVACY,
			component: TextLoaderPage,
			props: {pageTitle: 'Privacy Policy', url: 'privacy_policy.html'},
		},
		{
			path: '/termsOfUse',
			name: Constants.ROUTE_NAME_TERMS_OF_USE,
			component: TextLoaderPage,
			props: {pageTitle: 'Terms of Use', url: 'terms_of_use.html'},
		},
		{
			path: '/noResults',
			name: Constants.ROUTE_NAME_NO_RESULTS,
			component: ErrorPage,
			props: {
				default: true,
				title: 'No results found!',
				description: 'Your search query did not return any results.',
				imageSource: NoResultsImg,
			},
		},
		{
			path: '/error',
			name: Constants.ROUTE_NAME_ERROR,
			component: ErrorPage,
			props: {
				default: true,
				title: 'Error',
				description: '',
				imageSource: BugsImg,
			},
		},
		{
			path: '/:catchAll(.*)',
			name: Constants.ROUTE_NAME_404_PAGE,
			component: ErrorPage,
			props: {
				default: true,
				title: '404 - Page not found!',
				description: 'The requested page does not exist.',
				imageSource: BugsImg,
			},
		},
	],
});

router.beforeEach((to, from) => {
	// Check for role
	if (to.meta.limitToRole) {
		let fn = null;
		if (to.meta.limitToRole === 'privileged') {
			fn = isPrivileged;
		} else if (to.meta.limitToRole === 'admin') {
			fn = isAdmin;
		}

		const routeTo = checkSession(to, fn);
		if (routeTo !== null) {
			return routeTo;
		}
	}

	if (from && to.name !== from.name) {
		// Clear all notifications belonging to the previous page
		msgStore.filterMessages(from.name);
	}

	return true;
});

