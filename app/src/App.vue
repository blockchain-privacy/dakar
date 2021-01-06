<template>
  <v-app>
    <v-app-bar app absolute color="primary" dark>
      <v-img
          @click="goToRoot()" style="cursor:pointer"
          alt="Dakar Logo"
          class="shrink mr-2"
          contain
          src="./assets/dakar_dash.svg"
          transition="scale-transition"
          width="32"/>
      <v-toolbar-title class="mx-2 hidden-sm-and-down" @click="goToRoot()" style="cursor:pointer">
        {{ applicationName }}
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <QueryInput class="mx-4"/>
      <v-spacer></v-spacer>
      <v-btn icon v-on:click="changeTheme()">
        <v-icon dark>mdi-invert-colors</v-icon>
      </v-btn>
    </v-app-bar>
    <v-main>
      <v-container fluid>
        <MsgBox/>
        <transition name="component-fade" mode="out-in">
          <router-view/>
        </transition>
      </v-container>
    </v-main>
    <v-footer app absolute>
      <v-spacer></v-spacer>
      <div>
        &copy; {{ new Date().getFullYear() }}
        <b>Dakar</b> - <a href="https://ntnu.no">NTNU</a>
      </div>
    </v-footer>
  </v-app>
</template>

<script>
import QueryInput from './components/QueryInput.vue';
import MsgBox from './components/MsgBox.vue';
import * as Constants from './constants';

export default {
  name: 'App',
  components: {
    MsgBox,
    QueryInput,
  },
  data() {
    return {
      applicationName: Constants.APPLICATION_NAME,
    };
  },
  methods: {
    changeTheme() {
      this.$vuetify.theme.dark = !this.$vuetify.theme.dark;
    },
    goToRoot() {
      // only change route if not already on entry page
      if (this.$route.name === Constants.ROUTE_NAME_ENTRY_PAGE) return;
      this.$router.push({ name: Constants.ROUTE_NAME_ENTRY_PAGE });
    },
  },
  beforeMount() {
    // eslint-disable-next-line no-console
    console.log(`Branch: ${__BRANCH__}, commit: ${__COMMIT_HASH__}`);
  },
};
</script>

<style>
.component-fade-enter-active, .component-fade-leave-active {
  transition: opacity 0.2s ease;
}
.component-fade-enter, .component-fade-leave-to
  /* .component-fade-leave-active below version 2.1.8 */ {
  opacity: 0;
}
</style>
