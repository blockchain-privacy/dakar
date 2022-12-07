<template>
  <div v-if="settingsFlow">
    <ory-flow class="mt-4"
              :flow="settingsFlow"
              :form-id="formID"
              @submit="handleOrySubmitSettings"
              :disabled-forms="disabledForms"
              embed
    />
    <v-card class="mx-auto elevation-4 my-5" max-width="700">
      <v-toolbar color="red" dark flat>
        <v-toolbar-title>
          Delete Account
        </v-toolbar-title>
      </v-toolbar>
      <v-card-text>
        To delete your account and all associated data click below. This action can not be reversed.
        <v-btn class="mt-2" color="red" dark block
               @click="showAccountDeletionDialog=true">Delete</v-btn>
      </v-card-text>
    </v-card>
    <v-dialog
        v-model="showAccountDeletionDialog"
        max-width="700px">
      <v-card>
        <v-card-title>Delete Account</v-card-title>
        <v-card-text>
          <p class="font-weight-black text-body-1 my-0">
            Do you really want to delete your account? This action can not be reversed.</p>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text
                 @click="showAccountDeletionDialog = false">Cancel
          </v-btn>
          <v-btn color="red darken-1"
                 text
                 @click="deleteIdentity()">Yes, delete my account
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
  <v-skeleton-loader v-else class="mx-auto" type="article, actions"/>
</template>

<script>
import { mdiAccountDetails } from '@mdi/js';
import {
  PAGE_TITLE,
  ROUTE_IDENTITY_DELETE,
  ROUTE_NAME_ENTRY_PAGE,
} from '../../constants';
import OryFlow from './flows/OryFlow.vue';
import handleGetFlowError from '../../kratos';
import { doGet, handleError } from '../../utilities';

export default {
  name: 'Profile',
  components: { OryFlow },
  data() {
    return {
      icons: { mdiAccountDetails },
      formID: 'settings-form',
      settingsFlow: null,
      disabledForms: [],
      showAccountDeletionDialog: false,
      route: {
        rootPage: ROUTE_NAME_ENTRY_PAGE,
      },
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
  methods: {
    setSuccessMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'success', temporary: true });
    },
    setErrorMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'error', temporary: true });
    },
    deleteIdentity() {
      return doGet(ROUTE_IDENTITY_DELETE, this.$router, this.$store).then((data) => {
        if (!data.success) throw Error('error deleting account');
        this.$store.dispatch('resetMessages');
        this.setSuccessMessage('Your account was successfully deleted. Goodbye!');
        this.session = null;
        this.$router.push({ name: this.route.rootPage });
      }).catch((e) => {
        handleError(this.$store, e);
        return e;
      }).finally(() => {
        this.showAccountDeletionDialog = false;
      });
    },
    initSettingsFlow() {
      this.ory.frontend.createBrowserSettingsFlow()
        .then((d) => this.setFlowData(d.data))
        .catch((err) => {
          handleGetFlowError(this.$router, this.$store, err);
        });
    },
    setFlowData(d) {
      this.settingsFlow = d;
      if (!this.$route.query.flow || this.$route.query.flow !== d.id) {
        this.$router.replace({ query: { flow: d.id } });
      }
    },
    handleOrySubmitSettings(formID) {
      const form = document.getElementById(formID);
      if (!form || !this.settingsFlow.ui.action) return;

      // disable submitting from this form
      this.disabledForms.push(formID);

      const body = Object.fromEntries(new FormData(form));
      const { flow } = this.$route.query;

      this.ory.frontend.updateSettingsFlow({ flow, updateSettingsFlowBody: body })
        .then((response) => {
          // something went wrong and we need to display some data
          if (response.data && response.data.ui) {
            this.setFlowData(response.data);
          }

          // if an account is being recovered the session is empty,
          // therefore it has to be refreshed.
          this.refreshSession();

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
        })
        .finally(() => {
          // enable submitting for this form again
          this.disabledForms = this.disabledForms.filter((d) => d !== formID);
        });
    },
    refreshSession() {
      this.ory.frontend.toSession()
        .then((d) => {
          if (d.status && d.status === 200) {
            this.session = d.data;
          }
        })
        .catch((err) => {
          handleGetFlowError(this.$router, this.$store, err);
        });
    },
    initFlow() {
      const { flow } = this.$route.query;

      if (typeof flow !== 'string') {
        // if there's no flow in our route,
        // we need to initialize our login flow
        this.initSettingsFlow();
      } else {
        this.ory.frontend.getSettingsFlow({ id: flow })
          .then((d) => {
            this.setFlowData(d.data);
            // if an account is being recovered the session is empty,
            // therefore it has to be refreshed.
            this.refreshSession();
          })
          .catch((err) => {
            handleGetFlowError(this.$router, this.$store, err);
          });
      }
    },
  },
  mounted() {
    document.title = `Profile - ${PAGE_TITLE}`;
    this.initFlow();
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
