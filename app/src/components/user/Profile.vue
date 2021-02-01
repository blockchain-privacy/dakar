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
                    :rules="rules.pwRequired">
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
                <v-row>
                  <v-col cols="6">
                    <v-text-field
                        v-model="editedPasswordItem.new_password"
                        label="New password"
                        type="password"
                        :rules="rules.passwordRules">
                    </v-text-field>
                  </v-col>
                  <v-col cols="6">
                    <v-text-field
                        v-model="editedPasswordItem.new_password_confirm"
                        label="Confirm new password"
                        type="password"
                        :rules="[(editedPasswordItem.new_password
                    === editedPasswordItem.new_password_confirm) || 'Password must match']">
                    </v-text-field>
                  </v-col>
                </v-row>
                <v-row>
                  <v-col cols="6">
                    <v-text-field
                        v-model="editedPasswordItem.current_password"
                        label="Current password"
                        type="password"
                        :rules="rules.pwRequired">
                    </v-text-field>
                  </v-col>
                </v-row>

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
import { PAGE_TITLE, ROUTE_USER_MODIFY } from '../../constants';
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
        emailRules,
        passwordRules,
        pwRequired: [(v) => !!v || 'Password is required'],
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
        new_password_confirm: '',
        current_password: '',
      },
      defaultPasswordItem: {
        new_password: '',
        new_password_confirm: '',
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
    successMsg: {
      get() {
        return this.$store.getters.getSuccessMsg;
      },
      set(value) {
        this.$store.dispatch('setSuccessMsg', value);
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
          val: '••••••••••••••••••',
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
      this.$store.dispatch('resetMsg');

      doPost(ROUTE_USER_MODIFY, this.$router, {
        uid: this.userData.uid,
        email: this.editedEmailItem.email,
        current_password: this.editedEmailItem.current_password,
      })
        .then((data) => {
          if (data.success === undefined) throw Error('error modifying e-mail');
          if (data.success === false) {
            throw new Error(data.msg);
          }

          if (data.user) {
            this.userData = data.user;
            this.successMsg = 'Successfully changed E-mail';
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
      this.$store.dispatch('resetMsg');

      doPost(ROUTE_USER_MODIFY, this.$router, {
        uid: this.userData.uid,
        new_password: this.editedPasswordItem.new_password,
        current_password: this.editedPasswordItem.current_password,
      })
        .then((data) => {
          if (data.success === undefined) throw Error('error modifying password');
          if (data.success === false) {
            throw new Error(data.msg);
          }

          if (data.user) {
            this.userData = data.user;
            this.successMsg = 'Successfully changed password';
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
