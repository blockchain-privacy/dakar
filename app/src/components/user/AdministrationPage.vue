<template>
  <v-container>
    <v-row
      align="center"
      justify="center"
    >
      <v-col
        cols="12"
        sm="12"
        md="10"
        lg="9"
        xl="8"
      >
        <v-data-table
          v-model:sort-by="usersSortBy"
          :headers="headers"
          :items="users?users:[]"
          :search="search"
          :loading="isLoading || !users"
          item-value="uid"
          class="elevation-2"
        >
          <template #top>
            <v-toolbar
              :flat="true"
              class="hidden-sm-and-up"
            >
              <v-toolbar-title>User Administration</v-toolbar-title>
            </v-toolbar>
            <v-toolbar
              :flat="true"
              class="hidden-sm-and-up"
            >
              <v-text-field
                v-model="search"
                :append-inner-icon="icons.mdiMagnify"
                label="Filter users"
                single-line
                hide-details
                style="max-width: 500px"
              />
              <v-spacer />
              <v-btn
                variant="outlined"
                :disabled="isLoading"
                @click="refreshUsers"
              >
                <v-icon>{{ icons.mdiRefresh }}</v-icon>
              </v-btn>
            </v-toolbar>
            <v-toolbar
              :flat="true"
              class="d-none d-sm-flex"
            >
              <v-toolbar-title>User Administration</v-toolbar-title>
              <v-spacer />
              <v-text-field
                v-model="search"
                :append-inner-icon="icons.mdiMagnify"
                label="Filter users"
                single-line
                hide-details
                style="max-width: 500px"
              />
              <v-spacer />
              <v-btn
                variant="outlined"
                :disabled="isLoading"
                @click="refreshUsers"
              >
                <v-icon>{{ icons.mdiRefresh }}</v-icon>
                <div class="ml-2 hidden-sm-and-down">
                  Refresh
                </div>
              </v-btn>
            </v-toolbar>
          </template>
          <template #item.created="{ item }">
            <span>{{ new Date(item.created).toLocaleString() }}</span>
          </template>
          <template #item.modified="{ item }">
            <span>{{ new Date(item.modified).toLocaleString() }}</span>
          </template>
        </v-data-table>
        <v-data-table
          v-model:sort-by="identitiesSortBy"
          :headers="identityHeaders"
          :items="identities?identities:[]"
          :search="search"
          :loading="isLoading || !identities"
          item-key="id"
          class="elevation-4 mt-2"
        >
          <template #top>
            <v-toolbar
              :flat="true"
              class="hidden-sm-and-up"
            >
              <v-toolbar-title>Identities</v-toolbar-title>
            </v-toolbar>
            <v-toolbar
              :flat="true"
              class="hidden-sm-and-up"
            >
              <v-text-field
                v-model="search"
                :append-inner-icon="icons.mdiMagnify"
                label="Filter users"
                single-line
                hide-details
                style="max-width: 500px"
              />
              <v-spacer />
              <v-btn
                variant="outlined"
                :disabled="isLoading"
                @click="refreshUsers"
              >
                <v-icon>{{ icons.mdiRefresh }}</v-icon>
              </v-btn>
              <v-btn
                variant="outlined"
                class="ml-1"
                @click="showCreateDialog"
              >
                <v-icon>{{ icons.mdiAccountPlus }}</v-icon>
                <div class="ml-2 hidden-sm-and-down">
                  Create Identity
                </div>
              </v-btn>
            </v-toolbar>
            <v-toolbar
              :flat="true"
              class="d-none d-sm-flex"
            >
              <v-toolbar-title>Identities</v-toolbar-title>
              <v-spacer />
              <v-text-field
                v-model="search"
                :append-inner-icon="icons.mdiMagnify"
                label="Filter users"
                single-line
                hide-details
                style="max-width: 500px"
              />
              <v-spacer />
              <v-btn
                variant="outlined"
                :disabled="isLoading"
                @click="refreshUsers"
              >
                <v-icon>{{ icons.mdiRefresh }}</v-icon>
                <div class="ml-2 hidden-sm-and-down">
                  Refresh
                </div>
              </v-btn>
              <v-btn
                variant="outlined"
                class="ml-1"
                @click="showCreateDialog"
              >
                <v-icon>{{ icons.mdiAccountPlus }}</v-icon>
                <div class="ml-2 hidden-sm-and-down">
                  Create Identity
                </div>
              </v-btn>
            </v-toolbar>
          </template>
          <template #item.actions="{ item }">
            <v-menu>
              <template #activator="{ props }">
                <v-btn
                  v-bind="props"
                  icon
                  variant="text"
                >
                  <v-icon>{{ icons.mdiDotsVertical }}</v-icon>
                </v-btn>
              </template>
              <v-list>
                <v-list-item @click="showEditDialog(item)">
                  <template #prepend>
                    <v-icon :icon="icons.mdiPencil" />
                  </template>
                  Edit
                </v-list-item>
                <v-list-item @click="showPropertyDialog(item)">
                  <template #prepend>
                    <v-icon :icon="icons.mdiUnfoldMoreVertical" />
                  </template>
                  Details
                </v-list-item>
                <v-list-item @click="showDeleteDialog(item)">
                  <template #prepend>
                    <v-icon :icon="icons.mdiDelete" />
                  </template>
                  Delete
                </v-list-item>
              </v-list>
            </v-menu>
          </template>
          <template #item.createdAt="{ item }">
            <span>{{ new Date(item.createdAt).toLocaleString() }}</span>
          </template>
          <template #item.updatedAt="{ item }">
            <span>{{ new Date(item.updatedAt).toLocaleString() }}</span>
          </template>
        </v-data-table>
        <v-data-table
          v-model:sort-by="sessionsSortBy"
          :headers="sessionHeaders"
          :items="sessions?sessions:[]"
          :search="searchSessions"
          :loading="isLoading || !sessions"
          item-key="id"
          class="elevation-4 mt-2"
        >
          <template #top>
            <v-toolbar
              :flat="true"
              class="hidden-sm-and-up"
            >
              <v-toolbar-title>Sessions</v-toolbar-title>
            </v-toolbar>
            <v-toolbar
              :flat="true"
              class="hidden-sm-and-up"
            >
              <v-text-field
                v-model="searchSessions"
                :append-inner-icon="icons.mdiMagnify"
                label="Filter sessions"
                single-line
                hide-details
                style="max-width: 500px"
              />
            </v-toolbar>
            <v-toolbar
              :flat="true"
              class="d-none d-sm-flex"
            >
              <v-toolbar-title>Sessions</v-toolbar-title>
              <v-spacer />
              <v-text-field
                v-model="searchSessions"
                :append-inner-icon="icons.mdiMagnify"
                label="Filter sessions"
                single-line
                hide-details
                style="max-width: 500px"
              />
            </v-toolbar>
          </template>
          <template #item.authenticated_at="{ item }">
            <span>{{ new Date(item.authenticated_at).toLocaleString() }}</span>
          </template>
          <template #item.expires_at="{ item }">
            <span>{{ new Date(item.expires_at).toLocaleString() }}</span>
          </template>
        </v-data-table>
        <edit-identity-dialog
          v-if="showCreateIdentityDialog"
          v-model="showCreateIdentityDialog"
          :create-new-user="createNewUser"
          :identity="editedItem"
          @saved="refreshUsers()"
        />
        <v-dialog
          v-if="identityToDelete"
          v-model="showDeleteIdentityDialog"
          max-width="500px"
        >
          <v-card>
            <v-card-title>
              <span class="text-h5">Delete Identity</span>
            </v-card-title>
            <v-card-text>
              <p class="text-subtitle-1">
                Do you really want to delete this identity?
              </p>
              <p class="text-subtitle-1">
                ID: {{ identityToDelete.id }}
              </p>
              <p class="text-subtitle-1">
                E-mail: {{ identityToDelete.email }}
              </p>
            </v-card-text>
            <v-card-actions>
              <v-spacer />
              <v-btn @click="closeDeletionDialog">
                Cancel
              </v-btn>
              <v-btn
                color="red"
                @click="deleteIdentity(identityToDelete)"
              >
                Delete
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
        <v-dialog
          v-model="showIdentityPropertyDialog"
          max-width="700px"
        >
          <v-card>
            <v-card-title>Identity Properties</v-card-title>
            <v-textarea
              :model-value="identityPropertyDialogData"
              auto-grow
              :readonly="true"
              hide-details
            />
          </v-card>
        </v-dialog>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import {
	mdiPencil, mdiDelete, mdiRefresh, mdiAccountPlus,
	mdiMagnify, mdiUnfoldMoreVertical, mdiDotsVertical,
} from '@mdi/js';
import {PAGE_TITLE} from '@/constants';
import {handleError} from '@/utilities';
import EditIdentityDialog from '@/components/user/EditIdentityDialog.vue';

