<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div
    class="d-flex align-center justify-center"
    style="height: 100%; width:100%"
  >
    <v-card
      max-width="600px"
      style="flex:1"
    >
      <div class="pa-5">
        <h3 class="text-h3 font-weight-bold text-center mb-2">
          Login
        </h3>
        <ory-flow
          v-if="loginFlow"
          :flow="loginFlow"
          form-id="login-form"
          :disabled-forms="disabledForms"
          class="mt-3"
          @submit="handleOrySubmitLogin"
        />
        <v-skeleton-loader
          v-else
          class="mx-auto"
          type="article, actions"
        />
        <div class="d-flex align-center mt-2">
          <v-btn
            class="ms-auto"
            variant="text"
            size="small"
            @click="logoutAndGoToPage({name: ROUTE_NAME_ACCOUNT_RECOVERY})"
          >
            Recover account
          </v-btn>
          <v-btn
            v-if="showLogoutButton"
            variant="text"
            size="small"
            color="red"
            @click="logoutAndGoToPage({name: ROUTE_NAME_ENTRY_PAGE})"
          >
            Log out
          </v-btn>
        </div>
      </div>
    </v-card>
  </div>
</template>

<script setup>
import {
	PAGE_TITLE, ROUTE_NAME_ENTRY_PAGE, ROUTE_NAME_ACCOUNT_RECOVERY, ROUTE_NAME_LOGIN_PAGE,
} from '@/constants';
import handleGetFlowError from '@/kratos';
import OryFlow from './ory/OryFlow.vue';
import {
	computed, inject, onMounted, ref, watch,
} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local';
import {useNavStore} from '@/pinia/nav';
import {useMsgStore} from '@/pinia/msg';

const ory = inject('ory');
const router = useRouter();
const route = useRoute();
const localStore = useLocalStore();
const navStore = useNavStore();
const msgStore = useMsgStore();
const {failedRoute} = storeToRefs(navStore);
const context = {
	$route: route, $router: router, navStore, localStore, msgStore,
};

const loginFlow = ref(null);
const showLogoutButton = ref(false);
const disabledForms = ref([]);

// Computed
const session = computed({
	get() {
		return localStore.getSession;
	},
	set(value) {
		localStore.setSession(value);
	},
});

// Watch
watch(route, to => {
	if (to.name === ROUTE_NAME_LOGIN_PAGE && !to.query.flow) {
		// This happens if the users manually navigates to the route of this page,
		// in this case flow is not set and needs to be reinitialized
		initFlow();
	}
});

// Hooks
onMounted(() => {
	document.title = `Login - ${PAGE_TITLE}`;

	// Check if flow id is set
	if (route.query.flow) {
		initFlow();
		return;
	}

	// If session is not set, user might be logged in already -> get session
	if (session.value) {
		leave();
	} else {
		tryToGetSession();
	}
});

// Functions
function setErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: true, category: route.name,
	});
}

function goToPage(pageObj) {
	router.push(pageObj);
}

async function tryToGetSession() {
	try {
		const response = await ory.frontend.toSession();
		session.value = response;
		leave();
	} catch (e) {
		if (e.response?.error?.id === 'session_aal2_required') {
			await initLoginFlow('aal2');
			return;
		}

		// This request fails if the user is not logged in -> init login form
		await initFlow();
	}
}

function leave() {
	if (loginFlow.value?.return_to) {
		window.location.href = loginFlow.value.return_to;
		return;
	}

	if (failedRoute.value !== null && failedRoute.value.name !== ROUTE_NAME_LOGIN_PAGE) {
		goToPage(failedRoute.value);
		failedRoute.value = null;
		return;
	}

	goToPage({name: ROUTE_NAME_ENTRY_PAGE});
}

// Used to break login flow (when aal2 or higher is required) and go to a different page
async function logoutAndGoToPage(toObj) {
	try {
		const response = await ory.frontend.createBrowserLogoutFlow();
		if (!response.logout_token) {
			return;
		}

		await ory.frontend.updateLogoutFlow({token: response.logout_token});
		session.value = null;
		goToPage(toObj);
	} catch (e) {
		// Could not log out because no session was found -> go to requested page
		if (e.response?.error?.id === 'session_inactive') {
			goToPage(toObj);
		} else {
			await handleGetFlowError(context, e, null);
		}
	}
}

async function handleOrySubmitLogin(formID) {
	const form = document.getElementById(formID);
	if (!form || !loginFlow.value.ui.action) {
		return;
	}

	// Disable submitting from this form
	disabledForms.value.push(formID);

	const body = Object.fromEntries(new FormData(form));
	const {flow} = route.query;

	try {
		const response = await ory.frontend.updateLoginFlow({flow, updateLoginFlowBody: body});

		if (response?.session?.identity) {
			session.value = response.session;
			leave();
			return;
		}

		// Something went wrong and we need to display some data
		if (response?.ui) {
			setFlowData(response);
		}

		if (response.error && response.error.reason) {
			setErrorMessage(response.error.reason);
		}
	} catch (e) {
		if (e.response?.ui) {
			setFlowData(e.response);
		} else {
			handleGetFlowError(context, e, () => {
				initLoginFlow('aal1');
			}).catch(e => {
				setErrorMessage(e);
			});
		}
	} finally {
		// Enable submitting for this form again
		disabledForms.value = disabledForms.value.filter(d => d !== formID);
	}
}

async function initFlow() {
	const {flow} = route.query;

	if (typeof flow === 'string') {
		try {
			const response = await ory.frontend.getLoginFlow({id: flow});
			setFlowData(response);
		} catch (e) {
			await handleGetFlowError(context, e, () => initLoginFlow('aal1'));
		}
	} else {
		// If there's no flow in our route,
		// we need to initialize our login flow
		await initLoginFlow('aal1');
	}
}

async function initLoginFlow(aal) {
	try {
		const response = await ory.frontend.createBrowserLoginFlow({refresh: false, aal});
		setFlowData(response);
	} catch (e) {
		await handleGetFlowError(context, e, null);
	}
}

function setFlowData(d) {
	loginFlow.value = d;
	showLogoutButton.value = Boolean(d.requested_aal) && d.requested_aal !== 'aal1';
	if (!route.query.flow || route.query.flow !== d.id) {
		router.replace({query: {flow: d.id}});
	}
}

</script>

<style scoped>

</style>
