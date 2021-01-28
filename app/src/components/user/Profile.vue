<template>
  <v-card
      class="mx-auto elevation-12"
      max-width="700">
    <v-toolbar color="primary" dark flat>
      <v-toolbar-title>
        <v-icon>{{ icon.mdiAccountDetails }}</v-icon>
        Profile
      </v-toolbar-title>
    </v-toolbar>
    <ProfileItem v-for="(item, index) in listItems"
                 :key="index"
                 :title="item.title"
                 :icon="item.icon"
                 :item-value="item.val"
                 :action-function="item.actionFunction"/>
    <v-dialog v-model="showModifyEmailDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">Change E-mail</span>
        </v-card-title>
        <v-card-text>
          <v-container>
            <v-row>
              <v-form ref="modifyEmailForm">
                <v-text-field
                    v-model="editedEmailItem.email"
                    label="E-mail"
                    type="email"
                    :rules="rules.emailRules">
                </v-text-field>
                <v-text-field
                    v-model="editedEmailItem.current_password"
                    label="Current password"
                    type="password"
                    :rules="rules.passwordRules">
                </v-text-field>
              </v-form>
            </v-row>
          </v-container>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text @click="closeEmailForm">Cancel</v-btn>
          <v-btn color="blue darken-1" text @click="saveEmailForm">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="showModifyPasswordDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">Change Password</span>
        </v-card-title>
        <v-card-text>
          <v-container>
            <v-row>
              <v-form ref="modifyPasswordForm">
                <v-text-field
                    v-model="editedPasswordItem.new_password"
                    label="New Password"
                    type="password"
                    :rules="rules.passwordRules">
                </v-text-field>
                <v-text-field
                    v-model="editedEmailItem.current_password"
                    label="Current password"
                    type="password"
                    :rules="rules.passwordRules">
                </v-text-field>
              </v-form>
            </v-row>
          </v-container>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text @click="closePasswordForm">Cancel</v-btn>
          <v-btn color="blue darken-1" text @click="savePasswordForm">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script>
import {
  mdiLock, mdiEmail, mdiCalendar, mdiCalendarEdit, mdiAccountDetails,
} from '@mdi/js';
import {
  LOCALSTORAGE_FIELD_USER, PAGE_TITLE, ROUTE_USER_MODIFY,
} from '../../constants';
import ProfileItem from './ProfileItem.vue';
import {
  doPost, emailRules, handleError, passwordRules,
} from '../../utilities';

export default {
  name: 'Profile',
  components: { ProfileItem },
  data() {
    return {
      icon: {
        mdiLock, mdiEmail, mdiCalendar, mdiCalendarEdit, mdiAccountDetails,
      },
      showModifyEmailDialog: false,
      showModifyPasswordDialog: false,
      rules: {
        emailRules, passwordRules,
      },
      editedEmailItem: {
        email: '',
        current_password: '',
      },
      defaultEmailItem: {
        email: '',
        current_password: '',
      },
      editedPasswordItem: {
        new_password: '',
        current_password: '',
      },
      defaultPasswordItem: {
        new_password: '',
        current_password: '',
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
    modifiedDate() {
      return new Date(this.userData.modified).toLocaleString();
    },
    createdDate() {
      return new Date(this.userData.created).toLocaleString();
    },
    listItems() {
      return [
        {
          title: 'Email:',
          val: this.userData.email,
          icon: this.icon.mdiEmail,
          actionFunction: this.openEmailForm,
        },
        {
          title: 'Change Password:',
          val: '********',
          icon: this.icon.mdiLock,
          actionFunction: this.openPasswordForm,
        },
        {
          title: 'Account last modified:',
          val: this.modifiedDate,
          icon: this.icon.mdiCalendarEdit,
        },
        {
          title: 'Account created:',
          val: this.createdDate,
          icon: this.icon.mdiCalendar,
        },
      ];
    },
  },
  methods: {
    /** Email methods * */
    closeEmailForm() {
      this.showModifyEmailDialog = false;
      this.$nextTick(() => {
        this.editedEmailItem = { ...this.defaultEmailItem };
      });
    },
    openEmailForm() {
      this.showModifyEmailDialog = true;
    },
    validateEmailForm() {
      return this.$refs.modifyEmailForm.validate();
    },
    saveEmailForm() {
      if (!this.validateEmailForm()) return;

      doPost(ROUTE_USER_MODIFY, this.$router, {
        uid: this.userData.uid,
        email: this.editedEmailItem.email,
        current_password: this.editedEmailItem.current_password,
      })
        .then((data) => {
          if (data.user) {
            localStorage.setItem(LOCALSTORAGE_FIELD_USER, JSON.stringify(data.user));
            this.userData = data.user;
          }
        })
        .catch((e) => {
          handleError(this.$store, e);
        })
        .finally(() => {
          this.closeEmailForm();
        });
    },
    /** Password methods * */
    closePasswordForm() {
      this.showModifyPasswordDialog = false;
      this.$nextTick(() => {
        this.editedPasswordItem = { ...this.defaultPasswordItem };
      });
    },
    openPasswordForm() {
      this.showModifyPasswordDialog = true;
    },
    validatePasswordForm() {
      return this.$refs.modifyPasswordForm.validate();
    },
    savePasswordForm() {
      if (!this.validatePasswordForm()) return;

      doPost(ROUTE_USER_MODIFY, this.$router, {
        uid: this.userData.uid,
        new_password: this.editedPasswordItem.new_password,
        current_password: this.editedEmailItem.current_password,
      })
        .then((data) => {
          if (data.user) {
            localStorage.setItem(LOCALSTORAGE_FIELD_USER, JSON.stringify(data.user));
            this.userData = data.user;
          }
        })
        .catch((e) => {
          handleError(this.$store, e);
        })
        .finally(() => {
          this.closePasswordForm();
        });
    },
  },
  mounted() {
    document.title = `Profile - ${PAGE_TITLE}`;
  },
};
</script>

<style scoped>

</style>
