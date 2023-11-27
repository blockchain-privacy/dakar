<template>
  <!-- @click.stop so event does not bubble to parent component -->
  <v-switch
    v-model="darkModeEnabled"
    :inset="true"
    density="compact"
    hide-details
    :true-icon="mdiWeatherNight"
    :false-icon="mdiWeatherSunny"
    @click.stop="darkModeChange(!darkModeEnabled)"
  />
</template>

<script setup>
import {getLocalSettings} from '@/utilities';
import {mdiWeatherNight, mdiWeatherSunny} from '@mdi/js';
import {onBeforeMount, ref} from 'vue';
import {useStore} from 'vuex';
import {useTheme} from 'vuetify';

const store = useStore();
const theme = useTheme();

const darkModeEnabled = ref(false);

// Hooks
onBeforeMount(() => {
	const localSettings = getLocalSettings();
	darkModeEnabled.value = localSettings.dark;
});

// Functions
function persistDarkTheme(isDark) {
	const set = store.getters.getSettings;
	set.dark = isDark;
	store.dispatch('setSettings', set);
}

function darkModeChange(enabled) {
	darkModeEnabled.value = enabled;
	theme.global.name.value = enabled ? 'dark' : 'light';
	persistDarkTheme(enabled);
}

</script>

<style scoped>

</style>
