<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-progress-linear indeterminate />
</template>

<script setup>
import {PAGE_TITLE, ROUTE_NAME_ENTRY_PAGE} from '@/constants';
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
	document.title = `OAuth Logout - ${PAGE_TITLE}`;

	const logoutChallenge = route.query.logout_challenge;
	if (logoutChallenge) {
		try {
			console.log('logging out');
			const r = await kratosAdmin.oauth.logoutLogoutChallengePost({logoutChallenge});

			if (r.redirectTo) {
				window.location.href = r.redirectTo;
				return;
			}
		} catch (e) {
			setErrorMessage(e);
			router.push({name: ROUTE_NAME_ENTRY_PAGE});
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
