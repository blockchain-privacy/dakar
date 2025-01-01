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
	BLOCKCHAIN_BTC, RESPONSE_TYPE_ADDRESS, RESPONSE_TYPE_BLOCK, RESPONSE_TYPE_TRANSACTION,
	ROUTE_NAME_ADDRESS_PAGE,
	ROUTE_NAME_BLOCK_PAGE,
	ROUTE_NAME_ENTRY_PAGE,
	ROUTE_NAME_TRANSACTION_PAGE,
	ROUTE_NAME_WORKSPACE_PAGE,
} from './constants';
import AppBar from './components/appbar/AppBar.vue';
import FadeTransition from '@/components/common/FadeTransition.vue';
import {computed, onBeforeMount, watch} from 'vue';
import {useRoute} from 'vue-router';
import {useTheme} from 'vuetify';
import {useLocalStore} from '@/pinia/local';
import {mdiTestTube} from '@mdi/js';
import {
	getDakarClients,
	handleError, handleQuery, isAdminIdentity, isPrivilegedIdentity,
} from '@/utilities/index.js';
import {useExplorerStore} from '@/pinia/explorer.js';
import {storeToRefs} from 'pinia';
import {useNavStore} from '@/pinia/nav.js';
import {useMsgStore} from '@/pinia/msg.js';

const route = useRoute();
const theme = useTheme();
const msgStore = useMsgStore();
const localStore = useLocalStore();
const explorerStore = useExplorerStore();
const {pushFromUserInput} = storeToRefs(useNavStore());
const context = {addMessage: msgStore.addMessage, $route: route};

// When the blockchain mode is switched and the current component is not reloaded,
// the dakar client is in the wrong state. As a workaround, get all available dakar
// clients and select the right one when doing a request.
const dakarClients = getDakarClients();

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

// Watch
watch(route, () => {
	newRouting();
});

// Functions
async function newRouting() {
	const {id} = route.params;
	const isPushFromUserInput = pushFromUserInput.value;

	if (isPushFromUserInput) {
		pushFromUserInput.value = false;
	}

	if (isPushFromUserInput || !id
		|| !(route.name === ROUTE_NAME_BLOCK_PAGE
		|| route.name === ROUTE_NAME_ADDRESS_PAGE
		|| route.name === ROUTE_NAME_TRANSACTION_PAGE)) {
		return;
	}

	let e;
	switch (route.name) {
		case ROUTE_NAME_TRANSACTION_PAGE:
			e = await handleQuery(id, explorerStore, dakarClients[settings.value.blockchainMode], RESPONSE_TYPE_TRANSACTION);
			break;
		case ROUTE_NAME_BLOCK_PAGE:
			e = await handleQuery(id, explorerStore, dakarClients[settings.value.blockchainMode], RESPONSE_TYPE_BLOCK);
			break;
		case ROUTE_NAME_ADDRESS_PAGE:
			e = await handleQuery(id, explorerStore, dakarClients[settings.value.blockchainMode], RESPONSE_TYPE_ADDRESS);
			break;
		default:
			e = await handleQuery(id, explorerStore, dakarClients[settings.value.blockchainMode]);
	}

	if (!e) {
		handleError(context, e);
	}
}

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

// Initial routing
newRouting();

</script>

<style scoped>

</style>
