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
          Account Recovery
        </h3>
        <ory-flow
          v-if="recoveryFlow"
          class="mt-3"
          :flow="recoveryFlow"
          :form-id="formID"
          :disabled-forms="disabledForms"
          @submit="handleOrySubmitRecovery"
        />
        <v-skeleton-loader
          v-else
          class="mx-auto"
          type="article, actions"
        />
      </div>
    </v-card>
  </div>
</template>

<script>
import {APPLICATION_NAME, PAGE_TITLE} from '@/constants';
import handleGetFlowError from '@/kratos';
import OryFlow from './ory/OryFlow.vue';
export default {
	name: 'RecoveryPage',
	components: {OryFlow},
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
	async mounted() {
		document.title = `Account Recovery - ${PAGE_TITLE}`;

		const {flow} = this.$route.query;

		if (typeof flow === 'string') {
			try {
				const response = await this.ory.frontend.getRecoveryFlow({id: flow});
				this.setFlowData(response.data);
			} catch (e) {
				await handleGetFlowError(this, e, this.initRecoveryFlow);
			}
		} else {
			// If there's no flow in our route,
			// we need to initialize our login flow
			await this.initRecoveryFlow();
		}
	},
	methods: {
		setErrorMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'error', temporary: true, category: this.$route.name});
		},
		async initRecoveryFlow() {
			try {
				const response = await 	this.ory.frontend.createBrowserRecoveryFlow();
				this.setFlowData(response.data);
			} catch (e) {
				await handleGetFlowError(this, e, null);
			}
		},
		setFlowData(d) {
			this.recoveryFlow = d;
			if (!this.$route.query.flow || this.$route.query.flow !== d.id) {
				this.$router.replace({query: {flow: d.id}});
			}
		},
		async handleOrySubmitRecovery(formID, btnName) {
			const form = document.getElementById(formID);
			if (!form || !this.recoveryFlow.ui.action) {
				return;
			}

			// Disable submitting from this form
			this.disabledForms.push(formID);

			const body = Object.fromEntries(new FormData(form));
			const {flow} = this.$route.query;

			// The recovery form has two submit buttons:
			// - submit code (button id: method)
			// - resend code (button id: email)
			if (btnName === 'method' && body.code !== undefined) {
				const c = body.code.trim();
				if (c.length > 0) {
					body.code = c;
					delete body.email;
				} else {
					// Enable submitting for this form again
					this.disabledForms = this.disabledForms.filter(d => d !== formID);
					// Nothing to submit -> just return
					return;
				}
			}

			try {
				const response = await this.ory.frontend.updateRecoveryFlow({flow, updateRecoveryFlowBody: body});
				if (response.data?.ui) {
					this.setFlowData(response.data);
				}

				if (response.error?.reason) {
					this.setErrorMessage(response.error.reason);
				}
			} catch (e) {
				if (e.response?.data?.ui) {
					this.setFlowData(e.response.data);
				} else {
					try {
						await handleGetFlowError(this, e, this.initRecoveryFlow);
					} catch (e) {
						this.setErrorMessage(e);
					}
				}
			}

			// Enable submitting for this form again
			this.disabledForms = this.disabledForms.filter(d => d !== formID);
		},
	},
};
</script>

<style scoped>

</style>
