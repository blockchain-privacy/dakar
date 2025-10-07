<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-btn-toggle
    v-model="themeModel"
    color="primary"
    mandatory
    variant="plain"
    density="comfortable"
    @click.stop="handleThemeChange(themeModel)"
  >
    <v-btn
      v-tooltip="{'text': 'Light Theme', 'location':'top', 'open-delay': 400}"
      :icon="mdiWeatherSunny"
      value="light"
    />
    <v-btn
      v-tooltip="{'text': 'Dark Theme', 'location':'top', 'open-delay': 400}"
      :icon="mdiWeatherNight"
      value="dark"
    />
    <v-btn
      v-tooltip="{'text': 'System Theme', 'location':'top', 'open-delay': 400}"
      :icon="mdiThemeLightDark"
      value="system"
    />
  </v-btn-toggle>
</template>

<script setup>
import {mdiThemeLightDark, mdiWeatherNight, mdiWeatherSunny} from '@mdi/js';
import {onBeforeMount, ref} from 'vue';
import {useTheme} from 'vuetify';
import {useLocalStore} from '@/pinia/local';

const localStore = useLocalStore();
const theme = useTheme();

const themeModel = ref('system');

// Hooks
onBeforeMount(() => {
	themeModel.value = localStore.getSettings.theme;
});

// Functions
function handleThemeChange(t) {
	theme.change(t);

	// Persist dark theme
	const set = localStore.getSettings;
	set.theme = t;
	localStore.setSettings(set);
}

</script>

<style scoped>

</style>
