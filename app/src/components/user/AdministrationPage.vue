<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-container fluid>
    <v-row
      align="center"
      justify="center"
    >
      <v-col cols="12">
        <v-data-table
          v-model:sort-by="identitiesSortBy"
          :headers="identityHeaders"
          :items="identities?identities:[]"
          :search="search"
          :loading="isLoading || !identities"
          item-key="id"
          class="my-10 elevation-4"
        >
          <template #top>
            <v-toolbar flat>
              <v-toolbar-title>Identities</v-toolbar-title>
              <v-spacer />
              <v-text-field
                v-model="search"
                :append-inner-icon="mdiMagnify"
                label="Filter users"
                single-line
                hide-details
                style="max-width: 500px"
              />
              <v-spacer />
              <v-btn
                class="me-2"
                variant="outlined"
                :disabled="isLoading"
                @click="refreshUsers"
              >
                <v-icon>{{ mdiRefresh }}</v-icon>
                <div class="ml-2 hidden-sm-and-down">
                  Refresh
                </div>
              </v-btn>
              <v-btn
                variant="outlined"
                @click="showCreateDialog"
              >
                <v-icon>{{ mdiAccountPlus }}</v-icon>
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
                  <v-icon>{{ mdiDotsVertical }}</v-icon>
                </v-btn>
              </template>
              <v-list>
                <v-list-item @click="showEditDialog(item)">
                  <template #prepend>
                    <v-icon :icon="mdiPencil" />
                  </template>
                  Edit
                </v-list-item>
                <v-list-item @click="showPropertyDialog(item)">
                  <template #prepend>
                    <v-icon :icon="mdiUnfoldMoreVertical" />
                  </template>
                  Details
                </v-list-item>
                <v-list-item @click="showDeleteDialog(item)">
                  <template #prepend>
                    <v-icon :icon="mdiDelete" />
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
          class="my-10 elevation-4"
        >
          <template #top>
            <v-toolbar flat>
              <v-toolbar-title>Sessions</v-toolbar-title>
              <v-spacer />
              <v-text-field
                v-model="searchSessions"
                class="me-3"
                :append-inner-icon="mdiMagnify"
                label="Filter sessions"
                single-line
                hide-details
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
              readonly
              hide-details
            />
          </v-card>
        </v-dialog>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import {
	mdiPencil, mdiDelete, mdiRefresh, mdiAccountPlus,
	mdiMagnify, mdiUnfoldMoreVertical, mdiDotsVertical,
} from '@mdi/js';
import {PAGE_TITLE} from '@/constants';
import {handleError} from '@/utilities';
import EditIdentityDialog from '@/components/user/EditIdentityDialog.vue';
import {inject, onMounted, ref} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const kratosAdmin = inject('kratosadmin');
const route = useRoute();
const msgStore = useMsgStore();
const context = {addMessage: msgStore.addMessage, $route: route};

const isLoading = ref(false);
const showCreateIdentityDialog = ref(false);
const showDeleteIdentityDialog = ref(false);
const showIdentityPropertyDialog = ref(false);
const identityToDelete = ref(null);
const search = ref('');
const searchSessions = ref('');

const identitiesSortBy = ref([{key: 'modified', order: 'desc'}]);

const identityHeaders = [
	{
		title: 'ID', key: 'id', align: 'start', sortable: false,
	},
	{title: 'E-Mail', key: 'email'},
	{title: 'Dakar Dash User', key: 'dakarDashUser'},
	{title: 'Dakar BTC User', key: 'dakarBTCUser'},
	{title: 'State', key: 'state'},
	{title: 'Schema ID', key: 'schema_id'},
	{title: 'Role Dakar Dash', key: 'roleDakarDash'},
	{title: 'Role Dakar BTC', key: 'roleDakarBTC'},
	{title: 'Role Kratos Admin', key: 'roleKratosAdmin'},
	{title: 'Created', key: 'createdAt'},
	{title: 'Updated', key: 'updatedAt'},
	{
		title: '', key: 'actions', sortable: false, align: 'end',
	},
];

const sessionsSortBy = ref([{key: 'authenticated_at', order: 'desc'}]);
const sessionHeaders = [
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
];

const createNewUser = ref(false);
const editedItem = ref({
	id: '', email: '', state: '', roles: {},
});
const defaultItem = ref({
	id: '', email: '', state: '', roles: {},
});
const identities = ref(null);
const sessions = ref(null);
const identityPropertyDialogData = ref(null);

onMounted(() => {
	document.title = `User Administration - ${PAGE_TITLE}`;
	refreshUsers();
});

function setErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: true, category: route.name,
	});
}

async function loadUserList() {
	isLoading.value = true;
	try {
		const response = await kratosAdmin.identitiesGet();

		identities.value = response.identities;
		sessions.value = response.sessions;
		msgStore.resetMessages();
	} catch (e) {
		handleError(context, e);
	}

	isLoading.value = false;
}

async function refreshUsers() {
	await loadUserList();

	search.value = '';
	if (!identities.value) {
		return;
	}

	identities.value = identities.value.map(d => {
		// Convert dates to unix time so, they can be sorted in data table
		d.updatedAt = new Date(d.updated_at).getTime();
		d.createdAt = new Date(d.created_at).getTime();
		d.email = d.traits.email;

		if (d.metadata_public) {
			// Extract roles
			if (d.metadata_public.roles) {
				// For table
				d.roleDakarDash = d.metadata_public.roles.dakar_dash;
				d.roleDakarBTC = d.metadata_public.roles.dakar_btc;
				d.roleKratosAdmin = d.metadata_public.roles.kratos_admin;

				// For dialog
				d.roles = {
					// eslint-disable-next-line camelcase
					dakar_dash: d.metadata_public.roles.dakar_dash,
					// eslint-disable-next-line camelcase
					dakar_btc: d.metadata_public.roles.dakar_btc,
					// eslint-disable-next-line camelcase
					kratos_admin: d.metadata_public.roles.kratos_admin,
				};
			}

			// Extract user UIDs
			if (d.metadata_public.dakar_dash_user) {
				d.dakarDashUser = d.metadata_public.dakar_dash_user;
			}

			if (d.metadata_public.dakar_btc_user) {
				d.dakarBTCUser = d.metadata_public.dakar_btc_user;
			}
		}

		return d;
	});
}

function showEditDialog(item) {
	if (isLoading.value) {
		return;
	}

	createNewUser.value = false;
	editedItem.value = {...item};
	showCreateIdentityDialog.value = true;
}

function showCreateDialog() {
	createNewUser.value = true;
	editedItem.value = {...defaultItem.value};
	showCreateIdentityDialog.value = true;
}

function showDeleteDialog(identity) {
	if (isLoading.value) {
		return;
	}

	showDeleteIdentityDialog.value = true;
	identityToDelete.value = identity;
}

function showPropertyDialog(identity) {
	if (isLoading.value) {
		return;
	}

	showIdentityPropertyDialog.value = true;
	identityPropertyDialogData.value = JSON.stringify(identity, null, '\t');
}

async function deleteIdentity(identity) {
	isLoading.value = true;

	try {
		await kratosAdmin.identitiesUidDelete({uid: identity.id});
		await refreshUsers();
	} catch (e) {
		setErrorMessage(e);
	}

	isLoading.value = false;
	closeDeletionDialog();
}

function closeDeletionDialog() {
	showDeleteIdentityDialog.value = false;
	identityToDelete.value = null;
}

</script>

<style scoped>

</style>
