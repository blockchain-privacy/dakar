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
const settings = computed({
	get() {
		return localStore.getSettings;
	},
	set(value) {
		localStore.setSettings(value);
	},
});
const isEntryPage = computed(() => route.name === ROUTE_NAME_ENTRY_PAGE);

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

function setDarkTheme() {
	// Create new settings object if necessary
	if (settings.value.dark === null) {
		persistDarkTheme(window.matchMedia('(prefers-color-scheme: dark)').matches);
	}

	theme.global.name.value = settings.value.dark ? 'dark' : 'light';
}

// Hooks
onBeforeMount(() => {
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
