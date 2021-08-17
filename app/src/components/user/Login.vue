<template>
  <v-row align="center" no-gutters class="fill-height">
    <v-col cols="12" md="6" class="hidden-md-and-down fill-height">
      <v-sheet color="primary darken-2" dark height="100%" width="100%">
        <v-container class="justify-center fill-height">
          <div class="d-flex align-center flex-column text-center">
            <h1 class="text-xl-h1 text-md-h2 font-weight-bold">
              {{ applicationName }}
            </h1>
            <h3 class="text-xl-h3 text-md-h4 mt-4">
              {{ applicationSubtitle }}
            </h3>
          </div>
        </v-container>
      </v-sheet>
    </v-col>
    <v-col cols="12" lg="6">
      <v-container>
        <v-row justify="center">
          <v-col cols="12" lg="8" md="8" xl="5">
            <v-card class="elevation-4">
              <div class="pa-5">
                <h3 class="text-h3 font-weight-bold text-center">
                  Welcome!
                </h3>
                <v-form ref="loginForm" class="mt-4">
                  <v-text-field
                      autocomplete="username"
                      name="username"
                      v-model="email.value"
                      label="E-mail"
                      :prepend-inner-icon="icon.mdiEmail"
                      type="email"
                      :disabled="isSubmittingForm"
                      :rules="rules.emailRules"
                      @keydown.enter="submitForm"/>
                  <v-text-field
                      label="Password"
                      name="password"
                      autocomplete="current-password"
                      :prepend-inner-icon="icon.mdiLockOutline"
                      v-model="password.value"
                      :disabled="isSubmittingForm"
                      :type="password.show ? 'text' : 'password'"
                      :append-icon="password.show ?  icon.mdiEye : icon.mdiEyeOff"
                      @click:append="password.show = !password.show"
                      :hint="`At least ${passwordMinCharacters} characters`"
                      :rules="rules.passwordRules"
                      @keydown.enter="submitForm"/>
                  <v-alert type="error" v-if="loginFailed" dense>
                    Login failed!
                  </v-alert>
                  <v-btn
                      :loading="isSubmittingForm"
                      :disabled="isSubmittingForm"
                      block
                      class="font-weight-bold" color="primary darken-1"
                      @click="submitForm">
                    Login
                  </v-btn>
                </v-form>
                <NamedDivider title="Or"/>
                <div class="text-center">
                  <v-btn disabled
                         block
                         class="font-weight-bold" color="primary darken-1" large to="/">
                    Register
                  </v-btn>
                </div>
              </div>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-col>
  </v-row>
</template>

<script>
import {
  mdiLockOutline, mdiEye, mdiEyeOff, mdiEmail,
} from '@mdi/js';
import NamedDivider from '../common/NamedDivider.vue';
import {
  APPLICATION_NAME, PAGE_TITLE, PASSWORD_MIN_CHARACTERS, ROUTE_NAME_ENTRY_PAGE,
  PASSWORD_MAX_CHARACTERS, ROUTE_USER_LOGIN, DEFAULT_SETTINGS, APPLICATION_SUBTITLE,
} from '../../constants';
import {
  doPost, emailRules, getLocalSettings, passwordRules,
} from '../../utilities';

function goToPage(context, pageObj) {
  context.$router.push(pageObj);
}

function goToRoot(context) {
  goToPage(context, { name: ROUTE_NAME_ENTRY_PAGE });
}

export default {
  name: 'Login',
  components: { NamedDivider },
  data() {
    return {
      icon: {
        mdiLockOutline, mdiEye, mdiEyeOff, mdiEmail,
      },
      isSubmittingForm: false,
      loginFailed: false,
      applicationName: APPLICATION_NAME,
      applicationSubtitle: APPLICATION_SUBTITLE,
      passwordMinCharacters: PASSWORD_MIN_CHARACTERS,
      passwordMaxCharacters: PASSWORD_MAX_CHARACTERS,
      rules: { passwordRules, emailRules },
      email: {
        value: '',
      },
      password: {
        value: '',
        show: false,
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
    failedRoute: {
      get() {
        return this.$store.getters.getFailedRoute;
      },
      set(value) {
        this.$store.dispatch('setFailedRoute', value);
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
  },
  methods: {
    setErrorMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'error', temporary: true });
    },
    validateLoginForm() {
      return this.$refs.loginForm.validate();
    },
    leave() {
      if (this.failedRoute) {
        goToPage(this, this.failedRoute);
        this.failedRoute = null;
      } else goToRoot(this);
    },
    sendLoginRequest() {
      this.isSubmittingForm = true;
      this.loginFailed = false;

      doPost(ROUTE_USER_LOGIN, this.$router, this.$store,
        { user_pw: this.password.value, user_email: this.email.value })
        .then((data) => {
          if (data.success === undefined
                || data.user === undefined) throw Error('error logging in.');
          if (data.success === false) {
            throw Error(data.msg);
          }

          // set user data
          this.userData = data.user;

          // load settings from localStorage
          const localStorageSettingsData = getLocalSettings();
          if (localStorageSettingsData !== null) {
            this.settings = localStorageSettingsData;
            // dark mode according to settings
            this.$vuetify.theme.dark = this.settings.dark;
          } else {
            const defaultSettings = DEFAULT_SETTINGS;
            defaultSettings.dark = this.$vuetify.theme.dark;
            this.settings = defaultSettings;
          }

          // user is logged in -> leave login page
          this.leave();
        })
        .catch((error) => {
          this.setErrorMessage(error);
          this.loginFailed = true;
        })
        .finally(() => {
          this.isSubmittingForm = false;
        });
    },
    submitForm() {
      // already submitting
      if (this.isSubmittingForm) return;
      this.loginFailed = false;
      if (!this.validateLoginForm()) {
        this.isSubmittingForm = false;
        return;
      }

      this.sendLoginRequest();
    },
  },
  mounted() {
    document.title = `Login - ${PAGE_TITLE}`;
  },
};
</script>

<style scoped>

</style>
