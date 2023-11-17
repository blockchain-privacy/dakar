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
          :form-id="formID"
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
            @click="logoutAndGoToPage(routeAccountRecovery)"
          >
            Recover account
          </v-btn>
          <v-btn
            v-if="showLogoutButton"
            variant="text"
            size="small"
            color="red"
            @click="logoutAndGoToPage(routeEntryPage)"
          >
            Log out
          </v-btn>
        </div>
      </div>
    </v-card>
  </div>
</template>

<script>
import {
	mdiLockOutline, mdiEye, mdiEyeOff, mdiEmail,
} from '@mdi/js';
import {
	APPLICATION_NAME, PAGE_TITLE, PASSWORD_MIN_CHARACTERS, ROUTE_NAME_ENTRY_PAGE,
	PASSWORD_MAX_CHARACTERS, APPLICATION_SUBTITLE, ROUTE_NAME_ACCOUNT_RECOVERY, ROUTE_NAME_LOGIN_PAGE,
} from '@/constants';
import {emailRules, passwordRules} from '@/utilities';
import handleGetFlowError from '@/kratos';
import OryFlow from './ory/OryFlow.vue';

export default {
	name: 'LoginPage',
	components: {OryFlow},
	data() {
		return {
			icon: {
				mdiLockOutline, mdiEye, mdiEyeOff, mdiEmail,
			},
			isSubmittingForm: false,
			loginFailed: false,
			applicationName: APPLICATION_NAME,
			applicationSubtitle: APPLICATION_SUBTITLE,
			passwordMinCharacters: PASSWORD_MIN_CHARACTERS,
			passwordMaxCharacters: PASSWORD_MAX_CHARACTERS,
			routeAccountRecovery: ROUTE_NAME_ACCOUNT_RECOVERY,
			routeEntryPage: ROUTE_NAME_ENTRY_PAGE,
			rules: {passwordRules, emailRules},
			email: {
				value: '',
			},
			password: {
				value: '',
				show: false,
			},
			loginFlow: null,
			showLogoutButton: false,
			formID: 'login-form',
			orySession: null,
			disabledForms: [],
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
		failedRoute: {
			get() {
				return this.$store.getters.getFailedRoute;
			},
			set(value) {
				this.$store.dispatch('setFailedRoute', value);
			},
		},
	},
	watch: {
		$route(to) {
			if (to.name === ROUTE_NAME_LOGIN_PAGE && !to.query.flow) {
				// This happens if the users manually navigates to the route of this page,
				// in this case flow is not set and needs to be reinitialized
				this.initFlow();
			}
		},
	},
	mounted() {
		document.title = `Login - ${PAGE_TITLE}`;

		// Check if flow id is set
		if (this.$route.query.flow) {
			this.initFlow();
			return;
		}

		// If session is not set, user might be logged in already -> get session
		if (this.session) {
			this.leave();
		} else {
			this.tryToGetSession();
		}
	},
	methods: {
		setErrorMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'error', temporary: true, category: this.$route.name});
		},
		goToPage(pageObj) {
			this.$router.push(pageObj);
		},
		async tryToGetSession() {
			try {
				const response = await this.ory.frontend.toSession();
				if (response.status === 200) {
					this.session = response.data;
					this.leave();
				}
			} catch (e) {
				if (e.response?.data?.error?.id === 'session_aal2_required') {
					await this.initLoginFlow('aal2');
					return;
				}

				// This request fails if the user is not logged in -> init login form
				await this.initFlow();
			}
		},
		leave() {
			if (this.loginFlow && this.loginFlow.return_to) {
				window.location.href = this.loginFlow.return_to;
			} else if (this.failedRoute) {
				this.goToPage(this.failedRoute);
				this.failedRoute = null;
			} else {
				this.goToPage({name: this.routeEntryPage});
			}
		},
		// Used to break login flow (when aal2 or higher is required) and go to a different page
		async logoutAndGoToPage(pageName) {
			try {
				const response = await this.ory.frontend.createBrowserLogoutFlow();
				if (!response.data.logout_token) {
					return;
				}

				const newResponse = await this.ory.frontend.updateLogoutFlow({token: response.data.logout_token});

				if (newResponse.status === 204) {
					this.session = null;
					this.goToPage({name: pageName});
				}
			} catch (e) {
				// Could not log out because no session was found -> go to requested page
				if (e.response?.data?.error?.id === 'session_inactive') {
					this.goToPage({name: pageName});
				} else {
					await handleGetFlowError(this, e, null);
				}
			}
		},
		async handleOrySubmitLogin(formID) {
			const form = document.getElementById(formID);
			if (!form || !this.loginFlow.ui.action) {
				return;
			}

			// Disable submitting from this form
			this.disabledForms.push(formID);

			const body = Object.fromEntries(new FormData(form));
			const {flow} = this.$route.query;

			try {
				const response = await this.ory.frontend.updateLoginFlow({flow, updateLoginFlowBody: body});

				if (response.status === 200 && response.data?.session) {
					// Reminder: check when https://github.com/ory/kratos/pull/3572 is released
					if (response.data.session.identity) {
						this.session = response.data.session;
						this.leave();
						return;
					}

					// Aal2 not done yet
					await this.initLoginFlow('aal2');

					return;
				}

				// Something went wrong and we need to display some data
				if (response.data && response.data.ui) {
					this.setFlowData(response.data);
				}

				if (response.error && response.error.reason) {
					this.setErrorMessage(response.error.reason);
				}
			} catch (e) {
				if (e.response?.data?.ui) {
					this.setFlowData(e.response.data);
				} else {
					handleGetFlowError(this, e, () => {
						this.initLoginFlow('aal1');
						this.setErrorMessage('The login flow has expired, please try again.');
					}).catch(e => {
						this.setErrorMessage(e);
					});
				}
			} finally {
				// Enable submitting for this form again
				this.disabledForms = this.disabledForms.filter(d => d !== formID);
			}
		},
		async initFlow() {
			const {flow} = this.$route.query;

			if (typeof flow === 'string') {
				try {
					const response = await this.ory.frontend.getLoginFlow({id: flow});
					this.setFlowData(response.data);
				} catch (e) {
					await handleGetFlowError(this, e, () => this.initLoginFlow('aal1'));
				}
			} else {
				// If there's no flow in our route,
				// we need to initialize our login flow
				await this.initLoginFlow('aal1');
			}
		},
		async initLoginFlow(aal) {
			// Set refresh to true, for the case when the local
			// session data was deleted but the user is still logged in

			try {
				const response = await this.ory.frontend.createBrowserLoginFlow({refresh: false, aal});
				this.setFlowData(response.data);
			} catch (e) {
				if (e.response?.data?.error?.id === 'session_already_available') {
					// Reminder: check when https://github.com/ory/kratos/pull/3572 is released
					// If the response indicates that the session is already available,
					// it might be only aal1 even though aal2 might be required.
					// This can be checked by requesting the session. If it fails the aal2 dialog will be rendered
					await this.tryToGetSession();
					return;
				}

				await handleGetFlowError(this, e, null);
			}
		},
		setFlowData(d) {
			this.loginFlow = d;
			this.showLogoutButton = Boolean(d.requested_aal) && d.requested_aal !== 'aal1';
			if (!this.$route.query.flow || this.$route.query.flow !== d.id) {
				this.$router.replace({query: {flow: d.id}});
			}
		},
	},
};
</script>

<style scoped>

</style>
