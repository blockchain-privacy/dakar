<template>
  <v-row align="center" no-gutters class="fill-height">
    <v-col cols="12" md="6" class="hidden-md-and-down fill-height">
      <v-sheet color="primary darken-2" dark height="100%" width="100%">
        <v-container class="justify-center fill-height">
          <div class="d-flex align-center flex-column text-center">
            <h1 class="text-xl-h1 text-md-h2 font-weight-bold">
              {{ applicationName }}
            </h1>
            <h3 class="text-xl-h3 text-md-h4 mt-4">
              {{ applicationSubtitle }}
            </h3>
          </div>
        </v-container>
      </v-sheet>
    </v-col>
    <v-col cols="12" lg="6">
      <v-container>
        <v-row justify="center">
          <v-col cols="12" lg="8" md="8" xl="5">
            <v-card class="elevation-4">
              <div class="pa-5 pb-8">
                <h3 class="text-h3 font-weight-bold text-center">
                  Welcome!
                </h3>
                <ory-flow v-if="loginFlow"
                          :flow="loginFlow"
                          :form-id="formID"
                          @submit="handleOrySubmitLogin"/>
                <v-skeleton-loader
                    v-else
                    class="mx-auto"
                    type="article, actions"/>
                <router-link  class="float-right" :to="{name: this.routeAccountRecovery}">
                  Recover account
                </router-link>
              </div>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-col>
  </v-row>
</template>

<script>
import {
  mdiLockOutline, mdiEye, mdiEyeOff, mdiEmail,
} from '@mdi/js';
import {
  APPLICATION_NAME, PAGE_TITLE, PASSWORD_MIN_CHARACTERS, ROUTE_NAME_ENTRY_PAGE,
  PASSWORD_MAX_CHARACTERS, APPLICATION_SUBTITLE, ROUTE_NAME_ACCOUNT_RECOVERY,
} from '../../constants';
import { emailRules, passwordRules } from '../../utilities';
import handleGetFlowError from '../../kratos';
import OryFlow from './flows/OryFlow.vue';

function goToPage(context, pageObj) {
  context.$router.push(pageObj);
}

function goToRoot(context) {
  goToPage(context, { name: ROUTE_NAME_ENTRY_PAGE });
}

export default {
  name: 'Login',
  components: { OryFlow },
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
      rules: { passwordRules, emailRules },
      email: {
        value: '',
      },
      password: {
        value: '',
        show: false,
      },
      loginFlow: null,
      formID: 'login-form',
      orySession: null,
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
  methods: {
    setErrorMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'error', temporary: true });
    },
    leave() {
      if (this.loginFlow && this.loginFlow.return_to) {
        window.location.href = this.loginFlow.return_to;
      } else if (this.failedRoute) {
        goToPage(this, this.failedRoute);
        this.failedRoute = null;
      } else goToRoot(this);
    },
    handleOrySubmitLogin(formID) {
      const form = document.getElementById(formID);
      if (!form || !this.loginFlow.ui.action) return;

      const body = Object.fromEntries(new FormData(form));
      const { flow } = this.$route.query;
      this.ory.submitSelfServiceLoginFlow(flow, JSON.stringify(body))
        .then((response) => {
          if (response.status === 200 && response.data && response.data.session) {
            if (!response.data.session.identity) {
              // aal2 not done yet
              this.initLoginFlow('aal2');
              return;
            }

            this.session = response.data.session;
            this.leave();
            return;
          }

          // something went wrong and we need to display some data
          if (response.data && response.data.ui) {
            this.setFlowData(response.data);
          }

          if (response.error && response.error.reason) this.setErrorMessage(response.error.reason);
        })
        .catch((err) => {
          if (err.response && err.response.data && err.response.data.ui) {
            this.setFlowData(err.response.data);
          } else {
            handleGetFlowError(this.$router, this.$store, err).catch((e) => {
              this.setErrorMessage(e);
            });
          }
        });
    },
    initFlow() {
      const { flow } = this.$route.query;

      if (typeof flow !== 'string') {
        // if there's no flow in our route,
        // we need to initialize our login flow
        this.initLoginFlow('aal1');
      } else {
        this.ory.getSelfServiceLoginFlow(flow)
          .then((d) => this.setFlowData(d.data))
          .catch((err) => {
            handleGetFlowError(this.$router, this.$store, err);
          });
      }
    },
    initLoginFlow(aal) {
      // set refresh to true, for the case when the local
      // session data was deleted but the user is still logged in
      this.ory.initializeSelfServiceLoginFlowForBrowsers(false, aal)
        .then((d) => this.setFlowData(d.data))
        .catch((err) => {
          handleGetFlowError(this.$router, this.$store, err);
        });
    },
    setFlowData(d) {
      this.loginFlow = d;
      if (!this.$route.query.flow || this.$route.query.flow !== d.id) {
        this.$router.replace({ query: { flow: d.id } });
      }
    },
  },
  mounted() {
    document.title = `Login - ${PAGE_TITLE}`;

    // if session is not set, user might be logged in already -> get session
    if (!this.session) {
      this.ory.toSession()
        .then((d) => {
          if (d.status && d.status === 200) {
            this.session = d.data;
            this.leave();
          }
        })
        .catch((error) => {
          if (error.response && error.response.data && error.response.data.error
              && error.response.data.error.id && error.response.data.error.id === 'session_aal2_required') {
            this.initLoginFlow('aal2');
            return;
          }

          // this request fails if the user is not logged in -> init login form
          this.initFlow();
        });
    } else {
      this.leave();
    }
  },
  watch: {
    $route(to) {
      if (!to.query.flow) {
        // this happens if the users manually navigates to the route of this page,
        // in this case flow is not set and needs to be reinitialized
        this.initFlow();
      }
    },
  },
};
</script>

<style scoped>

</style>
