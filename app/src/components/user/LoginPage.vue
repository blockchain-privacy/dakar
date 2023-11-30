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
            @click="logoutAndGoToPage(ROUTE_NAME_ACCOUNT_RECOVERY)"
          >
            Recover account
          </v-btn>
          <v-btn
            v-if="showLogoutButton"
            variant="text"
            size="small"
            color="red"
            @click="logoutAndGoToPage(ROUTE_NAME_ENTRY_PAGE)"
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
import {inject, onMounted, ref, watch} from 'vue';
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
const {session} = storeToRefs(localStore);
const {failedRoute} = storeToRefs(navStore);
const context = {$route: route, $router: router, navStore, localStore, msgStore};

const loginFlow = ref(null);
const showLogoutButton = ref(false);
const disabledForms = ref([]);

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
	msgStore.addMessage({text: msg, type: 'error', temporary: true, category: route.name});
}

function goToPage(pageObj) {
	router.push(pageObj);
}

async function tryToGetSession() {
	try {
		const response = await ory.frontend.toSession();
		if (response.status === 200) {
			session.value = response.data;
			leave();
		}
	} catch (e) {
		if (e.response?.data?.error?.id === 'session_aal2_required') {
			await initLoginFlow('aal2');
			return;
		}

		// This request fails if the user is not logged in -> init login form
		await initFlow();
	}
}

function leave() {
	if (loginFlow.value && loginFlow.value.return_to) {
		window.location.href = loginFlow.value.return_to;
	} else if (failedRoute.value) {
		goToPage(failedRoute);
		failedRoute.value = null;
	} else {
		goToPage({name: ROUTE_NAME_ENTRY_PAGE});
	}
}

// Used to break login flow (when aal2 or higher is required) and go to a different page
async function logoutAndGoToPage(pageName) {
	try {
		const response = await ory.frontend.createBrowserLogoutFlow();
		if (!response.data.logout_token) {
			return;
		}

		const newResponse = await ory.frontend.updateLogoutFlow({token: response.data.logout_token});

		if (newResponse.status === 204) {
			session.value = null;
			goToPage({name: pageName});
		}
	} catch (e) {
		// Could not log out because no session was found -> go to requested page
		if (e.response?.data?.error?.id === 'session_inactive') {
			goToPage({name: pageName});
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

		if (response.status === 200 && response.data?.session) {
			// Reminder: check when https://github.com/ory/kratos/pull/3572 is released
			if (response.data.session.identity) {
				session.value = response.data.session;
				leave();
				return;
			}

			// Aal2 not done yet
			await initLoginFlow('aal2');

			return;
		}

		// Something went wrong and we need to display some data
		if (response.data && response.data.ui) {
			setFlowData(response.data);
		}

		if (response.error && response.error.reason) {
			setErrorMessage(response.error.reason);
		}
	} catch (e) {
		if (e.response?.data?.ui) {
			setFlowData(e.response.data);
		} else {
			handleGetFlowError(context, e, () => {
				initLoginFlow('aal1');
				setErrorMessage('The login flow has expired, please try again.');
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
			setFlowData(response.data);
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
	// Set refresh to true, for the case when the local
	// session data was deleted but the user is still logged in

	try {
		const response = await ory.frontend.createBrowserLoginFlow({refresh: false, aal});
		setFlowData(response.data);
	} catch (e) {
		if (e.response?.data?.error?.id === 'session_already_available') {
			// Reminder: check when https://github.com/ory/kratos/pull/3572 is released
			// If the response indicates that the session is already available,
			// it might be only aal1 even though aal2 might be required.
			// This can be checked by requesting the session. If it fails the aal2 dialog will be rendered
			await tryToGetSession();
			return;
		}

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
