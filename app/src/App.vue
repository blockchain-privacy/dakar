<template>
  <v-app>
    <v-app-bar app color="primary" dark>
      <v-img
          @click="goToRoot()" style="cursor:pointer"
          alt="Explorer Logo"
          class="shrink mr-2"
          contain
          src="https://cdn.vuetifyjs.com/images/logos/vuetify-logo-dark.png"
          transition="scale-transition"
          width="40"/>
      <v-toolbar-title class="mx-1" @click="goToRoot()" style="cursor:pointer">Dash Explorer</v-toolbar-title>
      <v-spacer></v-spacer>
      <QueryInput class="mx-4"/>
      <v-spacer></v-spacer>
    </v-app-bar>
    <v-main>
      <v-content>
        <v-container>
          <MsgBox/>
          <transition name="fade">
            <router-view/>
          </transition>
        </v-container>
      </v-content>
    </v-main>

    <v-footer app>
      <v-spacer></v-spacer>
      <div>
        &copy; {{ new Date().getFullYear() }}
        <b>Dakar project</b> - <a href="http://ntnu.no">NTNU</a>
      </div>
    </v-footer>
  </v-app>
</template>

<script>
import QueryInput from "./components/QueryInput";
import MsgBox from "./components/MsgBox";
import * as Utility from './utilities';
import * as Constants from './constants';

export default {
  name: 'App',
  components: {
    MsgBox,
    QueryInput
  },
  methods: {
    goToRoot() {
      // only change route if not already on entry page
      if (this.$route.name === Constants.ROUTE_NAME_ENTRY_PAGE)
        return;
      this.$router.push({name: Constants.ROUTE_NAME_ENTRY_PAGE});
    }
  },
  watch: {
    // global route watcher
    '$route'() {
      // reset data on every route change
      Utility.resetData(this);
    }
  }
};
</script>
