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
          Verification
        </h3>
        <h6 class="text-h6 text-center mb-2">
          Please input the verification code
        </h6>
        <v-form>
          <v-alert
            v-if="isError"
            type="error"
            class="mt-2"
            variant="text"
          >
            Please try again. Codes are case-sensitive. If the error continues, restart the authentication process.
          </v-alert>
          <v-text-field
            v-model="code"
            label="Code"
            :prepend-inner-icon="mdiFormTextboxPassword"
            min-width="300px"
          />
          <v-btn
            block
            @click="handleClick"
          >
            Submit
          </v-btn>
        </v-form>
      </div>
    </v-card>
  </div>
</template>

<script setup>
import {PAGE_TITLE, ROUTE_NAME_ENTRY_PAGE} from '@/constants';
import {inject,	onMounted, ref} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {mdiFormTextboxPassword} from '@mdi/js';
import DakarImg from '@/assets/dakar.svg?url';
const kratosAdmin = inject('kratosadmin');

const router = useRouter();
const route = useRoute();
const code = ref('');
const challenge = ref('');
const isError = ref(false);

// Hooks
onMounted(async () => {
	document.title = `OAuth Verification - ${PAGE_TITLE}`;
	challenge.value = route.query.device_challenge;

	if (!challenge.value) {
		router.push({name: ROUTE_NAME_ENTRY_PAGE});
	}
});

// Functions

async function handleClick() {
	if (!challenge.value || !code.value) {
		return;
	}

	isError.value = false;

	try {
		const r = await kratosAdmin.oauth.verifyPost({challenge: {challenge: challenge.value, code: code.value.trim()}});
		if (r.redirectTo) {
			window.location.href = r.redirectTo;
		}
	} catch (_) {
		isError.value = true;
	}
}

</script>

<style scoped>

</style>
