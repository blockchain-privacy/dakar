// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later
import {createRouter, createWebHistory} from 'vue-router';
import * as Constants from '../constants/index.js';
import {isAnyAdminIdentity, isAnyPrivilegedIdentity} from '@/utilities';
import {useLocalStore} from '@/pinia/local';
import {useNavStore} from '@/pinia/nav';
import NoResultsImg from '@/assets/no_results.webp';
import BugsImg from '@/assets/bugs.webp';

const EntryPage = () => import('../components/EntryPage.vue');
const SettingsPage = () => import('../components/user/SettingsPage.vue');
const ProfilePage = () => import('../components/user/ProfilePage.vue');
const AdministrationPage = () => import('../components/user/AdministrationPage.vue');
const LoginPage = () => import('../components/user/LoginPage.vue');
const TransactionPage = () => import('../components/explorer/transaction/TransactionPage.vue');
const BlockPage = () => import('../components/explorer/BlockPage.vue');
const AddressPage = () => import('../components/explorer/address/AddressPage.vue');
const WorkspaceEditorPage = () => import('../components/workspace/WorkspaceEditorPage.vue');
const StatusPage = () => import('../components/StatusPage.vue');
const ToolsPage = () => import('../components/tools/ToolsPage.vue');
const OAuthPage = () => import('../components/user/OAuthPage.vue');
const ClusterPage = () => import('../components/tools/clusters/ClusterPage.vue');
const AttributionsPage = () => import('../components/tools/attributions/AttributionsPage.vue');
const RecoveryPage = () => import('../components/user/RecoveryPage.vue');
const OAuthSuccessPage = () => import('../components/user/OAuthSuccessPage.vue');
const WikiPage = () => import('../components/wiki/WikiPage.vue');
const TextLoaderPage = () => import('../components/TextLoaderPage.vue');
const WorkspacePage = () => import('@/components/tools/workspaces/WorkspacePage.vue');
const ErrorPage = () => import('@/components/ErrorPage.vue');
const OAuthConsentPage = () => import('@/components/user/OAuthConsentPage.vue');
const OAuthVerificationPage = () => import('@/components/user/OAuthVerificationPage.vue');

let navStore = null;
let localStore = null;

// Call this right after the pinia store was created and added to the vue instance
export function setupStore() {
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
			path: '/updates',
			name: Constants.ROUTE_NAME_UPDATE_PAGE,
			component: () => import('../components/UpdatesPage.vue'),
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
			],
		},
		{
			path: '/userAdministration',
			name: Constants.ROUTE_NAME_USER_ADMIN_PAGE,
			component: AdministrationPage,
			meta: {limitToRole: 'admin'},
		},
		{
			path: '/oauth/',
			component: OAuthPage,
			children: [
				{
					path: 'login',
					name: Constants.ROUTE_NAME_OAUTH_LOGIN_PAGE,
					component: LoginPage,
					props: {
						default: true,
						title: 'Login with Dakar',
						isOAuth: true,
					},
				},
				{
					path: 'consent',
					name: Constants.ROUTE_NAME_OAUTH_CONSENT_PAGE,
					component: OAuthConsentPage,
				},
				{
					path: 'verification',
					name: Constants.ROUTE_NAME_OAUTH_VERIFICATION_PAGE,
					component: OAuthVerificationPage,
				},
				{
					path: 'success',
					name: Constants.ROUTE_NAME_OAUTH_SUCCESS_PAGE,
					component: OAuthSuccessPage,
				},
				{
					path: 'error',
					name: Constants.ROUTE_NAME_OAUTH_ERROR_PAGE,
					component: ErrorPage,
					props: {
						default: true,
						title: 'Authentication Error',
						hideActions: true,
						description: 'While authenticating an error occurred. Close this page and try again.',
						imageSource: NoResultsImg,
					},
				},
			],
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

router.beforeEach(to => {
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

	return true;
});

