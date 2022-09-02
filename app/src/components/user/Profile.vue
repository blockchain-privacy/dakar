<template>
  <v-card class="mx-auto elevation-4" max-width="700">
    <v-toolbar color="primary" dark flat>
      <v-toolbar-title>
        <v-icon>{{ icons.mdiAccountDetails }}</v-icon>
        Profile
      </v-toolbar-title>
    </v-toolbar>
    <v-card>
      <v-card-text>
        <ory-flow v-if="settingsFlow" class="mt-4"
                  :flow="settingsFlow"
                  :form-id="formID"
                  @submit="handleOrySubmitSettings"/>
      </v-card-text>
    </v-card>
  </v-card>
</template>

<script>
import { mdiAccountDetails } from '@mdi/js';
import { PAGE_TITLE } from '../../constants';
import OryFlow from './flows/OryFlow.vue';
import handleGetFlowError from '../../kratos';

export default {
  name: 'Profile',
  components: { OryFlow },
  data() {
    return {
      icons: { mdiAccountDetails },
      formID: 'settings-form',
      settingsFlow: null,
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
    initSettingsFlow() {
      this.ory.initializeSelfServiceSettingsFlowForBrowsers()
        .then((d) => this.setFlowData(d.data))
        .catch((err) => {
          if (err.ui) this.setFlowData(err);
          else handleGetFlowError(this.$router, this.$store, err);
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

      const body = Object.fromEntries(new FormData(form));
      const { flow } = this.$route.query;
      this.ory.submitSelfServiceSettingsFlow(flow, JSON.stringify(body))
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
        });
    },
    refreshSession() {
      this.ory.toSession()
        .then((d) => {
          if (d.status && d.status === 200) {
            this.session = d.data;
          }
        })
        .catch();
    },
    initFlow() {
      const { flow } = this.$route.query;

      if (typeof flow !== 'string') {
        // if there's no flow in our route,
        // we need to initialize our login flow
        this.initSettingsFlow();
      } else {
        this.ory.getSelfServiceSettingsFlow(flow)
          .then((d) => {
            this.setFlowData(d.data);
            // if an account is being recovered the session is empty,
            // therefore it has to be refreshed.
            this.refreshSession();
          })
          .catch((err) => {
            if (err.ui) this.setFlowData(err);
            else handleGetFlowError(this.$router, this.$store, err);
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
