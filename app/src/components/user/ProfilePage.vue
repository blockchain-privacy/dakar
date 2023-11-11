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
            :icon="icons.mdiDotsVertical"
            variant="text"
            v-bind="props"
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
                class="me-2"
              >
                {{ icons.mdiAlert }}
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
          :form-id="formID"
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
      User sessions
    </p>

    <v-data-table
      v-model:sort-by="userSessionSortBy"
      class="mb-10"
      :headers="userSessionHeaders"
      :items="userSessions?userSessions:[]"
      :loading="!userSessions"
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
          {{ icons.mdiDelete }}
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

<script>
import {mdiAccountDetails, mdiDelete, mdiAlert, mdiDotsVertical,
	mdiLinux, mdiAndroid, mdiApple, mdiLaptop, mdiMicrosoftWindows,
} from '@mdi/js';
import {
	PAGE_TITLE, ROUTE_NAME_ENTRY_PAGE, ROUTE_NAME_USER_PROFILE_PAGE,
} from '@/constants';
import OryFlow from './ory/OryFlow.vue';
import handleGetFlowError from '@/kratos';
import {handleError} from '@/utilities';

export default {
	name: 'ProfilePage',
	components: {OryFlow},
	data() {
		return {
			icons: {mdiAccountDetails, mdiDelete, mdiAlert, mdiDotsVertical,
				mdiLinux, mdiAndroid, mdiApple, mdiLaptop, mdiMicrosoftWindows},
			formID: 'settings-form',
			settingsFlow: null,
			userSessions: [],
			userSessionsLoading: false,
			disabledForms: [],
			showAccountDeletionDialog: false,
			route: {
				rootPage: ROUTE_NAME_ENTRY_PAGE,
			},
			userSessionSortBy: [{key: 'authenticated_at', order: 'desc'}],
			userSessionHeaders: [
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
			],
		};
	},
	computed: {
		session: {
			get() {
				return this.$store.getters.getSession;
			},
			set(value) {
				this.$store.dispatch('setSession', value);
			},
		},
	},
	watch: {
		$route(to) {
			if (to.name === ROUTE_NAME_USER_PROFILE_PAGE && !to.query.flow) {
				// This happens if the users manually navigates to the route of this page,
				// in this case flow is not set and needs to be reinitialized
				this.initFlow();
			}
		},
	},
	mounted() {
		document.title = `Profile - ${PAGE_TITLE}`;

		// Init the flow and get sessions in parallel
		Promise.all([this.initFlow(), this.getSessions()]);
	},
	methods: {
		getDeviceIcon(userAgent) {
			if (userAgent.length === 0) {
				return '';
			}

			const ua = userAgent.toLowerCase();
			if (ua.includes('linux')) {
				return this.icons.mdiLinux;
			}

			if (ua.includes('android')) {
				return this.icons.mdiAndroid;
			}

			if (ua.includes('iphone') || ua.includes('mac')) {
				return this.icons.mdiApple;
			}

			if (ua.includes('windows')) {
				return this.icons.mdiMicrosoftWindows;
			}

			return this.icons.mdiLaptop;
		},
		setSuccessMessage(msg) {
			// Do not limit message to current route
			this.$store.dispatch('addMessage', {text: msg, type: 'success', temporary: true});
		},
		setErrorMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'error', temporary: true, category: this.$route.name});
		},
		async deleteIdentity() {
			try {
				await this.dakar.authentication.deleteIdentityGet();
				this.$store.dispatch('resetMessages');
				this.setSuccessMessage('Your account was successfully deleted. Goodbye!');
				this.session = null;
				this.$router.push({name: this.route.rootPage});
			} catch (e) {
				handleError(this, e);
			}

			this.showAccountDeletionDialog = false;
		},
		async deleteUserSession(session) {
			if (!session.id) {
				return;
			}

			try {
				const response = await this.ory.frontend.disableMySession({id: session.id});

				if (response.status === 204) {
					this.userSessions = this.userSessions.filter(d => d.id !== session.id);
				} else {
					throw new Error('unable to delete session');
				}
			} catch (e) {
				handleError(this, e);
			}
		},
		async getSessions() {
			this.userSessionsLoading = true;

			try {
				// Get a maximum of 30 sessions
				const response = await 	this.ory.frontend.listMySessions({page: 1, perPage: 30});

				this.userSessions = response.data.map(d => {
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
				handleError(this, e);
			}

			this.userSessionsLoading = false;
		},
		async initSettingsFlow() {
			try {
				const response = await this.ory.frontend.createBrowserSettingsFlow();
				this.setFlowData(response.data);
			} catch (e) {
				await handleGetFlowError(this, e, null);
			}
		},
		setFlowData(d) {
			this.settingsFlow = d;
			if (!this.$route.query.flow || this.$route.query.flow !== d.id) {
				this.$router.replace({query: {flow: d.id}});
			}
		},
		async handleOrySubmitSettings(formID) {
			const form = document.getElementById(formID);
			if (!form || !this.settingsFlow.ui.action) {
				return;
			}

			// Disable submitting from this form
			this.disabledForms.push(formID);

			const body = Object.fromEntries(new FormData(form));
			const {flow} = this.$route.query;

			try {
				const response = await this.ory.frontend.updateSettingsFlow({flow, updateSettingsFlowBody: body});

				// Something went wrong and we need to display some data
				if (response.data && response.data.ui) {
					this.setFlowData(response.data);
				}

				// If an account is being recovered the session is empty,
				// therefore it has to be refreshed.
				await this.refreshSession();

				if (response.error && response.error.reason) {
					this.setErrorMessage(response.error.reason);
				}
			} catch (e) {
				if (e.response?.data?.ui) {
					this.setFlowData(e.response.data);
				} else {
					handleGetFlowError(this, e, async () => {
						await this.initSettingsFlow();
						this.setErrorMessage('The settings flow has expired, please try again.');
					}).catch(e => {
						this.setErrorMessage(e);
					});
				}
			}

			// Enable submitting for this form again
			this.disabledForms = this.disabledForms.filter(d => d !== formID);
		},
		async refreshSession() {
			try {
				const response = await this.ory.frontend.toSession();

				if (response.status === 200) {
					this.session = response.data;
				}
			} catch (e) {
				await handleGetFlowError(this, e, null);
			}
		},
		async tryRefreshSession() {
			let success = false;

			try {
				const response = await 	this.ory.frontend.toSession();
				if (response.status === 200) {
					this.session = response.data;
					success = true;
				}
			} catch (_) {
				success = false;
			}

			return success;
		},
		async initFlow() {
			const {flow} = this.$route.query;

			if (typeof flow === 'string') {
				try {
					const response = await this.ory.frontend.getSettingsFlow({id: flow});
					this.setFlowData(response.data);

					// Try to refresh session. This might fail if the identity
					// is in the process of being recovered and aal2 is set.
					await this.tryRefreshSession();
				} catch (e) {
					await handleGetFlowError(this, e, this.initSettingsFlow);
				}
			} else {
				// If there's no flow in our route,
				// we need to initialize our login flow
				await this.initSettingsFlow();
			}
		},
	},
};
</script>

<style scoped>

</style>
