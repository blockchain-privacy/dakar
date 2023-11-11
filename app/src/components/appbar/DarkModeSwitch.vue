<template>
  <!-- @click.stop so event does not bubble to parent component -->
  <v-switch
    v-model="model"
    :inset="true"
    density="compact"
    hide-details
    :true-icon="icons.mdiWeatherNight"
    :false-icon="icons.mdiWeatherSunny"
    @click.stop="model = !model"
  />
</template>

<script>
import {getLocalSettings} from '@/utilities';
import {mdiWeatherNight, mdiWeatherSunny} from '@mdi/js';

export default {
	name: 'DarkModeSwitch',
	data() {
		return {
			darkModeEnabled: false,
			icons: {mdiWeatherNight, mdiWeatherSunny},
		};
	},
	computed: {
		settings: {
			get() {
				return this.$store.getters.getSettings;
			},
			set(value) {
				this.$store.dispatch('setSettings', value);
			},
		},
		model: {
			get() {
				return this.darkModeEnabled;
			},
			set(enabled) {
				this.darkModeEnabled = enabled;
				this.darkModeChange(enabled);
			},
		},
	},
	beforeMount() {
		const settings = getLocalSettings();
		this.darkModeEnabled = settings.dark;
	},
	methods: {
		darkModeChange(enabled) {
			this.$vuetify.theme.global.name = enabled ? 'dark' : 'light';
			this.persistDarkTheme(enabled);
		},
		persistDarkTheme(isDark) {
			const set = this.settings;
			set.dark = isDark;
			this.settings = set;
		},
	},
};
</script>

<style scoped>

</style>
