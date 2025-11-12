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
        <div class="d-flex">
          <v-img
            alt="Dakar Logo"
            :src="DakarImg"
            class="mb-4"
            transition="fade-transition"
            width="64"
            max-height="75px"
          />
        </div>
        <h3 class="text-h3 font-weight-bold text-center mb-2">
          Consent Request
        </h3>
      </div>
      <v-card-text>
        <error-message
          v-if="errorMsg"
          :text="errorMsg"
        />
        <div
          v-else-if="consentRejected"
          class="text-subtitle-1"
        >
          Consent rejected. You can close this page now.
        </div>
        <template v-else-if="consentData.show">
          <div class="text-subtitle-1">
            Hi <code v-if="consentData.userIdentifier">{{ consentData.userIdentifier }}</code>,
            <strong>{{ consentData.client.name || consentData.client.id }}</strong> wants access resources on your behalf with the following permissions:
          </div>
          <v-list>
            <v-list-item
              v-for="item in consentData.requestedScope"
              :key="item"
              :title="item.title"
              :subtitle="item.description"
            />
          </v-list>
        </template>
        <template v-else>
          <h5 class="text-h5 font-weight-bold text-center my-2">
            Checking consent
          </h5>
          <v-progress-linear indeterminate />
        </template>
      </v-card-text>
      <v-card-actions v-if="consentData.show && !consentRejected">
        <v-btn
          class="ms-auto"
          :loading="isLoading"
          @click="requestConsent(false)"
        >
          Reject
        </v-btn>
        <v-btn
          color="success"
          :loading="isLoading"
          @click="requestConsent(true)"
        >
          Accept
        </v-btn>
      </v-card-actions>
    </v-card>
  </div>
</template>

<script setup>
import {PAGE_TITLE, ROUTE_NAME_OAUTH_ERROR_PAGE} from '@/constants';
import {
	inject, onMounted, ref,
} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import ErrorMessage from '@/components/common/ErrorMessage.vue';
import DakarImg from '@/assets/dakar.svg?url';

const route = useRoute();
const router = useRouter();
const kratosAdmin = inject('kratosadmin');

const isLoading = ref(false);
const errorMsg = ref('');
const consentData = ref({
	show: false,
	client: null,
	requestedScope: [],
	userIdentifier: null,
});
const consentRejected = ref(false);

// Hooks
onMounted(async () => {
	document.title = `OAuth Consent - ${PAGE_TITLE}`;

	await requestConsent();
});

// Functions

// Maps an array of scope strings to an array of objects containing the scope and its description.
function addScopeDescriptions(scopes) {
	if (!scopes) {
		return [];
	}

	return scopes.map(scope => {
		switch (scope) {
			case 'openid':
				return {
					title: scope,
					description: 'Allows the requesting application to verify your identity using your account data.',
				};
			case 'offline':
				return {
					title: scope,
					description: 'Allows the requesting application refresh the authentication session.',
				};
			default:
				return {title: scope};
		}
	});
}

// Accepted: undefined -> check if consent request can be skipped
// accepted: true -> user accepted consent request, complete the consent flow
// accepted: false -> user rejected consent request, abort the consent flow
async function requestConsent(accepted) {
	const consentChallenge = route.query.consent_challenge;
	if (!consentChallenge) {
		router.push({name: ROUTE_NAME_OAUTH_ERROR_PAGE});
		return;
	}

	isLoading.value = true;

	try {
		const r = await kratosAdmin.oauth.consentPost({consent: {challenge: consentChallenge, accepted}});

		if (r.redirectTo) {
			if (accepted === false) {
				fetch(r.redirectTo);
				consentRejected.value = true;
				return;
			}

			window.location.href = r.redirectTo;
			return;
		}

		if (r.client) {
			consentData.value.client = r.client;
		}

		if (r.requestedScope) {
			consentData.value.requestedScope = addScopeDescriptions(r.requestedScope);
		}

		if (r.userIdentifier) {
			consentData.value.userIdentifier = r.userIdentifier;
		}

		if (consentData.value.client || consentData.value.requestedScope.length > 0 || consentData.value.userIdentifier) {
			consentData.value.show = true;
		}
	} catch (e) {
		errorMsg.value = e.message;
	} finally {
		isLoading.value = false;
	}
}

</script>

<style scoped>

</style>
