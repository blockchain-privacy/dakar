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
    <v-btn icon :to="{name: route.shortestPathPage}" v-if="showTools">
      <v-icon>{{ icon.mdiToolbox }}</v-icon>
    </v-btn>
    <v-menu offset-y style="z-index: 99">
      <template v-slot:activator="{ on, attrs }">
        <v-btn
            icon
            v-bind="attrs"
            v-on="on">
          <v-icon>{{ icon.mdiAccount }}</v-icon>
        </v-btn>
      </template>
      <v-list nav dense>
        <v-list-item v-if="this.userData">
          <v-list-item-icon>
            <v-icon>{{ icon.mdiAccountCircle }}</v-icon>
          </v-list-item-icon>
          <v-list-item-title> {{ this.userData.email }}</v-list-item-title>
        </v-list-item>
        <v-divider v-if="this.userData"/>
        <v-list-item :to="{name: route.userProfilePage}" v-if="this.userData"
                     :disabled="isUserProfileDisabled">
          <v-list-item-icon>
            <v-icon>{{ icon.mdiCog }}</v-icon>
          </v-list-item-icon>
          <v-list-item-title>Settings</v-list-item-title>
        </v-list-item>
        <v-list-item :to="{ name: route.userLoginPage }"
                     v-if="!this.userData" :disabled="isUserLoginDisabled">
          <v-list-item-icon>
            <v-icon>{{ icon.mdiLogin }}</v-icon>
          </v-list-item-icon>
          <v-list-item-title>Login</v-list-item-title>
        </v-list-item>
        <v-list-item :to="{ name: route.userAdminPage }" :disabled="isUserAdminDisabled"
                     v-if="showUserAdmin">
          <v-list-item-icon>
            <v-icon>{{ icon.mdiAccountSupervisor }}</v-icon>
          </v-list-item-icon>
          <v-list-item-title>User Administration</v-list-item-title>
        </v-list-item>
        <v-list-item @click="logout" v-if="this.userData">
          <v-list-item-icon>
            <v-icon>{{ icon.mdiLogout }}</v-icon>
          </v-list-item-icon>
          <v-list-item-title>Logout</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
  </v-app-bar>
</template>

<script>
import {
  mdiAccount, mdiAccountCircle, mdiAccountSupervisor, mdiCog, mdiLogin, mdiLogout, mdiToolbox,
} from '@mdi/js';
import QueryInput from './QueryInput.vue';
import {
  APPLICATION_NAME, ROUTE_NAME_ENTRY_PAGE,
  ROUTE_NAME_LOGIN_PAGE, ROUTE_NAME_SHORTEST_PATH_PAGE,
  ROUTE_NAME_USER_ADMIN_PAGE,
  ROUTE_NAME_USER_PROFILE_PAGE, ROUTE_USER_LOGOUT,
} from '../constants';
import {
  doGet, isAdminUser, isPrivilegedUser, resetLocal,
} from '../utilities';

export default {
  name: 'AppBar',
  components: {
    QueryInput,
  },
  props: {
    minimize: Boolean,
  },
  data() {
    return {
      applicationName: APPLICATION_NAME,
      icon: {
        mdiAccount,
        mdiLogin,
        mdiLogout,
        mdiAccountSupervisor,
        mdiAccountCircle,
        mdiCog,
        mdiToolbox,
      },
      route: {
        userProfilePage: ROUTE_NAME_USER_PROFILE_PAGE,
        userAdminPage: ROUTE_NAME_USER_ADMIN_PAGE,
        userLoginPage: ROUTE_NAME_LOGIN_PAGE,
        shortestPathPage: ROUTE_NAME_SHORTEST_PATH_PAGE,
        rootPage: ROUTE_NAME_ENTRY_PAGE,
      },
      isUserAdminDisabled: false,
      isUserLoginDisabled: false,
      isUserProfileDisabled: false,
    };
  },
  computed: {
    userData: {
      get() {
        return this.$store.getters.getActiveUser;
      },
      set(value) {
        this.$store.dispatch('setActiveUser', value);
      },
    },
    settings: {
      get() {
        return this.$store.getters.getSettings;
      },
      set(value) {
        this.$store.dispatch('setSettings', value);
      },
    },
    showUserAdmin() {
      return isAdminUser(this.userData);
    },
    showTools() {
      return isPrivilegedUser(this.userData) || isAdminUser(this.userData);
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
    logout() {
      doGet(ROUTE_USER_LOGOUT, this.$router)
        .then((data) => {
          if (data.success === undefined) throw Error('error logging out');
          if (data.success === false) {
            throw Error(data.msg);
          }
          resetLocal();
          this.userData = null;
          this.settings = null;
          this.goToPage(ROUTE_NAME_LOGIN_PAGE);
        })
        .catch((error) => {
          this.setErrorMessage(error);
        });
    },
  },
};
</script>

<style scoped>

</style>
