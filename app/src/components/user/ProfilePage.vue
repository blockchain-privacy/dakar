<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div
    class="mx-auto"
    style="max-width: 1200px;"
  >
    <p class="text-h5 my-5 d-flex align-center justify-space-between">
      Settings
      <v-menu>
        <template #activator="{ props }">
          <v-btn
            v-bind="props"
            variant="text"
            :icon="mdiDotsVertical"
          />
        </template>
        <v-list>
          <v-list-item
            class="text-justify"
            @click="showAccountDeletionDialog=true"
          >
            <v-list-item-title class="d-flex align-center">
              <v-icon
                color="red"
                start
              >
                {{ mdiAlert }}
              </v-icon>
              Delete Account
            </v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </p>
    <div
      v-if="settingsFlow"
      style="max-width: 700px;"
      class="mx-auto"
    >
      <v-card variant="text">
        <ory-flow
          class="mt-4"
          :flow="settingsFlow"
          form-id="settings-form"
          :disabled-forms="disabledForms"
          embed
          @submit="handleOrySubmitSettings"
        />
      </v-card>
    </div>
    <v-skeleton-loader
      v-else
      class="mx-auto"
      type="article, actions"
    />
    <v-divider
      thickness="3"
      class="mt-5 mb-15"
    />
    <p class="text-h5 mb-5">
      Sessions
    </p>
    <v-data-table
      v-model:sort-by="userSessionSortBy"
      class="mb-10"
      :headers="userSessionHeaders"
      :items="userSessions?userSessions:[]"
      :loading="userSessionsLoading"
    >
      <template #item.authenticatedAt="{ item }">
        <span>{{ new Date(item.authenticatedAt).toLocaleString() }}</span>
      </template>
      <template #item.expiresAt="{ item }">
        <span>{{ new Date(item.expiresAt).toLocaleString() }}</span>
      </template>
      <template #item.userAgent="{ item }">
        <v-icon>{{ getDeviceIcon(item.userAgent) }}</v-icon>
        <span class="ms-2">{{ item.userAgent }}</span>
      </template>
      <template #item.actions="{ item }">
        <v-icon
          size="small"
          @click="deleteUserSession(item)"
        >
          {{ mdiDelete }}
        </v-icon>
      </template>
      <template #no-data>
        No other active sessions found
      </template>
    </v-data-table>
    <v-dialog
      v-model="showAccountDeletionDialog"
      max-width="700px"
    >
      <v-card>
        <v-card-title>Delete Account</v-card-title>
        <v-card-text>
          <p class="text-subtitle-1">
            Do you really want to delete your account? This action can not be reversed.
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="showAccountDeletionDialog = false">
            Cancel
          </v-btn>
          <v-btn
            color="red"
            @click="deleteIdentity()"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import {
	mdiAlert,
	mdiAndroid,
	mdiApple,
	mdiDelete,
	mdiDotsVertical,
	mdiLaptop,
	mdiLinux,
	mdiMicrosoftWindows,
} from '@mdi/js';
import {PAGE_TITLE, ROUTE_NAME_ENTRY_PAGE, ROUTE_NAME_USER_PROFILE_PAGE} from '@/constants';
import OryFlow from './ory/OryFlow.vue';
import handleGetFlowError from '@/kratos';
import {handleError} from '@/utilities';
import {
	computed, inject, onMounted, ref, watch,
} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {useLocalStore} from '@/pinia/local';
import {useNavStore} from '@/pinia/nav';
import {useMsgStore} from '@/pinia/msg';

const ory = inject('ory');
const kratosAdmin = inject('kratosadmin');
const route = useRoute();
const router = useRouter();
const localStore = useLocalStore();
const navStore = useNavStore();
const msgStore = useMsgStore();
const context = {
	$route: route, $router: router, navStore, localStore, msgStore, addMessage: msgStore.addMessage,
};

const settingsFlow = ref(null);
const userSessions = ref([]);
const userSessionsLoading = ref(false);
const disabledForms = ref([]);
const showAccountDeletionDialog = ref(false);
const userSessionSortBy = ref([{key: 'authenticated_at', order: 'desc'}]);
const userSessionHeaders = [
	{
		title: 'Authentication Date', key: 'authenticatedAt', align: 'start',
	},
	{
		title: 'Expiration Date', key: 'expiresAt',
	},
	{
		title: 'Device', key: 'userAgent',
	},
	{
		title: 'IP Address', key: 'ipAddress',
	},
	{
		title: '', key: 'actions', sortable: false,
	},
];

// Computed
const session = computed({
	get() {
		return localStore.getSession;
	},
	set(value) {
		localStore.setSession(value);
	},
});

// Watchers
watch(route, to => {
	if (to.name === ROUTE_NAME_USER_PROFILE_PAGE && !to.query.flow) {
		// This happens if the users manually navigates to the route of this page,
		// in this case flow is not set and needs to be reinitialized
		initFlow();
	}
});

// Hooks
onMounted(() => {
	document.title = `Profile - ${PAGE_TITLE}`;

	// Init the flow and get sessions in parallel
	Promise.all([initFlow(), getSessions()]);
});

