<template>
  <v-container>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <v-data-table
            :headers="headers"
            :items="this.users?this.users:[]"
            :search="search"
            :loading="this.isLoading || !this.users"
            item-key="uid"
            sort-by="modified"
            sort-desc
            class="elevation-4">
          <template v-slot:top>
            <v-toolbar flat class="hidden-sm-and-up">
              <v-toolbar-title>User Administration</v-toolbar-title>
            </v-toolbar>
            <v-toolbar flat class="hidden-sm-and-up">
              <v-text-field
                  v-model="search"
                  :append-icon="icons.mdiMagnify"
                  label="Filter users"
                  single-line
                  hide-details
                  style="max-width: 500px"
              ></v-text-field>
              <v-spacer></v-spacer>
              <v-btn outlined @click="refreshUsers" :disabled="isLoading">
                <v-icon>{{ icons.mdiRefresh }}</v-icon>
              </v-btn>
            </v-toolbar>
            <v-toolbar flat class="hidden-xs-only">
              <v-toolbar-title>User Administration</v-toolbar-title>
              <v-spacer></v-spacer>
              <v-text-field
                  v-model="search"
                  :append-icon="icons.mdiMagnify"
                  label="Filter users"
                  single-line
                  hide-details
                  style="max-width: 500px"
              ></v-text-field>
              <v-spacer></v-spacer>
              <v-btn outlined @click="refreshUsers" :disabled="isLoading">
                <v-icon>{{ icons.mdiRefresh }}</v-icon>
                <div class="ml-2 hidden-sm-and-down">Refresh</div>
              </v-btn>
            </v-toolbar>
          </template>
          <template v-slot:[`item.created`]="{ item }">
            <span>{{ new Date(item.created).toLocaleString() }}</span>
          </template>
          <template v-slot:[`item.modified`]="{ item }">
            <span>{{ new Date(item.modified).toLocaleString() }}</span>
          </template>
        </v-data-table>
        <v-data-table
            :headers="identityHeaders"
            :items="this.identities?this.identities:[]"
            :search="search"
            :loading="this.isLoading || !this.identities"
            item-key="id"
            sort-by="modified"
            sort-desc
            class="elevation-4 mt-2">
          <template v-slot:top>
            <v-toolbar flat class="hidden-sm-and-up">
              <v-toolbar-title>Identities</v-toolbar-title>
            </v-toolbar>
            <v-toolbar flat class="hidden-sm-and-up">
              <v-text-field
                  v-model="search"
                  :append-icon="icons.mdiMagnify"
                  label="Filter users"
                  single-line
                  hide-details
                  style="max-width: 500px"
              ></v-text-field>
              <v-spacer></v-spacer>
              <v-btn outlined @click="refreshUsers" :disabled="isLoading">
                <v-icon>{{ icons.mdiRefresh }}</v-icon>
              </v-btn>
              <v-btn
                  outlined
                  class="ml-1"
                  @click="showCreateDialog">
                <v-icon>{{ icons.mdiAccountPlus }}</v-icon>
                <div class="ml-2 hidden-sm-and-down">Create Identity</div>
              </v-btn>
            </v-toolbar>
            <v-toolbar flat class="hidden-xs-only">
              <v-toolbar-title>User Administration</v-toolbar-title>
              <v-spacer></v-spacer>
              <v-text-field
                  v-model="search"
                  :append-icon="icons.mdiMagnify"
                  label="Filter users"
                  single-line
                  hide-details
                  style="max-width: 500px"
              ></v-text-field>
              <v-spacer></v-spacer>
              <v-btn outlined @click="refreshUsers" :disabled="isLoading">
                <v-icon>{{ icons.mdiRefresh }}</v-icon>
                <div class="ml-2 hidden-sm-and-down">Refresh</div>
              </v-btn>
              <v-btn
                  outlined
                  class="ml-1"
                  @click="showCreateDialog">
                <v-icon>{{ icons.mdiAccountPlus }}</v-icon>
                <div class="ml-2 hidden-sm-and-down">Create Identity</div>
              </v-btn>
            </v-toolbar>
          </template>
          <template v-slot:[`item.actions`]="{ item }">
            <v-icon
                small
                class="mr-2"
                :disabled="isLoading"
                @click="editIdentity(item)">
              {{ icons.mdiPencil }}
            </v-icon>
            <v-icon
                small
                class="mr-2"
                :disabled="isLoading"
                @click="showDeleteDialog(item)">
              {{ icons.mdiDelete }}
            </v-icon>
            <v-icon
                small
                :disabled="isLoading"
                @click="showPropertyDialog(item)">
              {{ icons.mdiUnfoldMoreVertical }}
            </v-icon>
          </template>
          <template v-slot:[`item.created_at`]="{ item }">
            <span>{{ new Date(item.created_at).toLocaleString() }}</span>
          </template>
          <template v-slot:[`item.updated_at`]="{ item }">
            <span>{{ new Date(item.updated_at).toLocaleString() }}</span>
          </template>
        </v-data-table>
        <v-dialog v-model="showCreateIdentityDialog" max-width="500px">
          <v-card>
            <v-card-title>
              <span class="text-h5">{{ formTitle }}</span>
            </v-card-title>
            <v-card-text>
              <v-container>
                <v-row>
                  <v-form ref="modifyIdentityForm">
                    <v-text-field
                        v-model="editedItem.email"
                        label="E-mail"
                        type="email"
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
              <v-btn color="blue darken-1" text @click="saveIdentity">Save</v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
        <v-dialog
            v-model="showDeleteIdentityDialog"
            max-width="500px"
            v-if="this.identityToDelete">
          <v-card>
            <v-card-title>
              <span class="text-h5">Delete Identity</span>
            </v-card-title>
            <v-card-text>
              <p class="font-weight-black text-body-1 my-0">
                Do you really want to delete the identity?</p>
              <p class="font-weight-black text-body-1 my-0">
                ID: {{ this.identityToDelete.id }} </p>
              <p class="font-weight-black text-body-1">
                E-mail: {{ this.identityToDelete.email }} </p>
            </v-card-text>
            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn color="blue darken-1" text @click="closeDeletionDialog">Cancel</v-btn>
              <v-btn
                  color="blue darken-1"
                  text
                  @click="deleteIdentity(identityToDelete)">Yes, delete identity
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
        <v-dialog
            v-model="showIdentityPropertyDialog"
            max-width="700px">
          <v-card>
            <v-card-title>Identity Properties</v-card-title>
            <v-card-text>
              <v-textarea :value="identityPropertyDialogData" auto-grow readonly outlined/>
            </v-card-text>
          </v-card>
        </v-dialog>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import {
  mdiPencil, mdiDelete, mdiRefresh, mdiAccountPlus, mdiMagnify, mdiUnfoldMoreVertical,
} from '@mdi/js';
import {
  PAGE_TITLE, ROUTE_IDENTITY_LIST, ROUTE_IDENTITY_CREATE,
  ROUTE_IDENTITY_MODIFY, ROUTE_IDENTITY_DELETE,
} from '../../constants';
import {
  emailRules, doGet, doPost, handleError,
} from '../../utilities';

