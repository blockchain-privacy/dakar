<template>
  <v-row align="center" justify="center">
    <v-col cols="12" sm="12" md="10" lg="9" xl="8">
      <v-data-table
          :headers="headers"
          :items="this.users?this.users:[]"
          :search="search"
          :loading="this.isLoading || !this.users"
          item-key="uid"
          sort-by="user_modified"
          sort-desc
          class="elevation-1">
        <template v-slot:top>
          <v-toolbar flat class="hidden-sm-and-up">
            <v-toolbar-title>User Administration</v-toolbar-title>
          </v-toolbar>
          <v-toolbar flat class="hidden-sm-and-up">
            <v-text-field
                v-model="search"
                append-icon="mdi-magnify"
                label="Filter users"
                single-line
                hide-details
                style="max-width: 500px"
            ></v-text-field>
            <v-spacer></v-spacer>
            <v-btn outlined @click="refreshUsers" :disabled="isLoading">
              <v-icon>{{ icon.mdiRefresh }}</v-icon>
            </v-btn>
            <v-btn
                outlined
                class="ml-1"
                @click.stop="showCreateUserDialog = true">
              <v-icon>{{ icon.mdiAccountPlus }}</v-icon>
              <div class="ml-2 hidden-sm-and-down">Create User</div>
            </v-btn>
          </v-toolbar>
          <v-toolbar flat class="hidden-xs-only">
            <v-toolbar-title>User Administration</v-toolbar-title>
            <v-spacer></v-spacer>
            <v-text-field
                v-model="search"
                append-icon="mdi-magnify"
                label="Filter users"
                single-line
                hide-details
                style="max-width: 500px"
            ></v-text-field>
            <v-spacer></v-spacer>
            <v-btn outlined @click="refreshUsers" :disabled="isLoading">
              <v-icon>{{ icon.mdiRefresh }}</v-icon>
              <div class="ml-2 hidden-sm-and-down">Refresh</div>
            </v-btn>
            <v-btn
                outlined
                class="ml-1"
                @click.stop="showCreateUserDialog = true">
              <v-icon>{{ icon.mdiAccountPlus }}</v-icon>
              <div class="ml-2 hidden-sm-and-down">Create User</div>
            </v-btn>
          </v-toolbar>
        </template>
        <template v-slot:[`item.actions`]="{ item }">
          <v-icon
              small
              class="mr-2"
              :disabled="isLoading"
              @click="editItem(item)">
            {{ icon.mdiPencil }}
          </v-icon>
          <v-icon
              small
              :disabled="isLoading"
              @click="showDeleteDialog(item)">
            {{ icon.mdiDelete }}
          </v-icon>
        </template>
        <template v-slot:[`item.created`]="{ item }">
          <span>{{ item.created.toLocaleString() }}</span>
        </template>
        <template v-slot:[`item.modified`]="{ item }">
          <span>{{ item.modified.toLocaleString() }}</span>
        </template>
      </v-data-table>
      <v-dialog v-model="showCreateUserDialog" max-width="500px">
        <v-card>
          <v-card-title>
            <span class="headline">{{ formTitle }}</span>
          </v-card-title>
          <v-card-text>
            <v-container>
              <v-row>
                <v-form ref="modifyUserForm">
                  <v-text-field
                      v-model="editedItem.user_email"
                      label="E-mail"
                      type="email"
                      :rules="rules.emailRules">
                  </v-text-field>
                  <v-select
                      :rules="rules.roleRules"
                      :items="roles"
                      label="Roles"
                      multiple
                      v-model="editedItem.user_roles"/>
                </v-form>
              </v-row>
            </v-container>
          </v-card-text>
          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn color="blue darken-1" text @click="close">Cancel</v-btn>
            <v-btn color="blue darken-1" text @click="save">Save</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
      <v-dialog
          v-model="showDeleteUserDialog"
          max-width="500px"
          v-if="this.userToDelete">
        <v-card>
          <v-card-title>
            <span class="headline">Delete User</span>
          </v-card-title>
          <v-card-text>
            <p class="font-weight-black body-1 my-0">Do you really want to delete the user?</p>
            <p class="font-weight-black body-1 my-0"> Uid: {{ this.userToDelete.uid }} </p>
            <p class="font-weight-black body-1"> E-mail: {{ this.userToDelete.user_email }} </p>
          </v-card-text>
          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn color="blue darken-1" text @click="closeDeletionDialog">Cancel</v-btn>
            <v-btn
                color="blue darken-1"
                text
                @click="deleteItem(userToDelete)">Yes, delete user
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-col>
  </v-row>
