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
    <v-menu offset-y style="z-index: 99" v-if="this.userData">
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
          <v-list-item-title> {{ this.userData.email }}</v-list-item-title>
        </v-list-item>
        <v-divider />
        <v-list-item :to="{name: route.userProfilePage}">
          <v-list-item-icon>
            <v-icon>{{ icon.mdiCog }}</v-icon>
          </v-list-item-icon>
          <v-list-item-title>Settings</v-list-item-title>
        </v-list-item>
        <v-list-item @click="logout">
          <v-list-item-icon>
            <v-icon>{{ icon.mdiLogout }}</v-icon>
          </v-list-item-icon>
          <v-list-item-title>Logout</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
    <v-btn depressed color="primary" :to="{ name: route.userLoginPage }" v-if="!this.userData">
      <v-icon>{{ icon.mdiLogin }}</v-icon> Login
    </v-btn>
  </v-app-bar>
</template>

<script>
import {
  mdiAccount, mdiAccountCircle, mdiCog, mdiLogin, mdiLogout,
  mdiDotsGrid,
} from '@mdi/js';
import PageMenu from './PageMenu.vue';
import QueryInput from './QueryInput.vue';
import {
  APPLICATION_NAME, ROUTE_NAME_ENTRY_PAGE, ROUTE_NAME_LOGIN_PAGE,
  ROUTE_NAME_USER_PROFILE_PAGE, ROUTE_USER_LOGOUT,
} from '../constants';
import {
  doGet, isAdminUser, resetLocal, isPrivilegedUser,
} from '../utilities';

export default {
  name: 'AppBar',
  components: {
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
      },
      route: {
        userProfilePage: ROUTE_NAME_USER_PROFILE_PAGE,
        userLoginPage: ROUTE_NAME_LOGIN_PAGE,
        rootPage: ROUTE_NAME_ENTRY_PAGE,
      },
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
    isPrivilegedOrHigher() {
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
