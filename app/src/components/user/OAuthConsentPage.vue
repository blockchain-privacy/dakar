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
      <v-card-text>
        <h5 class="text-h5 font-weight-bold text-center my-2">
          Checking consent
        </h5>
        <v-progress-linear indeterminate />
      </v-card-text>
    </v-card>
  </div>
</template>

<script setup>
import {PAGE_TITLE, ROUTE_NAME_ENTRY_PAGE, ROUTE_NAME_OAUTH_ERROR_PAGE} from '@/constants';
import {
	inject, onMounted,
} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const route = useRoute();
const router = useRouter();
const msgStore = useMsgStore();
const kratosAdmin = inject('kratosadmin');

// Hooks
onMounted(async () => {
	document.title = `OAuth Consent - ${PAGE_TITLE}`;

	const consentChallenge = route.query.consent_challenge;
	if (consentChallenge) {
		try {
			const r = await kratosAdmin.oauth.consentConsentChallengePost({consentChallenge});

			if (r.redirectTo) {
				window.location.href = r.redirectTo;
				return;
			}
		} catch (e) {
			setErrorMessage(e);
			router.push({name: ROUTE_NAME_OAUTH_ERROR_PAGE});
			return;
		}
	}

	router.push({name: ROUTE_NAME_ENTRY_PAGE});
});

// Functions
function setErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: true, category: route.name,
	});
}

</script>

<style scoped>

</style>
