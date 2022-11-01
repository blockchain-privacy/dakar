<template>
  <v-row align="center" no-gutters class="fill-height">
    <v-col cols="12">
      <v-container>
        <v-row justify="center">
          <v-col cols="12" lg="8" md="8" xl="5">
            <v-card class="elevation-4">
              <div class="pa-5">
                <h3 class="text-h3 font-weight-bold text-center">
                  Account Recovery
                </h3>
                <ory-flow v-if="recoveryFlow" class="mt-4"
                          :flow="recoveryFlow"
                          :form-id="formID"
                          :disabled-forms="disabledForms"
                          @submit="handleOrySubmitRecovery"/>
                <v-skeleton-loader
                    v-else
                    class="mx-auto"
                    type="article, actions"/>
              </div>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-col>
  </v-row>
</template>

<script>
import { APPLICATION_NAME, PAGE_TITLE } from '../../constants';
import handleGetFlowError from '../../kratos';
import OryFlow from './flows/OryFlow.vue';

export default {
  name: 'Recovery',
  components: { OryFlow },
  data() {
    return {
      isSubmittingForm: false,
      loginFailed: false,
      applicationName: APPLICATION_NAME,
      recoveryFlow: null,
      formID: 'recovery-form',
      disabledForms: [],
    };
  },
  methods: {
    setErrorMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'error', temporary: true });
    },
    initRecoveryFlow() {
      this.ory.initializeSelfServiceRecoveryFlowForBrowsers()
        .then((d) => this.setFlowData(d.data))
        .catch((err) => {
          handleGetFlowError(this.$router, this.$store, err);
        });
    },
    setFlowData(d) {
      this.recoveryFlow = d;
      if (!this.$route.query.flow || this.$route.query.flow !== d.id) {
        this.$router.replace({ query: { flow: d.id } });
      }
    },
    handleOrySubmitRecovery(formID) {
      const form = document.getElementById(formID);
      if (!form || !this.recoveryFlow.ui.action) return;

      // disable submitting from this form
      this.disabledForms.push(formID);

      const body = Object.fromEntries(new FormData(form));
      const { flow } = this.$route.query;
      this.ory.submitSelfServiceRecoveryFlow(flow, JSON.stringify(body))
        .then((response) => {
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
        })
        .finally(() => {
          // enable submitting for this form again
          this.disabledForms = this.disabledForms.filter((d) => d !== formID);
        });
    },
  },
  mounted() {
    document.title = `Account Recovery - ${PAGE_TITLE}`;

    const { flow } = this.$route.query;

    if (typeof flow !== 'string') {
      // if there's no flow in our route,
      // we need to initialize our login flow
      this.initRecoveryFlow();
    } else {
      this.ory.getSelfServiceRecoveryFlow(flow)
        .then((d) => this.setFlowData(d.data))
        .catch((err) => {
          handleGetFlowError(this.$router, this.$store, err);
        });
    }
  },
};
</script>

<style scoped>

</style>
