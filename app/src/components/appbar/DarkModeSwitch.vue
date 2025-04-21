<template>
  <!-- @click.stop so event does not bubble to parent component -->
  <v-switch
    v-model="darkModeEnabled"
    inset
    density="compact"
    hide-details
    :true-icon="mdiWeatherNight"
    :false-icon="mdiWeatherSunny"
    @click.stop="darkModeChange(!darkModeEnabled)"
  />
</template>

<script setup>
import {getLocalstorageData} from '@/utilities';
import {mdiWeatherNight, mdiWeatherSunny} from '@mdi/js';
import {onBeforeMount, ref} from 'vue';
import {useTheme} from 'vuetify';
import {useLocalStore} from '@/pinia/local';
import {LOCALSTORAGE_FIELD_SETTINGS} from '@/constants/index.js';

const localStore = useLocalStore();
const theme = useTheme();

const darkModeEnabled = ref(false);

// Hooks
onBeforeMount(() => {
	const localSettings = getLocalstorageData(LOCALSTORAGE_FIELD_SETTINGS);
	darkModeEnabled.value = localSettings.dark;
});

// Functions
function darkModeChange(enabled) {
	darkModeEnabled.value = enabled;
	theme.global.name.value = enabled ? 'dark' : 'light';

	// Persist dark theme
	const set = localStore.getSettings;
	set.dark = enabled;
	localStore.setSettings(set);
}

</script>

<style scoped>

</style>
