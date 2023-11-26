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
import {
	DEFAULT_SETTINGS, ROUTE_NAME_ENTRY_PAGE,
} from './constants';
import AppBar from './components/appbar/AppBar.vue';
import {isSessionExpired} from './utilities';
import FadeTransition from '@/components/common/FadeTransition.vue';
import {computed, onBeforeMount, onMounted, toRaw} from 'vue';
import {useRoute} from 'vue-router';
import {useStore} from 'vuex';
import {useTheme} from 'vuetify';

const route = useRoute();
const store = useStore();
const theme = useTheme();

// Computed
const settings = computed({
	get() {
		return store.getters.getSettings;
	},
	set(value) {
		store.dispatch('setSettings', value);
	},
});

const session = computed({
	get() {
		return store.getters.getSession;
	},
	set(value) {
		store.dispatch('setSession', value);
	},
});

const isEntryPage = computed(() => route.name === ROUTE_NAME_ENTRY_PAGE);

// Functions
function persistDarkTheme(isDark) {
	const set = settings.value;
	set.dark = isDark;
	settings.value = set;
}

function setDarkTheme() {
	// Create new settings object if necessary
	if (settings.value === null) {
		const defaultSettings = DEFAULT_SETTINGS;
		defaultSettings.dark = window.matchMedia('(prefers-color-scheme: dark)').matches;
		settings.value = defaultSettings;
	}

	theme.global.name.value = settings.value.dark ? 'dark' : 'light';
}

// Hooks
onMounted(() => {
	if (isSessionExpired(toRaw(session.value))) {
		session.value = null;
	}
});

onBeforeMount(() => {
	setDarkTheme();
	window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', e => {
		persistDarkTheme(e.matches);
		setDarkTheme();
	});
});

</script>

<style scoped>

</style>
