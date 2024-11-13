<template>
  <v-app-bar
    absolute
    :flat="minimize"
    :color="minimize?'transparent':null"
  >
    <v-spacer v-if="minimize" />
    <router-link
      v-if="!minimize"
      id="app-logo"
      :to="{name: ROUTE_NAME_ENTRY_PAGE, params: {blockchainMode: settings.blockchainMode}}"
      class="ms-2"
    >
      <v-img
        style="cursor:pointer"
        alt="Dakar Logo"
        class="shrink mr-2"
        :src="DakarImg"
        transition="fade-transition"
        width="32"
      />
    </router-link>
    <router-link
      v-if="!minimize"
      :to="{name: ROUTE_NAME_ENTRY_PAGE,params: {blockchainMode: settings.blockchainMode}}"
      style="color: inherit; text-decoration: inherit"
    >
      <v-app-bar-title class="ms-2 d-none d-sm-flex">
        {{ APPLICATION_NAME }}
      </v-app-bar-title>
    </router-link>
    <query-input
      v-if="!minimize"
      class="mx-auto px-2"
      style="max-width: 600px"
    />
    <v-btn
      v-if="isPrivilegedOrHigher"
      icon
    >
      <v-icon>{{ mdiDotsGrid }}</v-icon>
      <page-menu />
    </v-btn>
    <v-menu v-if="isAdmin">
      <template #activator="{ props }">
        <v-btn
          v-bind="props"
          id="blockchain-mode"
          icon
        >
          <v-icon>{{ blockchainModeIcon }}</v-icon>
        </v-btn>
      </template>
      <v-list
        nav
        density="compact"
      >
        <v-list-item
          :active="settings.blockchainMode === BLOCKCHAIN_DASH"
          :to="{name: ROUTE_NAME_ENTRY_PAGE, params: {blockchainMode: BLOCKCHAIN_DASH}}"
        >
          <template #prepend>
            <v-icon :icon="getBlockchainModeIcon(BLOCKCHAIN_DASH)" />
          </template>
          <v-list-item-title>Dash</v-list-item-title>
        </v-list-item>
        <v-list-item
          :active="settings.blockchainMode === BLOCKCHAIN_BTC"
          :to="{name: ROUTE_NAME_ENTRY_PAGE, params: {blockchainMode: BLOCKCHAIN_BTC}}"
        >
          <template #prepend>
            <v-icon :icon="getBlockchainModeIcon(BLOCKCHAIN_BTC)" />
          </template>
          <v-list-item-title>Bitcoin</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
    <v-menu v-if="session">
      <template #activator="{ props }">
        <v-btn
          v-bind="props"
          id="app-bar-menu"
          icon
        >
          <v-icon>{{ mdiAccount }}</v-icon>
        </v-btn>
      </template>
      <v-list
        nav
        density="compact"
      >
        <v-list-item>
          <template #prepend>
            <v-icon :icon="mdiAccountCircle" />
          </template>
          <v-list-item-title> {{ session.identity.traits.email }}</v-list-item-title>
        </v-list-item>
        <v-divider />
        <v-list-item
          id="app-bar-settings"
          :to="{name: ROUTE_NAME_USER_PROFILE_PAGE}"
        >
          <template #prepend>
            <v-icon :icon="mdiCog" />
          </template>
          <v-list-item-title>Settings</v-list-item-title>
        </v-list-item>
        <v-list-item>
          <template #prepend>
            <v-icon :icon="mdiThemeLightDark" />
          </template>
          <div class="d-flex">
            <v-list-item-title style="display:flex; align-items:center">
              Dark Mode
            </v-list-item-title>
            <dark-mode-switch class="mt-0 ml-2" />
          </div>
        </v-list-item>
        <v-list-item
          id="app-bar-logout"
          @click="initLogoutFlow"
        >
          <template #prepend>
            <v-icon
              color="red"
              :icon="mdiLogout"
            />
          </template>
          <v-list-item-title>Logout</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
    <v-btn
      v-if="!session"
      variant="flat"
      color="primary"
      :to="{ name: ROUTE_NAME_LOGIN_PAGE }"
    >
      <v-icon>{{ mdiLogin }}</v-icon>
      Login
    </v-btn>
  </v-app-bar>
</template>

<script setup>
import {
	mdiAccount, mdiAccountCircle, mdiCog, mdiDotsGrid, mdiLogin, mdiLogout, mdiThemeLightDark,
} from '@mdi/js';
import PageMenu from './PageMenu.vue';
import QueryInput from './QueryInput.vue';
import DarkModeSwitch from './DarkModeSwitch.vue';
import {
	APPLICATION_NAME, BLOCKCHAIN_BTC, BLOCKCHAIN_DASH,
	ROUTE_NAME_ENTRY_PAGE,
	ROUTE_NAME_LOGIN_PAGE,
	ROUTE_NAME_USER_PROFILE_PAGE,
} from '@/constants';
import {isAdminIdentity, isPrivilegedIdentity} from '@/utilities';
import handleGetFlowError from '@/kratos';
import {computed, inject} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {useLocalStore} from '@/pinia/local';
import {useNavStore} from '@/pinia/nav';
import {useMsgStore} from '@/pinia/msg';
import DakarImg from '@/assets/dakar.svg?url';
import {bitcoinLogo, dashLogo} from '@/customIcons/index.js';

const ory = inject('ory');
const localStore = useLocalStore();
const route = useRoute();
const router = useRouter();
const context = {
	$route: route, $router: router, navStore: useNavStore(), localStore, msgStore: useMsgStore(),
};

defineProps({minimize: {type: Boolean, required: true}});

// Computed
const session = computed({
	get() {
		return localStore.getSession;
	},
	set(value) {
		localStore.setSession(value);
	},
});

const settings = computed({
	get() {
		return localStore.getSettings;
	},
	set(value) {
		localStore.setSettings(value);
	},
});

const isPrivilegedOrHigher = computed(() => isPrivilegedIdentity(session.value) || isAdminIdentity(session.value));
const isAdmin = computed(() => isAdminIdentity(session.value));
const blockchainModeIcon = computed(() => getBlockchainModeIcon(settings.value.blockchainMode));

// Functions
function getBlockchainModeIcon(mode) {
	switch (mode) {
		case BLOCKCHAIN_DASH: return dashLogo;
		case BLOCKCHAIN_BTC: return bitcoinLogo;
		default: return mdiCog;
	}
}

// GoToPage should receive a page name from ./constants
function goToPage(pageName, params) {
	// Only change route if not already on page
	if (route.name !== pageName || route.params !== params) {
		router.push({name: pageName, params});
	}
}

async function initLogoutFlow() {
	try {
		const response = await ory.frontend.createBrowserLogoutFlow();
		if (!response.logout_token) {
			return;
		}

		await ory.frontend.updateLogoutFlow({token: response.logout_token});
		session.value = null;
		goToPage(ROUTE_NAME_ENTRY_PAGE, {blockchainMode: settings.value.blockchainMode});
	} catch (e) {
		await handleGetFlowError(context, e, null);
	}
}

</script>

<style scoped>

</style>