// Functions
function getDeviceIcon(userAgent) {
	if (userAgent.length === 0) {
		return '';
	}

	const ua = userAgent.toLowerCase();
	if (ua.includes('linux')) {
		return mdiLinux;
	}

	if (ua.includes('android')) {
		return mdiAndroid;
	}

	if (ua.includes('iphone') || ua.includes('mac')) {
		return mdiApple;
	}

	if (ua.includes('windows')) {
		return mdiMicrosoftWindows;
	}

	return mdiLaptop;
}

function setSuccessMessage(msg) {
	// Do not limit message to current route
	msgStore.addMessage({text: msg, type: 'success', temporary: true});
}

function setErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: true, category: route.name,
	});
}

async function deleteIdentity() {
	try {
		await kratosAdmin.selfDelete();
		msgStore.resetMessages();
		setSuccessMessage('Your account was successfully deleted. Goodbye!');
		session.value = null;
		await router.push({name: ROUTE_NAME_ENTRY_PAGE});
	} catch (e) {
		handleError(context, e);
	}

	showAccountDeletionDialog.value = false;
}

async function deleteUserSession(session) {
	if (!session.id) {
		return;
	}

	try {
		await ory.frontend.disableMySession({id: session.id});
		userSessions.value = userSessions.value.filter(d => d.id !== session.id);
	} catch (e) {
		handleError(context, e);
	}
}

async function getSessions() {
	userSessionsLoading.value = true;

	try {
		// Get a maximum of 30 sessions
		const response = await ory.frontend.listMySessions({page: 1, perPage: 30});

		userSessions.value = response.map(d => {
			d.authenticatedAt = new Date(d.authenticated_at).getTime();
			d.expiresAt = new Date(d.expires_at).getTime();
			if (d.devices?.length > 0) {
				if (d.devices[0].user_agent) {
					d.userAgent = d.devices[0].user_agent;
				}

				if (d.devices[0].ip_address) {
					d.ipAddress = d.devices[0].ip_address.split(':')[0];
				}
			}

			return d;
		});
	} catch (e) {
		handleError({addMessage: msgStore.addMessage, $route: route}, e);
	}

	userSessionsLoading.value = false;
}

async function initSettingsFlow() {
	try {
		const response = await ory.frontend.createBrowserSettingsFlow();
		setFlowData(response);
	} catch (e) {
		await handleGetFlowError(context, e, null);
	}
}

function setFlowData(d) {
	settingsFlow.value = d;
	if (!route.query.flow || route.query.flow !== d.id) {
		router.replace({query: {flow: d.id}});
	}
}

async function handleOrySubmitSettings(formID) {
	const form = document.getElementById(formID);
	if (!form || !settingsFlow.value.ui.action) {
		return;
	}

	// Disable submitting from this form
	disabledForms.value.push(formID);

	const body = Object.fromEntries(new FormData(form));

	// Extract traits into object
	const traits = {};
	let foundTrait = false;
	for (const [key, value] of Object.entries(body)) {
		if (!key.startsWith('traits.')) {
			continue;
		}

		foundTrait = true;
		traits[key.substring(7, key.length)] = value;
		delete body[key];
	}

	if (foundTrait) {
		body.traits = traits;
	}

	const {flow} = route.query;

	try {
		const response = await ory.frontend.updateSettingsFlow({flow, updateSettingsFlowBody: body});

		// Something went wrong and we need to display some data
		if (response?.ui) {
			setFlowData(response);
		}

		// If an account is being recovered the session is empty,
		// therefore it has to be refreshed.
		await refreshSession();

		if (response.error && response.error.reason) {
			setErrorMessage(response.error.reason);
		}
	} catch (e) {
		if (e.response?.ui) {
			setFlowData(e.response);
		} else {
			handleGetFlowError(context, e, async () => {
				await initSettingsFlow();
				setErrorMessage('The settings flow has expired, please try again.');
			}).catch(e => {
				setErrorMessage(e);
			});
		}
	}

	// Enable submitting for this form again
	disabledForms.value = disabledForms.value.filter(d => d !== formID);
}

async function refreshSession() {
	try {
		const response = await ory.frontend.toSession();
		session.value = response;
	} catch (e) {
		await handleGetFlowError(context, e, null);
	}
}

async function tryRefreshSession() {
	let success = false;

	try {
		const response = await ory.frontend.toSession();
		session.value = response;
		success = true;
	} catch (_) {
		success = false;
	}

	return success;
}

async function initFlow() {
	const {flow} = route.query;

	if (typeof flow === 'string') {
		try {
			const response = await ory.frontend.getSettingsFlow({id: flow});
			setFlowData(response);

			// Try to refresh session. This might fail if the identity
			// is in the process of being recovered and aal2 is set.
			await tryRefreshSession();
		} catch (e) {
			await handleGetFlowError(context, e, initSettingsFlow);
		}
	} else {
		// If there's no flow in our route,
		// we need to initialize our login flow
		await initSettingsFlow();
	}
}

</script>

<style scoped>

</style>
