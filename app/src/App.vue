<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-app>
    <app-bar :minimize="isEntryPage" />
    <v-main>
      <div style="position: relative">
        <msg-box />
      </div>
      <router-view v-slot="{ Component }">
        <fade-transition>
          <component :is="Component" />
        </fade-transition>
      </router-view>
    </v-main>
  </v-app>
</template>

<script setup>
import MsgBox from './components/notification/MsgBox.vue';
import '@fontsource/roboto';
import {ROUTE_NAME_ENTRY_PAGE} from './constants';
import AppBar from './components/appbar/AppBar.vue';
import FadeTransition from '@/components/common/FadeTransition.vue';
import {computed, onBeforeMount} from 'vue';
import {useRoute} from 'vue-router';
import {useTheme} from 'vuetify';
import {useLocalStore} from '@/pinia/local';

const route = useRoute();
const theme = useTheme();
const localStore = useLocalStore();

// Computed
const isEntryPage = computed(() => route.name === ROUTE_NAME_ENTRY_PAGE);

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
	if (isNaN(expiryDate)) {
		return;
	}

	if (new Date() > expiryDate) {
		localStore.deleteSession();
	}
}

</script>

<style scoped>

</style>