</template>

<script>
import {
  mdiPencil, mdiDelete, mdiRefresh, mdiAccountPlus,
} from '@mdi/js';
import { PAGE_TITLE, ROUTE_USER_CREATE, ROUTE_USER_DELETE } from '../../constants';
import { emailRules, isInvalidTokenMsg, doGet } from '../../utilities';

export default {
  name: 'Administration',
  data: () => ({
    icon: {
      mdiPencil, mdiDelete, mdiRefresh, mdiAccountPlus,
    },
    isLoading: false,
    showCreateUserDialog: false,
    showDeleteUserDialog: false,
    userToDelete: null,
    search: '',
    headers: [
      {
        text: 'ID', value: 'uid', align: 'start', sortable: false,
      },
      {
        text: 'E-Mail', value: 'user_email',
      },
      {
        text: 'Roles', value: 'user_roles',
      },
      {
        text: 'Created', value: 'user_created',
      },
      {
        text: 'Modified', value: 'user_modified',
      },
      {
        text: 'Actions', value: 'actions', sortable: false, align: 'end',
      },
    ],
    rules: {
      roleRules: [
        (v) => v.length > 0 || 'At least one role is required',
      ],
      emailRules,
    },
    roles: ['admin', 'user', 'privileged'],
    editedIndex: -1,
    editedItem: {
      user_email: '',
      user_roles: [],
    },
    defaultItem: {
      user_email: '',
      user_roles: [],
    },
  }),
  computed: {
    formTitle() {
      return this.editedIndex === -1 ? 'Create User' : 'Edit User';
    },
    users: {
      get() {
        return this.$store.getters.getUserList;
      },
      set(value) {
        this.$store.dispatch('setUserList', value);
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
  },
  watch: {
    dialog(val) {
      if (!val) this.close();
    },
  },
  methods: {
    async refreshUsers() {
      this.isLoading = true;
      await this.$store.dispatch('updateUserList');
      this.isLoading = false;
      this.search = '';
      if (!this.users) return;

      this.users = this.users.map((d) => {
        // parse data to readable format
        d.user_modified = new Date(d.user_modified).toLocaleString();
        d.user_created = new Date(d.user_created).toLocaleString();
        d.user_roles = d.user_roles.map((f) => f.role_name);
        return d;
      });
    },
    editItem(item) {
      this.editedIndex = this.users.indexOf(item);
      this.editedItem = { ...item };
      this.showCreateUserDialog = true;
    },
    showDeleteDialog(user) {
      this.showDeleteUserDialog = true;
      this.userToDelete = user;
    },
    deleteItem(user) {
      this.isLoading = true;

      doGet(ROUTE_USER_DELETE, user.uid, this.$router)
        .then((data) => {
          if (data.success === undefined) throw Error('error deleting user');
          if (data.success === false) {
            throw Error(data.msg);
          }
          this.refreshUsers();
        })
        .catch((error) => {
          this.errMsg = error;
        })
        .finally(() => {
          this.isLoading = false;
          this.closeDeletionDialog();
        });
    },
    close() {
      this.showCreateUserDialog = false;
      this.$nextTick(() => {
        this.editedItem = { ...this.defaultItem };
        this.editedIndex = -1;
      });
    },
    closeDeletionDialog() {
      this.showDeleteUserDialog = false;
      this.userToDelete = null;
    },
    validateForm() {
      return this.$refs.modifyUserForm.validate();
    },
    save() {
      if (!this.validateForm()) return;

      if (this.editedIndex > -1) {
        Object.assign(this.users[this.editedIndex], this.editedItem);
        this.close();
      } else {
        this.isLoading = true;
        fetch(ROUTE_USER_CREATE, {
          method: 'POST', // or 'PUT'
          credentials: 'same-origin',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(this.editedItem),
        })
          .then((response) => response.json())
          .then((data) => {
            if (isInvalidTokenMsg(data, this.$router)) return;

            if (data.success === undefined) throw Error('error creating user');
            if (data.success === false) {
              throw Error(data.msg);
            }
            this.refreshUsers();
          })
          .catch((error) => {
            this.errMsg = error;
          })
          .finally(() => {
            this.isLoading = false;
            this.close();
          });
      }
    },
  },
  mounted() {
    document.title = `User Administration - ${PAGE_TITLE}`;
    this.refreshUsers();
  },
};
</script>

<style scoped>

</style>