export default {
  name: 'Administration',
  data: () => ({
    icons: {
      mdiPencil, mdiDelete, mdiRefresh, mdiAccountPlus, mdiMagnify, mdiUnfoldMoreVertical,
    },
    isLoading: false,
    showCreateIdentityDialog: false,
    showDeleteIdentityDialog: false,
    showIdentityPropertyDialog: false,
    identityToDelete: null,
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
    ],
    identityHeaders: [
      {
        text: 'ID', value: 'id', align: 'start', sortable: false,
      },
      {
        text: 'E-Mail', value: 'email',
      },
      {
        text: 'State', value: 'state',
      },
      {
        text: 'Schema ID', value: 'schema_id',
      },
      {
        text: 'Roles', value: 'roles',
      },
      {
        text: 'Created', value: 'created_at',
      },
      {
        text: 'Updated', value: 'updated_at',
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
      uid: '',
      email: '',
      roles: [],
    },
    defaultItem: {
      uid: '',
      email: '',
      roles: [],
    },
    users: null,
    identities: null,
    identityPropertyDialogData: null,
  }),
  computed: {
    formTitle() {
      return this.editedIndex === -1 ? 'Create Identity' : 'Edit Identity';
    },
  },
  watch: {
    dialog(val) {
      if (!val) this.close();
    },
  },
  methods: {
    setErrorMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'error', temporary: true });
    },
    loadUserList() {
      return doGet(ROUTE_IDENTITY_LIST, this.$router, this.$store).then((data) => {
        if (!data.success) throw Error('error getting user data');
        this.users = data.users;
        this.identities = data.identities;
        this.$store.dispatch('resetMessages');
      }).catch((e) => {
        handleError(this.$store, e);
        return e;
      });
    },
    async refreshUsers() {
      this.isLoading = true;
      await this.loadUserList();
      this.isLoading = false;
      this.search = '';
      if (!this.users || !this.identities) return;

      this.users = this.users.map((d) => {
        // convert dates to unix time so, they can be sorted in data table
        d.modified = new Date(d.modified).getTime();
        d.created = new Date(d.created).getTime();
        d.roles = d.roles.map((f) => f.name);
        return d;
      });

      this.identities = this.identities.map((d) => {
        // convert dates to unix time so, they can be sorted in data table
        d.updated_at = new Date(d.updated_at).getTime();
        d.created_at = new Date(d.created_at).getTime();
        d.email = d.traits.email;

        if (d.metadata_public && d.metadata_public.roles) {
          d.roles = d.metadata_public.roles.map((f) => f);
        }

        return d;
      });
    },
    editIdentity(item) {
      this.editedIndex = this.identities.indexOf(item);
      this.editedItem = { ...item };
      this.showCreateIdentityDialog = true;
    },
    showCreateDialog() {
      this.editedIndex = -1;
      this.editedItem = { ...this.defaultItem };
      this.showCreateIdentityDialog = true;
    },
    showDeleteDialog(identity) {
      this.showDeleteIdentityDialog = true;
      this.identityToDelete = identity;
    },
    showPropertyDialog(identity) {
      this.showIdentityPropertyDialog = true;
      this.identityPropertyDialogData = JSON.stringify(identity, null, '\t');
    },
    deleteIdentity(identity) {
      this.isLoading = true;

      doGet(ROUTE_IDENTITY_DELETE, this.$router, this.$store, identity.id)
        .then((data) => {
          if (data.success === undefined) throw Error('error deleting identity');
          if (data.success === false) {
            throw Error(data.msg);
          }
          this.refreshUsers();
        })
        .catch((error) => {
          this.setErrorMessage(error);
        })
        .finally(() => {
          this.isLoading = false;
          this.closeDeletionDialog();
        });
    },
    close() {
      this.showCreateIdentityDialog = false;
      this.$nextTick(() => {
        this.editedItem = { ...this.defaultItem };
        this.editedIndex = -1;
      });
    },
    closeDeletionDialog() {
      this.showDeleteIdentityDialog = false;
      this.identityToDelete = null;
    },
    validateForm() {
      if (this.$refs.modifyUserForm) return this.$refs.modifyUserForm.validate();

      return this.$refs.modifyIdentityForm.validate();
    },
    saveIdentity() {
      if (!this.validateForm()) return;

      if (this.editedIndex > -1) {
        this.isLoading = true;
        doPost(ROUTE_IDENTITY_MODIFY, this.$router, this.$store, {
          uid: this.editedItem.id,
          email: this.editedItem.email,
          roles: this.editedItem.roles.map((d) => ({ name: d })),
        })
          .then((data) => {
            if (data.success === undefined) throw Error('error modifying password');
            if (data.success === false) throw new Error(data.msg);

            this.refreshUsers();
          })
          .catch((e) => {
            handleError(this.$store, e);
          })
          .finally(() => {
            this.isLoading = false;
            this.close();
          });
      } else {
        this.isLoading = true;

        doPost(ROUTE_IDENTITY_CREATE, this.$router, this.$store, this.editedItem)
          .then((data) => {
            if (data.success === undefined) throw Error('error creating identity');
            if (data.success === false) {
              throw Error(data.msg);
            }
            this.refreshUsers();
          })
          .catch((error) => {
            this.setErrorMessage(error);
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
