<template>
  <v-app>
    <app-bar :minimize="isEntryPage" />
    <v-main>
      <fade-transition>
        <v-alert
          v-if="isPrivilegedOrHigher && isBitcoinAlertPage && !settings.hideBitcoinAlert"
          closable
          :icon="mdiTestTube"
          type="info"
          rounded="0"
          variant="tonal"
          @click:close="persistHideBitcoinAlert"
        >
          Bitcoin support is under active development. Results of transaction classification, address clustering and
          CoinJoin heuristics may change unannounced.
        </v-alert>
      </fade-transition>
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
import {
	BLOCKCHAIN_BTC, ROUTE_NAME_ENTRY_PAGE, ROUTE_NAME_WORKSPACE_PAGE,
} from './constants';
import AppBar from './components/appbar/AppBar.vue';
import FadeTransition from '@/components/common/FadeTransition.vue';
import {computed, onBeforeMount} from 'vue';
import {useRoute} from 'vue-router';
import {useTheme} from 'vuetify';
import {useLocalStore} from '@/pinia/local';
import {mdiTestTube} from '@mdi/js';
import {isAdminIdentity, isPrivilegedIdentity} from '@/utilities/index.js';

const route = useRoute();
const theme = useTheme();
const localStore = useLocalStore();

// Computed
const settings = computed({
	get() {
		return localStore.getSettings;
	},
	set(value) {
		localStore.setSettings(value);
	},
});

const isPrivilegedOrHigher = computed(() => isPrivilegedIdentity(localStore.getSession, settings.value.blockchainMode)
	|| isAdminIdentity(localStore.getSession, settings.value.blockchainMode));
const isEntryPage = computed(() => route.name === ROUTE_NAME_ENTRY_PAGE);
const isBitcoinAlertPage = computed(() => !isEntryPage.value && route.params.blockchainMode === BLOCKCHAIN_BTC
	&& route.name !== ROUTE_NAME_WORKSPACE_PAGE);

// Functions
function persistDarkTheme(isDark) {
	const set = settings.value;
	set.dark = isDark;
	settings.value = set;
}

function persistBlockchainMode(mode) {
	const set = settings.value;
	set.blockchainMode = mode;
	settings.value = set;
}

function persistHideBitcoinAlert() {
	const set = settings.value;
	set.hideBitcoinAlert = true;
	settings.value = set;
}

function setDarkTheme() {
	// Create new settings object if necessary
	if (settings.value.dark === null) {
		persistDarkTheme(window.matchMedia('(prefers-color-scheme: dark)').matches);
	}

	theme.global.name.value = settings.value.dark ? 'dark' : 'light';
}

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

// Hooks
onBeforeMount(() => {
	checkSessionExpiration();
	setDarkTheme();

	const mode = route.params.blockchainMode;
	if (mode !== undefined) {
		persistBlockchainMode(mode);
	}

	window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', e => {
		persistDarkTheme(e.matches);
		setDarkTheme();
	});
});

</script>

<style scoped>

</style>
