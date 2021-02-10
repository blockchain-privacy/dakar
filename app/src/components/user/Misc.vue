<template>
  <v-card
      class="mx-auto elevation-12"
      max-width="700">
    <v-toolbar color="primary" dark flat>
      <v-toolbar-title>
        <v-icon>{{ icon.mdiTune }}</v-icon>
        Misc
      </v-toolbar-title>
    </v-toolbar>
    <ProfileItem v-for="(item, index) in listItems"
                 :key="index"
                 :title="item.title"
                 :icon="item.icon"
                 :item-value="item.val"
                 :action-function="item.actionFunction"
                 :is-boolean="item.isBoolean"
                 :is-boolean-enabled="item.isBooleanEnabled"/>
  </v-card>
</template>

<script>
import {
  mdiInvertColors, mdiTune,
} from '@mdi/js';
import { PAGE_TITLE } from '../../constants';
import ProfileItem from './ProfileItem.vue';
import { getLocalSettings } from '../../utilities';

export default {
  name: 'Misc',
  components: { ProfileItem },
  data() {
    return {
      icon: {
        mdiInvertColors, mdiTune,
      },
      darkModeEnabled: false,
    };
  },
  computed: {
    listItems() {
      return [
        {
          title: 'Dark mode',
          icon: this.icon.mdiInvertColors,
          isBoolean: true,
          isBooleanEnabled: this.$vuetify.theme.dark,
          actionFunction: this.darkModeChange,
        },
      ];
    },
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
  mounted() {
    document.title = `Misc - ${PAGE_TITLE}`;
  },
};
</script>

<style scoped>

</style>
