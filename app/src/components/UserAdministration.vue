<template>

  <v-row align="center" justify="center">
    <v-col cols="12" sm="12" md="10" lg="9" xl="8">
      <v-data-table
          :headers="headers"
          :items="users"
          :search="search"
          item-key="uid"
          sort-by="email"
          class="elevation-1">
        <template v-slot:top>
          <v-toolbar flat>
            <v-toolbar-title>User Administration</v-toolbar-title>
            <v-spacer></v-spacer>
            <v-text-field
                v-model="search"
                append-icon="mdi-magnify"
                label="Search table"
                single-line
                hide-details
                style="max-width: 500px"
            ></v-text-field>
            <v-spacer></v-spacer>
            <v-dialog v-model="dialog" max-width="500px">
              <template v-slot:activator="{ on, attrs }">
                <v-btn
                    outlined
                    class="mb-2"
                    v-bind="attrs"
                    v-on="on">
                  Create User
                </v-btn>
              </template>
              <v-card>
                <v-card-title>
                  <span class="headline">{{ formTitle }}</span>
                </v-card-title>
                <v-card-text>
                  <v-container>
                    <v-row>
                      <v-form ref="modifyUserForm">
                        <v-text-field
                            v-model="editedItem.email"
                            label="E-mail"
                            type="text"
                            :rules="rules.emailRules">
                        </v-text-field>
                        <v-select
                            :rules="rules.roleRules"
                            :items="roles"
                            label="Roles"
                            multiple
                            v-model="editedItem.roles"/>
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
          </v-toolbar>
        </template>
        <template v-slot:[`item.actions`]="{ item }">
          <v-icon
              small
              class="mr-2"
              @click="editItem(item)">
            {{ icon.mdiPencil }}
          </v-icon>
          <v-icon
              small
              @click="deleteItem(item)">
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
    </v-col>
  </v-row>

</template>

<script>
import {
  mdiPencil, mdiDelete,
} from '@mdi/js';
import { emailRules } from '../utilities';

export default {
  name: 'UserAdministration',
  data: () => ({
    icon: {
      mdiPencil, mdiDelete,
    },
    dialog: false,
    search: '',
    headers: [
      {
        text: 'ID', value: 'uid', align: 'start', sortable: false,
      },
      {
        text: 'E-Mail', value: 'email',
      },
      {
        text: 'Roles', value: 'roles',
      },
      {
        text: 'Created', value: 'created',
      },
      {
        text: 'Modified', value: 'modified',
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
    users: [],
    editedIndex: -1,
    editedItem: {
      uid: '',
      email: '',
      roles: [],
      created: null,
      modified: null,
    },
    defaultItem: {
      uid: '< not set >',
      email: '',
      roles: [],
      created: null,
      modified: null,
    },
  }),
  computed: {
    formTitle() {
      return this.editedIndex === -1 ? 'Create User' : 'Edit User';
    },
  },
  watch: {
    dialog(val) {
      if (!val) this.close();
    },
  },
  created() {
    this.initialize();
  },
  methods: {
    initialize() {
      this.users = [
        {
          uid: '0x1',
          email: 'admin@example.com',
          roles: ['admin'],
          created: new Date(),
          modified: new Date(),
        },
        {
          uid: '0x2',
          email: 'user1@example.com',
          roles: ['user'],
          created: new Date(),
          modified: new Date(),
        },
        {
          uid: '0x3',
          email: 'user2@example.com',
          roles: ['user', 'privileged'],
          created: new Date(),
          modified: new Date(),
        },
        {
          uid: '0x4',
          email: 'user3@example.com',
          roles: ['user'],
          created: new Date(),
          modified: new Date(),
        },
      ];
    },
    editItem(item) {
      this.editedIndex = this.users.indexOf(item);
      this.editedItem = { ...item };
      this.dialog = true;
    },
    deleteItem(item) {
      const index = this.users.indexOf(item);
      // eslint-disable-next-line no-restricted-globals
      if (confirm('Are you sure you want to delete this item?')) this.users.splice(index, 1);
    },
    close() {
      this.dialog = false;
      this.$nextTick(() => {
        this.editedItem = { ...this.defaultItem };
        this.editedIndex = -1;
      });
    },
    validateForm() {
      return this.$refs.modifyUserForm.validate();
    },
    save() {
      if (!this.validateForm()) return;

      if (this.editedIndex > -1) {
        Object.assign(this.users[this.editedIndex], this.editedItem);
      } else {
        this.uid = 'not set yet';
        this.editedItem.modified = new Date();
        this.editedItem.created = new Date();
        this.users.push(this.editedItem);
      }
      this.close();
    },
  },
};
</script>

<style scoped>

</style>