export default {
	name: 'AdministrationPage',
	components: {EditIdentityDialog},
	data: () => ({
		icons: {
			mdiPencil, mdiDelete, mdiRefresh, mdiAccountPlus,
			mdiMagnify, mdiUnfoldMoreVertical, mdiDotsVertical,
		},
		isLoading: false,
		showCreateIdentityDialog: false,
		showDeleteIdentityDialog: false,
		showIdentityPropertyDialog: false,
		identityToDelete: null,
		search: '',
		searchSessions: '',
		usersSortBy: [{key: 'modified', order: 'desc'}],
		headers: [
			{
				title: 'ID', key: 'uid', align: 'start', sortable: false,
			},
			{
				title: 'E-Mail', key: 'email',
			},
			{
				title: 'Roles', key: 'renderedRoles',
			},
			{
				title: 'Created', key: 'created',
			},
			{
				title: 'Modified', key: 'modified',
			},
		],
		identitiesSortBy: [{key: 'modified', order: 'desc'}],
		identityHeaders: [
			{
				title: 'ID', key: 'id', align: 'start', sortable: false,
			},
			{
				title: 'Dgraph UID', key: 'dgraphUID',
			},
			{
				title: 'E-Mail', key: 'email',
			},
			{
				title: 'State', key: 'state',
			},
			{
				title: 'Schema ID', key: 'schema_id',
			},
			{
				title: 'Roles', key: 'renderedRoles',
			},
			{
				title: 'Created', key: 'createdAt',
			},
			{
				title: 'Updated', key: 'updatedAt',
			},
			{
				title: '', key: 'actions', sortable: false, align: 'end',
			},
		],
		sessionsSortBy: [{key: 'authenticated_at', order: 'desc'}],
		sessionHeaders: [
			{
				title: 'ID', key: 'id', align: 'start', sortable: false,
			},
			{
				title: 'E-Mail', key: 'identity.traits.email',
			},
			{
				title: 'Active', key: 'active',
			},
			{
				title: 'Authentication Date', key: 'authenticated_at',
			},
			{
				title: 'Expiry Date', key: 'expires_at',
			},
		],
		createNewUser: false,
		editedItem: {
			id: '',
			email: '',
			state: '',
			roles: [],
		},
		defaultItem: {
			id: '',
			email: '',
			state: '',
			roles: [],
		},
		users: null,
		identities: null,
		sessions: null,
		identityPropertyDialogData: null,
	}),
	computed: {
		formTitle() {
			return this.createNewUser ? 'Create Identity' : 'Edit Identity';
		},
	},
	mounted() {
		document.title = `User Administration - ${PAGE_TITLE}`;
		this.refreshUsers();
	},
	methods: {
		setErrorMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'error', temporary: true, category: this.$route.name});
		},
		async loadUserList() {
			this.isLoading = true;
			try {
				const response = await this.dakar.authentication.getIdentitiesGet();

				this.users = response.users;
				this.identities = response.identities;
				this.sessions = response.sessions;
				this.$store.dispatch('resetMessages');
			} catch (e) {
				handleError(this, e);
			}

			this.isLoading = false;
		},
		async refreshUsers() {
			await this.loadUserList();

			this.search = '';
			if (!this.users || !this.identities) {
				return;
			}

			this.users = this.users.map(d => {
				// Convert dates to unix time so, they can be sorted in data table
				d.modified = new Date(d.modified).getTime();
				d.created = new Date(d.created).getTime();
				if (d.roles) {
					d.roles = d.roles.map(f => f.name);
					d.renderedRoles = d.roles.toString();
				} else {
					d.renderedRoles = '';
				}

				return d;
			});

			this.identities = this.identities.map(d => {
				// Convert dates to unix time so, they can be sorted in data table
				d.updatedAt = new Date(d.updated_at).getTime();
				d.createdAt = new Date(d.created_at).getTime();
				d.email = d.traits.email;

				if (d.metadata_public) {
					// Extract roles
					if (d.metadata_public.roles) {
						d.roles = d.metadata_public.roles.map(f => f);
						d.renderedRoles = d.roles.toString();
					} else {
						d.renderedRoles = '';
					}

					// Extract dgraph uid
					if (d.metadata_public.dgraph_uid) {
						d.dgraphUID = d.metadata_public.dgraph_uid;
					}
				}

				return d;
			});
		},
		showEditDialog(item) {
			if (this.isLoading) {
				return;
			}

			this.createNewUser = false;
			this.editedItem = {...item};
			this.showCreateIdentityDialog = true;
		},
		showCreateDialog() {
			this.createNewUser = true;
			this.editedItem = {...this.defaultItem};
			this.showCreateIdentityDialog = true;
		},
		showDeleteDialog(identity) {
			if (this.isLoading) {
				return;
			}

			this.showDeleteIdentityDialog = true;
			this.identityToDelete = identity;
		},
		showPropertyDialog(identity) {
			if (this.isLoading) {
				return;
			}

			this.showIdentityPropertyDialog = true;
			this.identityPropertyDialogData = JSON.stringify(identity, null, '\t');
		},
		async deleteIdentity(identity) {
			this.isLoading = true;

			try {
				await this.dakar.authentication.adminDeleteIdentityIdentityUIDGet({identityUID: identity.id});
				await this.refreshUsers();
			} catch (e) {
				this.setErrorMessage(e);
			}

			this.isLoading = false;
			this.closeDeletionDialog();
		},
		closeDeletionDialog() {
			this.showDeleteIdentityDialog = false;
			this.identityToDelete = null;
		},
	},
};
</script>

<style scoped>

</style>
