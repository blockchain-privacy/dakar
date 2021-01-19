<template>
  <!--  negative margin so we do not have a gutter to the navbar-->
  <v-row align="center" no-gutters style="margin: -12px">
    <v-col cols="12" md="6" class="hidden-md-and-down">
      <v-sheet color="primary darken-2" dark height="90vh" width="100%">
        <v-container  fill-height class="justify-center">
          <div class="d-flex align-center flex-column text-center">
            <h1 class="text-xl-h1 text-md-h2 font-weight-bold">
              {{ applicationName }}
            </h1>
            <h3 class="text-xl-h3 text-md-h4 mt-4">
              Blockchain transaction analytics.
            </h3>
          </div>
        </v-container>
      </v-sheet>
    </v-col>
    <v-col cols="12" lg="6">
      <v-container>
        <v-row justify="center">
          <v-col cols="12" lg="8" md="8" xl="5">
            <v-card>
              <div class="pa-5">
                <h3 class="text-h3 font-weight-bold text-center">
                  Welcome!
                </h3>
                <v-form ref="loginForm" class="mt-4">
                  <v-text-field
                      v-model="email.value"
                      label="E-mail"
                      :prepend-inner-icon="icon.mdiEmail"
                      type="email"
                      :disabled="isSubmittingForm"
                      :rules="rules.emailRules"
                      @keydown.enter="submitForm"/>
                  <v-text-field label="Password"
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

                <div class="my-8">
                  <div class="d-flex align-center">
                    <v-sheet color="grey lighten-2" height="1" width="100%"/>
                    <div class="mx-2">Or</div>
                    <v-sheet color="grey lighten-2" height="1" width="100%"/>
                  </div>
                </div>
                <div class="text-center">
                  <v-btn block class="font-weight-bold" color="primary darken-1" large to="/">
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
import {
  APPLICATION_NAME, PAGE_TITLE, PASSWORD_MIN_CHARACTERS, ROUTE_NAME_ENTRY_PAGE,
  PASSWORD_MAX_CHARACTERS, ROUTE_USER_LOGIN,
} from '../constants';
import { emailRules } from '../utilities';

const notAllowedWhitespaceCharacters = [
  '\b', '\t', '\n', '\v', '\f', '\r',
  '\u0008', '\u0009', '\u000A', '\u000B', '\u000C',
  '\u000D', '\u0022', '\u0027', '\u005C',
  '\u00A0', '\u2028', '\u2029', '\uFEFF'];

// hasWhitespace checks if the given string
// contains any of the characters in notAllowedWhitespaceCharacters
// credit: https://stackoverflow.com/questions/1731190/check-if-a-string-has-white-space
const hasWhitespace = (char) => notAllowedWhitespaceCharacters.some(
  (w) => char.indexOf(w) > -1,
  notAllowedWhitespaceCharacters,
);

function goToRoot(context) {
  context.$router.push({ name: ROUTE_NAME_ENTRY_PAGE });
}

export default {
  name: 'Login',
  data() {
    return {
      icon: {
        mdiLockOutline, mdiEye, mdiEyeOff, mdiEmail,
      },
      isSubmittingForm: false,
      loginFailed: false,
      applicationName: APPLICATION_NAME,
      passwordMinCharacters: PASSWORD_MIN_CHARACTERS,
      passwordMaxCharacters: PASSWORD_MAX_CHARACTERS,
      rules: {
        passwordRules: [
          (v) => !!v || 'Password is required',
          (v) => !hasWhitespace(v) || 'Password contains white space characters',
          (v) => v.length >= PASSWORD_MIN_CHARACTERS || `At least ${PASSWORD_MIN_CHARACTERS} characters`,
          (v) => (v && v.length < PASSWORD_MAX_CHARACTERS)
              || `Password must be less than ${PASSWORD_MAX_CHARACTERS} characters`,
        ],
        emailRules,
      },
      email: {
        value: '',
      },
      password: {
        value: '',
        show: false,
      },
    };
  },
  methods: {
    validateLoginForm() {
      return this.$refs.loginForm.validate();
    },
    sendLoginRequest() {
      this.isSubmittingForm = true;
      this.loginFailed = false;

      fetch(ROUTE_USER_LOGIN, {
        method: 'POST', // or 'PUT'
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ user_pw: this.password.value, user_email: this.email.value }),
      })
        .then((response) => response.json())
        .then((data) => {
          if (data.success === undefined) throw Error('error logging in ');
          if (data.success === false) {
            throw Error(data.msg);
          }
          goToRoot(this);
        })
        .catch((error) => {
          this.errMsg = error;
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
