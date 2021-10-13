<template>
  <v-switch v-model="darkModeEnabled" inset dense hide-details @change="darkModeChange" />
</template>

<script>
import { getLocalSettings } from '../utilities';

export default {
  name: 'DarkModeSwitch',
  props: {
    icon: { type: String, required: false },
  },
  data() {
    return {
      darkModeEnabled: false,
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
  },
  methods: {
    darkModeChange(enabled) {
      this.$vuetify.theme.dark = enabled;
      this.persistDarkTheme(this.$vuetify.theme.dark);
    },
    persistDarkTheme(isDark) {
      const set = this.settings;
      set.dark = isDark;
      this.settings = set;
    },
  },
  beforeMount() {
    const settings = getLocalSettings();
    this.darkModeEnabled = settings.dark;
  },
};
</script>

<style scoped>

</style>
