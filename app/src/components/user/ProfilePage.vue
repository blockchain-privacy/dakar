<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div
    class="mx-auto"
    style="max-width: 1200px;"
  >
    <alert :text="errorMsg" />
    <icon-title
      title="Settings"
      one-line
    >
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
    </icon-title>
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
    <p class="text-headline-small mb-5">
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
          @click="deleteUserSession(item.id)"
        >
          {{ mdiDelete }}
        </v-icon>
      </template>
      <template #no-data>
        No other active sessions found
      </template>
    </v-data-table>
    <p class="text-headline-small mb-5">
      External Sessions - OAuth 2.0
    </p>
    <v-data-table
      v-model:sort-by="consentSessionSortBy"
      class="mb-10"
      :headers="consentSessionHeaders"
      :items="consentSessions?consentSessions:[]"
      :loading="consentSessionsLoading"
    >
      <template #item.grant_scope="{ item }">
        <span>{{ item.grant_scope.join(', ') }}</span>
      </template>
      <template #item.handled_at="{ item }">
        <span>{{ new Date(item.handled_at).toLocaleString() }}</span>
      </template>
      <template #item.actions="{ item }">
        <v-icon
          size="small"
          @click="deleteConsentSession(item.consent_request.consent_request_id)"
        >
          {{ mdiDelete }}
        </v-icon>
      </template>
      <template #no-data>
        No sessions found
      </template>
    </v-data-table>
    <v-dialog
      v-model="showAccountDeletionDialog"
      max-width="700px"
    >
      <v-card>
        <v-card-title>Delete Account</v-card-title>
        <v-card-text>
          <p class="text-body-large">
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
import {
	computed,
	inject,
	onMounted,
	ref,
	watch,
} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {storeToRefs} from 'pinia';
import OryFlow from './ory/OryFlow.vue';
import {PAGE_TITLE, ROUTE_NAME_ENTRY_PAGE, ROUTE_NAME_USER_PROFILE_PAGE} from '@/constants';
import handleGetFlowError from '@/kratos';
import {useLocalStore} from '@/pinia/local';
import {useNavStore} from '@/pinia/nav';
import Alert from '@/components/common/Alert.vue';
import IconTitle from '@/components/common/IconTitle.vue';

const ory = inject('ory');
const kratosAdmin = inject('kratosadmin');
const route = useRoute();
const router = useRouter();
const localStore = useLocalStore();
const navStore = useNavStore();
const {session} = storeToRefs(localStore);
const context = {
	$route: route, $router: router, navStore, localStore,
};

const errorMsg = ref('');
const disabledForms = ref([]);
const showAccountDeletionDialog = ref(false);
const settingsFlow = ref(null);

const userSessions = ref([]);
const userSessionsLoading = ref(false);
const userSessionSortBy = ref([{key: 'authenticated_at', order: 'desc'}]);
const userSessionHeaders = [
	{title: 'Authentication Date', key: 'authenticatedAt', align: 'start'},
	{title: 'Expiration Date', key: 'expiresAt'},
	{title: 'Device', key: 'userAgent'},
	{title: 'IP Address', key: 'ipAddress'},
	{
		title: '', key: 'actions', sortable: false, align: 'end',
	},
];

const consentSessions = ref([]);
const consentSessionsLoading = ref(false);
const consentSessionSortBy = ref([{key: 'handled_at', order: 'desc'}]);
const consentSessionHeaders = [
	{title: 'Client Name', key: 'consent_request.client.client_name', align: 'start'},
	{title: 'Permissions', key: 'grant_scope'},
	{title: 'Handled At', key: 'handled_at'},
	{
		title: '', key: 'actions', sortable: false, align: 'end',
	},
];

// Computed
const isAccountRecovery = computed(() => settingsFlow.value?.ui?.messages?.some(m => m.id === 1_060_001));

