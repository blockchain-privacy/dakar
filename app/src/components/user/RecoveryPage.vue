<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

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
        <h3 class="text-display-medium font-weight-bold text-center ma-0">
          Account Recovery
        </h3>
        <alert :text="errorMsg" />
        <ory-flow
          v-if="recoveryFlow"
          class="mt-3"
          :flow="recoveryFlow"
          form-id="recovery-form"
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

<script setup>
import {inject, onMounted, ref} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import OryFlow from './ory/OryFlow.vue';
import {PAGE_TITLE} from '@/constants';
import handleGetFlowError from '@/kratos';
import {useLocalStore} from '@/pinia/local';
import {useNavStore} from '@/pinia/nav';
import Alert from '@/components/common/Alert.vue';

const ory = inject('ory');
const route = useRoute();
const router = useRouter();
const localStore = useLocalStore();
const navStore = useNavStore();
const context = {
	$route: route, $router: router, navStore, localStore,
};

const recoveryFlow = ref(null);
const disabledForms = ref([]);
const errorMsg = ref('');

// Hooks
onMounted(async () => {
	document.title = `Account Recovery - ${PAGE_TITLE}`;

	const {flow} = route.query;
	errorMsg.value = '';

	if (typeof flow === 'string') {
		try {
			const response = await ory.frontend.getRecoveryFlow({id: flow});
			setFlowData(response);
		} catch (error) {
			await doErrorHandling(context, error, initRecoveryFlow);
		}
	} else {
		// If there's no flow in our route,
		// we need to initialize our login flow
		await initRecoveryFlow();
	}
});

// Functions
async function doErrorHandling(ctx, err, onRefreshFlow, isOAuth) {
	try {
		const msg = await handleGetFlowError(ctx, err, onRefreshFlow, isOAuth);
		if (msg) {
			errorMsg.value = msg;
		}
	} catch (error) {
		errorMsg.value = error.message;
	}
}

async function initRecoveryFlow() {
	errorMsg.value = '';
	try {
		const response = await ory.frontend.createBrowserRecoveryFlow();
		setFlowData(response);
	} catch (error) {
		await doErrorHandling(context, error, null);
	}
}

function setFlowData(d) {
	recoveryFlow.value = d;
	if (!route.query.flow || route.query.flow !== d.id) {
		router.replace({query: {flow: d.id}});
	}
}

async function handleOrySubmitRecovery(formID, btnName) {
	const form = document.getElementById(formID);
	if (!form || !recoveryFlow.value.ui.action) {
		return;
	}

	errorMsg.value = '';

	// Disable submitting from this form
	disabledForms.value.push(formID);

	const body = Object.fromEntries(new FormData(form));
	const {flow} = route.query;

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
			disabledForms.value = disabledForms.value.filter(d => d !== formID);
			// Nothing to submit -> just return
			return;
		}
	}

	try {
		const response = await ory.frontend.updateRecoveryFlow({flow, updateRecoveryFlowBody: body});
		if (response?.ui) {
			setFlowData(response);
		}

		if (response.error?.reason) {
			errorMsg.value = response.error.reason;
		}
	} catch (error) {
		if (error.response?.ui) {
			setFlowData(error.response);
		} else {
			await doErrorHandling(context, error, initRecoveryFlow);
		}
	}

	// Enable submitting for this form again
	disabledForms.value = disabledForms.value.filter(d => d !== formID);
}

</script>

<style scoped>

</style>
