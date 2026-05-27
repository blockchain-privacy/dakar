<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-app>
    <app-bar
      v-if="!isOAuthPage"
      :minimize="isEntryPage"
    />
    <v-main>
      <router-view v-slot="{ Component }">
        <fade-transition>
          <component :is="Component" />
        </fade-transition>
      </router-view>
    </v-main>
  </v-app>
</template>

<script setup>
// eslint-disable-next-line import-x/no-unassigned-import
import '@fontsource/roboto';
import {computed, onBeforeMount} from 'vue';
import {useRoute} from 'vue-router';
import {useTheme} from 'vuetify';
import {ROUTE_NAME_ENTRY_PAGE} from './constants';
import AppBar from './components/appbar/AppBar.vue';
import FadeTransition from '@/components/common/FadeTransition.vue';
import {useLocalStore} from '@/pinia/local';

const route = useRoute();
const theme = useTheme();
const localStore = useLocalStore();

// Computed
const isEntryPage = computed(() => route.name === ROUTE_NAME_ENTRY_PAGE);

const isOAuthPage = computed(() => route.path.startsWith('/oauth/'));

// Hooks
onBeforeMount(() => {
	checkSessionExpiration();
	theme.change(localStore.getSettings.theme);
});

// Functions
// CheckSessionExpiration removes the stored session if it expired
function checkSessionExpiration() {
	if (!localStore.getSession?.expires_at) {
		return;
	}

	const expiryDate = new Date(localStore.getSession.expires_at);
	if (Number.isNaN(expiryDate)) {
		return;
	}

	if (new Date() > expiryDate) {
		localStore.deleteSession();
	}
}

</script>

<style scoped>

</style>
