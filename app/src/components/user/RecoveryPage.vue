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
        <h3 class="text-h3 font-weight-bold text-center mb-2">
          Account Recovery
        </h3>
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
import {PAGE_TITLE} from '@/constants';
import handleGetFlowError from '@/kratos';
import OryFlow from './ory/OryFlow.vue';
import {inject, onMounted, ref} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {useLocalStore} from '@/pinia/local';
import {useNavStore} from '@/pinia/nav';
import {useMsgStore} from '@/pinia/msg';

const ory = inject('ory');
const route = useRoute();
const router = useRouter();
const localStore = useLocalStore();
const navStore = useNavStore();
const msgStore = useMsgStore();
const context = {
	$route: route, $router: router, navStore, localStore, msgStore,
};

const recoveryFlow = ref(null);
const disabledForms = ref([]);

// Hooks
onMounted(async () => {
	document.title = `Account Recovery - ${PAGE_TITLE}`;

	const {flow} = route.query;

	if (typeof flow === 'string') {
		try {
			const response = await ory.frontend.getRecoveryFlow({id: flow});
			setFlowData(response);
		} catch (e) {
			await handleGetFlowError(context, e, initRecoveryFlow);
		}
	} else {
		// If there's no flow in our route,
		// we need to initialize our login flow
		await initRecoveryFlow();
	}
});

// Functions
function setErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: true, category: route.name,
	});
}

async function initRecoveryFlow() {
	try {
		const response = await ory.frontend.createBrowserRecoveryFlow();
		setFlowData(response);
	} catch (e) {
		await handleGetFlowError(context, e, null);
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
			setErrorMessage(response.error.reason);
		}
	} catch (e) {
		if (e.response?.ui) {
			setFlowData(e.response);
		} else {
			try {
				await handleGetFlowError(context, e, initRecoveryFlow);
			} catch (e) {
				setErrorMessage(e);
			}
		}
	}

	// Enable submitting for this form again
	disabledForms.value = disabledForms.value.filter(d => d !== formID);
}

</script>

<style scoped>

</style>
