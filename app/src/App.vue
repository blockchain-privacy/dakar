<template>
  <v-app>
    <v-app-bar app absolute>
      <v-img
          @click="goToRoot()" style="cursor:pointer"
          alt="Dakar Logo"
          class="shrink mr-2"
          contain
          src="./assets/dakar_dash.svg"
          transition="scale-transition"
          width="32"/>
      <v-toolbar-title class="mx-2 d-none d-sm-flex" @click="goToRoot()" style="cursor:pointer">
        {{ applicationName }}
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <QueryInput class="mx-4"/>
      <v-spacer></v-spacer>
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
            <v-list-item-title>{{ this.userData.email }}</v-list-item-title>
          </v-list-item>
          <v-list-item @click="goToLogin" v-if="!this.userData" :disabled="isUserLoginDisabled">
            <v-list-item-icon>
              <v-icon>{{ icon.mdiLogin }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Login</v-list-item-title>
          </v-list-item>
          <v-list-item @click="goToUserAdministration" :disabled="isUserAdminDisabled"
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
      <v-btn icon v-on:click="changeTheme()">
        <v-icon dark>{{ icon.mdiInvertColors }}</v-icon>
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
import {
  mdiInvertColors, mdiAccount, mdiLogin, mdiLogout, mdiAccountSupervisor, mdiAccountCircle,
} from '@mdi/js';
import QueryInput from './components/QueryInput.vue';
import MsgBox from './components/MsgBox.vue';
import * as Constants from './constants';
import '@fontsource/roboto';
import { LOCALSTORAGE_FIELD_USER, ROUTE_USER_LOGOUT } from './constants';

export default {
  name: 'App',
  components: {
    MsgBox,
    QueryInput,
  },
  data() {
    return {
      applicationName: Constants.APPLICATION_NAME,
      icon: {
        mdiInvertColors, mdiAccount, mdiLogin, mdiLogout, mdiAccountSupervisor, mdiAccountCircle,
      },
      isUserAdminDisabled: false,
      isUserLoginDisabled: false,
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
    errMsg: {
      get() {
        return this.$store.getters.getErrorMsg;
      },
      set(value) {
        this.$store.dispatch('setErrorMsg', value);
      },
    },
    showUserAdmin() {
      return this.userData && this.userData.roles && this.userData.roles.some((d) => d.role_name === 'admin');
    },
  },
  methods: {
    changeTheme() {
      this.$vuetify.theme.dark = !this.$vuetify.theme.dark;
    },
    goToRoot() {
      this.goToPage(Constants.ROUTE_NAME_ENTRY_PAGE);
    },
    goToLogin() {
      this.goToPage(Constants.ROUTE_NAME_LOGIN_PAGE);
    },
    goToUserAdministration() {
      this.goToPage(Constants.ROUTE_NAME_USER_ADMIN_PAGE);
    },
    // goToPage should receive a page name from ./constants
    goToPage(pageName) {
      // only change route if not already on page
      if (this.$route.name !== pageName) this.$router.push({ name: pageName });
    },
    logout() {
      fetch(ROUTE_USER_LOGOUT)
        .then((response) => response.json())
        .then((data) => {
          if (data.success === undefined) throw Error('error deleting user');
          if (data.success === false) {
            throw Error(data.msg);
          }
          localStorage.removeItem(LOCALSTORAGE_FIELD_USER);
          this.userData = null;
          this.goToLogin();
        })
        .catch((error) => {
          this.errMsg = error;
        });
    },
    checkRoute(routeName) {
      this.isUserLoginDisabled = false;
      this.isUserAdminDisabled = false;

      if (routeName === Constants.ROUTE_NAME_LOGIN_PAGE) {
        this.isUserLoginDisabled = true;
      } else if (routeName === Constants.ROUTE_NAME_USER_ADMIN_PAGE) {
        this.isUserAdminDisabled = true;
      }
    },
  },
  beforeMount() {
    this.checkRoute(this.$router.currentRoute.name);
    // get user information from localStorage
    const localStorageUserData = localStorage.getItem(LOCALSTORAGE_FIELD_USER);
    if (localStorageUserData !== null) {
      this.userData = JSON.parse(localStorageUserData);
    }
    // eslint-disable-next-line no-console
    console.log(`Branch: ${__BRANCH__}, commit: ${__COMMIT_HASH__}`);
  },
  watch: {
    $route(to) {
      this.checkRoute(to.name);
    },
  },
};
</script>

<style>
.component-fade-enter-active, .component-fade-leave-active {
  transition: opacity 0.2s ease;
}

.component-fade-enter, .component-fade-leave-to {
  opacity: 0;
}
</style>
