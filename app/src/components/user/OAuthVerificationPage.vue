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
        <div class="text-display-medium font-weight-bold text-center ma-0">
          Verification
        </div>
        <div class="text-body-large text-center mb-2">
          Input the verification code
        </div>
        <v-form>
          <v-alert
            v-if="errorText"
            type="error"
            class="mt-2"
            variant="text"
          >
            {{ errorText }}
          </v-alert>
          <v-text-field
            v-model="code"
            label="Code"
            :prepend-inner-icon="mdiFormTextboxPassword"
            min-width="300px"
            :disabled="isLoading"
          />
          <v-btn
            block
            flat
            :loading="isLoading"
            @click="handleClick"
          >
            Submit
          </v-btn>
        </v-form>
      </v-card-text>
    </v-card>
  </div>
</template>

<script setup>
import {inject,	onMounted, ref} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {mdiFormTextboxPassword} from '@mdi/js';
import {PAGE_TITLE, ROUTE_NAME_ENTRY_PAGE} from '@/constants';
import DakarImg from '@/assets/dakar.svg?url';

const kratosAdmin = inject('kratosadmin');

const router = useRouter();
const route = useRoute();
const code = ref('');
const challenge = ref('');
const errorText = ref('');
const isLoading = ref(false);

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

	isLoading.value = true;
	errorText.value = '';

	try {
		const r = await kratosAdmin.oauth.verifyPost({challenge: {challenge: challenge.value, code: code.value.trim()}});
		if (r.redirectTo) {
			window.location.href = r.redirectTo;
		}
	} catch (error) {
		if (error.cause?.status === 400) {
			errorText.value = 'Please try again. Codes are case-sensitive. If the error continues, restart the authentication process.';
			return;
		}

		errorText.value = error.message;
	} finally {
		isLoading.value = false;
	}
}

</script>

<style scoped>

</style>
