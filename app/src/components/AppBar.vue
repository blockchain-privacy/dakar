<template>
  <v-app-bar app absolute :flat="minimize" :color="minimize?'transparent':null">
    <router-link v-if="!minimize" :to="{name: route.rootPage}">
      <v-img
          style="cursor:pointer"
          alt="Dakar Logo"
          class="shrink mr-2"
          contain
          src="../assets/dakar_dash.svg"
          transition="scale-transition"
          width="32">
      </v-img>
    </router-link>
    <router-link
        v-if="!minimize"
        :to="{name: route.rootPage}"
        style="color: inherit; text-decoration: inherit">
      <v-toolbar-title class="mx-2 d-none d-sm-flex" style="cursor:pointer">
        {{ applicationName }}
      </v-toolbar-title>
    </router-link>
    <v-spacer></v-spacer>
    <QueryInput v-if="!minimize" class="mx-4"/>
    <v-spacer></v-spacer>
    <PageMenu v-model="showPageMenu" v-if="isPrivilegedOrHigher">
      <v-btn icon @click="showPageMenu = !showPageMenu">
        <v-icon>{{ icon.mdiDotsGrid }}</v-icon>
      </v-btn>
    </PageMenu>
    <v-menu offset-y style="z-index: 99" v-if="this.session">
      <template v-slot:activator="{ on, attrs }">
        <v-btn
            icon
            v-bind="attrs"
            v-on="on">
          <v-icon>{{ icon.mdiAccount }}</v-icon>
        </v-btn>
      </template>
      <v-list nav dense>
        <v-list-item>
          <v-list-item-icon>
            <v-icon>{{ icon.mdiAccountCircle }}</v-icon>
          </v-list-item-icon>
          <v-list-item-title> {{ this.session.identity.traits.email }}</v-list-item-title>
        </v-list-item>
        <v-divider/>
        <v-list-item :to="{name: route.userProfilePage}">
          <v-list-item-icon>
            <v-icon>{{ icon.mdiCog }}</v-icon>
          </v-list-item-icon>
          <v-list-item-title>Settings</v-list-item-title>
        </v-list-item>
        <v-list-item>
          <v-list-item-icon>
            <v-icon>{{ icon.mdiWeatherNight }}</v-icon>
          </v-list-item-icon>
          <v-list-item-title>Dark Mode</v-list-item-title>
          <DarkModeSwitch class="mt-0 ml-2"/>
        </v-list-item>
        <v-list-item @click="initLogoutFlow">
          <v-list-item-icon>
            <v-icon color="red">{{ icon.mdiLogout }}</v-icon>
          </v-list-item-icon>
          <v-list-item-title>Logout</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
    <v-btn
        v-if="!this.session"
        depressed
        color="primary"
        :to="{ name: route.userLoginPage }">
      <v-icon>{{ icon.mdiLogin }}</v-icon>
      Login
    </v-btn>
  </v-app-bar>
</template>

<script>
import {
  mdiAccount, mdiAccountCircle, mdiCog, mdiLogin, mdiLogout,
  mdiDotsGrid, mdiWeatherNight,
} from '@mdi/js';
import PageMenu from './PageMenu.vue';
import QueryInput from './QueryInput.vue';
import DarkModeSwitch from './DarkModeSwitch.vue';
import {
  APPLICATION_NAME, ROUTE_NAME_ENTRY_PAGE, ROUTE_NAME_LOGIN_PAGE,
  ROUTE_NAME_USER_PROFILE_PAGE,
} from '../constants';
import { resetLocal, isAdminIdentity, isPrivilegedIdentity } from '../utilities';
import handleGetFlowError from '../kratos';

export default {
  name: 'AppBar',
  components: {
    DarkModeSwitch,
    PageMenu,
    QueryInput,
  },
  props: {
    minimize: Boolean,
  },
  data() {
    return {
      applicationName: APPLICATION_NAME,
      showPageMenu: false,
      icon: {
        mdiAccount,
        mdiLogin,
        mdiLogout,
        mdiAccountCircle,
        mdiCog,
        mdiDotsGrid,
        mdiWeatherNight,
      },
      route: {
        userProfilePage: ROUTE_NAME_USER_PROFILE_PAGE,
        userLoginPage: ROUTE_NAME_LOGIN_PAGE,
        rootPage: ROUTE_NAME_ENTRY_PAGE,
      },
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
    session: {
      get() {
        return this.$store.getters.getSession;
      },
      set(value) {
        this.$store.dispatch('setSession', value);
      },
    },
    showUserAdmin() {
      return isAdminIdentity(this.session);
    },
    isPrivilegedOrHigher() {
      return isPrivilegedIdentity(this.session) || this.showUserAdmin;
    },
  },
  methods: {
    setErrorMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'error', temporary: true });
    },
    // goToPage should receive a page name from ./constants
    goToPage(pageName) {
      // only change route if not already on page
      if (this.$route.name !== pageName) this.$router.push({ name: pageName });
    },
    initLogoutFlow() {
      this.ory.createSelfServiceLogoutFlowUrlForBrowsers()
        .then((d) => {
          if (d.status === 200 && d.data.logout_token) {
            return this.ory.submitSelfServiceLogoutFlow(d.data.logout_token);
          }
          return Promise.resolve();
        })
        .then((d) => {
          if (d.status === 204) {
            resetLocal();
            this.settings = null;
            this.session = null;
            this.goToPage(this.route.rootPage);
          }
        })
        .catch((err) => {
          handleGetFlowError(this.$router, this.$store, err);
        });
    },
  },
};
</script>

<style scoped>

</style>