// Watchers
watch(route, to => {
	if (to.name === ROUTE_NAME_USER_PROFILE_PAGE && !to.query.flow) {
		// This happens if the users manually navigates to the route of this page,
		// in this case flow is not set and needs to be reinitialized
		initFlow();
	}
});

// Hooks
onMounted(async () => {
	document.title = `Profile - ${PAGE_TITLE}`;

	await initFlow();
	await getSessions();
	await getConsentSessions();
});

// Functions
async function doErrorHandling(ctx, error, onRefreshFlow, isOAuth) {
	try {
		const err = await handleGetFlowError(ctx, error, onRefreshFlow, isOAuth);
		if (err) {
			errorMsg.value = err;
		}
	} catch (error_) {
		errorMsg.value = error_.message;
	}
}

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

async function deleteIdentity() {
	errorMsg.value = '';
	try {
		await kratosAdmin.identity.selfIdentitiesDelete();
		session.value = null;
		await router.push({name: ROUTE_NAME_ENTRY_PAGE});
	} catch (error) {
		errorMsg.value = error.message;
	}

	showAccountDeletionDialog.value = false;
}

async function deleteUserSession(id) {
	if (!id) {
		return;
	}

	errorMsg.value = '';

	try {
		await ory.frontend.disableMySession({id});
		await getSessions();
	} catch (error) {
		errorMsg.value = error.message;
	}
}

async function deleteConsentSession(consentId) {
	if (!consentId) {
		return;
	}

	errorMsg.value = '';

	try {
		await kratosAdmin.oauth.selfConsentsConsentIdDelete({consentId});
		await getConsentSessions();
	} catch (error) {
		errorMsg.value = error.message;
	}
}

async function getSessions() {
	if (isAccountRecovery.value) {
		// Can't get session during account recovery
		return;
	}

	userSessionsLoading.value = true;
	errorMsg.value = '';

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
	} catch (error) {
		errorMsg.value = error.message;
	}

	userSessionsLoading.value = false;
}

async function getConsentSessions() {
	if (isAccountRecovery.value) {
		// Can't get consent session during account recovery
		return;
	}

	consentSessionsLoading.value = true;
	errorMsg.value = '';

	try {
		const response = await kratosAdmin.oauth.selfConsentsGet();

		consentSessions.value = response.oauthSessions;
	} catch (error) {
		errorMsg.value = error.message;
	}

	consentSessionsLoading.value = false;
}

async function initSettingsFlow() {
	errorMsg.value = '';
	try {
		const response = await ory.frontend.createBrowserSettingsFlow();
		setFlowData(response);
	} catch (error) {
		await doErrorHandling(context, error, null);
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

	errorMsg.value = '';

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
		// Remove 'traits.' from key
		traits[key.slice(7)] = value;
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
			errorMsg.value = response.error.reason;
		}
	} catch (error) {
		if (error.response?.ui) {
			setFlowData(error.response);
		} else {
			await doErrorHandling(context, error, async () => {
				await initSettingsFlow();
				errorMsg.value = 'The settings flow has expired, please try again.';
			});
		}
	}

	// Enable submitting for this form again
	disabledForms.value = disabledForms.value.filter(d => d !== formID);
}

async function refreshSession() {
	errorMsg.value = '';
	try {
		session.value = await ory.frontend.toSession();
	} catch (error) {
		await doErrorHandling(context, error, null);
	}
}

async function tryRefreshSession() {
	let success;

	try {
		session.value = await ory.frontend.toSession();
		success = true;
	} catch {
		success = false;
	}

	return success;
}

async function initFlow() {
	const {flow} = route.query;
	errorMsg.value = '';

	if (typeof flow === 'string') {
		try {
			const response = await ory.frontend.getSettingsFlow({id: flow});
			setFlowData(response);

			if (!isAccountRecovery.value) {
				await tryRefreshSession();
			}
		} catch (error) {
			await doErrorHandling(context, error, initSettingsFlow);
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
