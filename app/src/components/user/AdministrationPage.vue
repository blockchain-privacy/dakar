<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-container fluid>
    <v-row class="align-center justify-center">
      <v-col>
        <alert :text="errorMsg" />
        <v-data-table
          v-model:sort-by="identitiesSortBy"
          :headers="identityHeaders"
          :items="identities?identities:[]"
          :search="search"
          :loading="isLoading || !identities"
          item-key="id"
          class="mb-10 elevation-4"
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
                  Refresh All
                </div>
              </v-btn>
              <v-btn
                variant="outlined"
                @click="showCreateIdentityDialog"
              >
                <v-icon>{{ mdiAccountPlus }}</v-icon>
                <div class="ml-2 hidden-sm-and-down">
                  Create
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
                <v-list-item @click="showDeleteIdentityDialog(item)">
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
                <v-list-item @click="showDeleteSessionDialog(item.id)">
                  <template #prepend>
                    <v-icon :icon="mdiDelete" />
                  </template>
                  Delete
                </v-list-item>
              </v-list>
            </v-menu>
          </template>
        </v-data-table>
        <v-data-table
          v-model:sort-by="oauthSessionsSortBy"
          :headers="oauthSessionHeaders"
          :items="oauthSessions?oauthSessions:[]"
          :search="searchOAuthSessions"
          :loading="isLoading || !oauthSessions"
          item-key="id"
          class="my-10 elevation-4"
        >
          <template #top>
            <v-toolbar flat>
              <v-toolbar-title>OAuth 2 Sessions</v-toolbar-title>
              <v-spacer />
              <v-text-field
                v-model="searchOAuthSessions"
                class="me-3"
                :append-inner-icon="mdiMagnify"
                label="Filter OAuth 2 sessions"
                single-line
                hide-details
              />
            </v-toolbar>
          </template>
          <template #item.handled_at="{ item }">
            <span>{{ new Date(item.handled_at).toLocaleString() }}</span>
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
                <v-list-item @click="showDeleteConsentDialog(item.consent_request.consent_request_id)">
                  <template #prepend>
                    <v-icon :icon="mdiDelete" />
                  </template>
                  Delete
                </v-list-item>
              </v-list>
            </v-menu>
          </template>
        </v-data-table>
        <v-data-table
          v-model:sort-by="oauthClientsSortBy"
          :headers="oauthClientsHeaders"
          :items="oauthClients?oauthClients:[]"
          :search="searchOAuthClients"
          :loading="isLoading || !oauthClients"
          item-key="id"
          class="my-10 elevation-4"
        >
          <template #top>
            <v-toolbar flat>
              <v-toolbar-title>OAuth 2 Clients</v-toolbar-title>
              <v-spacer />
              <v-text-field
                v-model="searchOAuthClients"
                class="me-3"
                :append-inner-icon="mdiMagnify"
                label="Filter OAuth 2 clients"
                single-line
                hide-details
              />
              <v-spacer />
              <v-btn
                variant="outlined"
                @click="showCreateClientDialog"
              >
                <v-icon :icon="mdiPlus" />
                Create
              </v-btn>
            </v-toolbar>
          </template>
          <template #item.updated_at="{ item }">
            <span>{{ new Date(item.updated_at).toLocaleString() }}</span>
          </template>
          <template #item.created_at="{ item }">
            <span>{{ new Date(item.created_at).toLocaleString() }}</span>
          </template>
          <template #item.redirect_uris="{ item }">
            <span>{{ item.redirect_uris.join(', ') }}</span>
          </template>
          <template #item.grant_types="{ item }">
            <span>{{ item.grant_types.join(', ') }}</span>
          </template>
          <template #item.response_types="{ item }">
            <span>{{ item.response_types.join(', ') }}</span>
          </template>
          <template #item.audience="{ item }">
            <span>{{ item.audience.join(', ') }}</span>
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
                <v-list-item @click="showEditClientDialog(item)">
                  <template #prepend>
                    <v-icon :icon="mdiPencil" />
                  </template>
                  Edit
                </v-list-item>
                <v-list-item @click="showDeleteClientDialog(item.client_id)">
                  <template #prepend>
                    <v-icon :icon="mdiDelete" />
                  </template>
                  Delete
                </v-list-item>
              </v-list>
            </v-menu>
          </template>
        </v-data-table>
        <edit-identity-dialog
          v-if="showCreateIdentityDialogModel"
          v-model="showCreateIdentityDialogModel"
          :create-new-user="createNewUser"
          :identity="editedItem"
          @saved="refreshUsers()"
        />
        <create-client-dialog
          v-if="showCreateClientDialogModel"
          v-model="showCreateClientDialogModel"
          :is-edit="updateClient"
          :client="updateClientData"
          @created="refreshUsers()"
        />
        <deletion-dialog
          v-if="identityToDelete"
          :id="identityToDelete.id"
          v-model="showDeleteIdentityDialogModel"
          title="Delete Identity"
          confirmation-text="Do you really want to delete this identity?"
          :properties="[['ID',identityToDelete.id ],['E-mail', identityToDelete.email]]"
          @canceled="identityToDelete = null"
          @accepted="deleteIdentity"
        />
        <deletion-dialog
          v-if="clientToDelete"
          :id="clientToDelete"
          v-model="showDeleteClientDialogModel"
          title="Delete Client"
          confirmation-text="Do you really want to delete this client?"
          :properties="[['ID',clientToDelete ]]"
          @canceled="clientToDelete = null"
          @accepted="deleteClient"
        />
        <deletion-dialog
          v-if="consentToDelete"
          :id="consentToDelete"
          v-model="showDeleteConsentDialogModel"
          title="Delete Consent Session"
          confirmation-text="Do you really want to delete this consent session?"
          :properties="[['ID',consentToDelete ]]"
          @canceled="consentToDelete = null"
          @accepted="deleteConsent"
        />
        <deletion-dialog
          v-if="sessionToDelete"
          :id="sessionToDelete"
          v-model="showDeleteSessionDialogModel"
          title="Delete Identity Session"
          confirmation-text="Do you really want to delete this identity session?"
          :properties="[['ID',sessionToDelete ]]"
          @canceled="sessionToDelete = null"
          @accepted="deleteSession"
        />
        <v-dialog
          v-model="showIdentityPropertyDialogModel"
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
	mdiPencil,
	mdiDelete,
	mdiRefresh,
	mdiAccountPlus,
	mdiMagnify,
	mdiUnfoldMoreVertical,
	mdiDotsVertical,
	mdiPlus,
} from '@mdi/js';
import {inject, onMounted, ref} from 'vue';
import {PAGE_TITLE} from '@/constants';
import EditIdentityDialog from '@/components/user/EditIdentityDialog.vue';
import Alert from '@/components/common/Alert.vue';
import DeletionDialog from '@/components/user/DeletionDialog.vue';
import CreateClientDialog from '@/components/user/CreateClientDialog.vue';

const kratosAdmin = inject('kratosadmin');

const isLoading = ref(false);
const showCreateIdentityDialogModel = ref(false);
const showDeleteIdentityDialogModel = ref(false);
const showIdentityPropertyDialogModel = ref(false);
const showDeleteClientDialogModel = ref(false);
const showDeleteConsentDialogModel = ref(false);
const showDeleteSessionDialogModel = ref(false);
const showCreateClientDialogModel = ref(false);

const identityToDelete = ref(null);
const clientToDelete = ref(null);
const sessionToDelete = ref(null);
const consentToDelete = ref(null);

const search = ref('');
const searchSessions = ref('');
const searchOAuthSessions = ref('');
const searchOAuthClients = ref('');
const errorMsg = ref('');

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
	{title: 'E-Mail', key: 'identity.traits.email'},
	{title: 'Active', key: 'active'},
	{title: 'Authentication Date', key: 'authenticated_at'},
	{title: 'Expiry Date', key: 'expires_at'},
	{
		title: '', key: 'actions', sortable: false, align: 'end',
	},
];

const oauthSessionsSortBy = ref([{key: 'handled_at', order: 'desc'}]);
const oauthSessionHeaders = [
	{
		title: 'ID', key: 'consent_request.subject', align: 'start', sortable: false,
	},
	{title: 'E-Mail', key: 'email'},
	{title: 'Client ID', key: 'consent_request.client.client_id'},
	{title: 'Handled At', key: 'handled_at'},
	{
		title: '', key: 'actions', sortable: false, align: 'end',
	},
];

const oauthClientsSortBy = ref([{key: 'updated_at', order: 'desc'}]);
const oauthClientsHeaders = [
	{
		title: 'ID', key: 'client_id', align: 'start', sortable: false,
	},
	{title: 'Name', key: 'client_name'},
	{title: 'Scopes', key: 'scope'},
	{title: 'Grant Types', key: 'grant_types'},
	{title: 'Redirect URIs', key: 'redirect_uris'},
	{title: 'Audience', key: 'audience'},
	{title: 'Skip Consent', key: 'skip_consent'},
	{title: 'Created At', key: 'created_at'},
	{title: 'Response Types', key: 'response_types'},
	{title: 'Updated At', key: 'updated_at'},
	{
		title: '', key: 'actions', sortable: false, align: 'end',
	},
];

const createNewUser = ref(false);
const updateClient = ref(false);
const editedItem = ref({
	id: '', email: '', state: '', roles: {},
});
const defaultItem = ref({
	id: '', email: '', state: '', roles: {},
});
// Client data for the client update dialog.
const updateClientData = ref(null);

const identities = ref(null);
const sessions = ref(null);
const oauthSessions = ref(null);
const oauthClients = ref(null);
const identityPropertyDialogData = ref(null);

onMounted(() => {
	document.title = `User Administration - ${PAGE_TITLE}`;
	refreshUsers();
});

function setErrorMessage(msg) {
	errorMsg.value = msg;
}

async function loadUserList() {
	isLoading.value = true;
	errorMsg.value = '';
	try {
		const response = await kratosAdmin.identity.identitiesGet();

		identities.value = response.identities || [];
		sessions.value = response.sessions || [];
		oauthSessions.value = response.oauthSessions || [];
		oauthClients.value = response.oauthClients || [];
	} catch (error) {
		errorMsg.value = error.message;
	}

	isLoading.value = false;
}

async function refreshUsers() {
	await loadUserList();

	search.value = '';
	if (!identities.value) {
		return;
	}

	// Add email to oauth sessions so they can be easier identified
	oauthSessions.value = oauthSessions.value.map(d => {
		const {subject} = d.consent_request;
		const identity = identities.value.find(i => i.id === subject);
		if (identity) {
			d.email = identity.traits.email;
		}

		return d;
	});

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
	showCreateIdentityDialogModel.value = true;
}

function showCreateIdentityDialog() {
	if (isLoading.value) {
		return;
	}

	createNewUser.value = true;
	editedItem.value = {...defaultItem.value};
	showCreateIdentityDialogModel.value = true;
}

function showCreateClientDialog() {
	if (isLoading.value) {
		return;
	}

	updateClient.value = false;
	updateClientData.value = null;
	showCreateClientDialogModel.value = true;
}

function showDeleteIdentityDialog(identity) {
	if (isLoading.value) {
		return;
	}

	showDeleteIdentityDialogModel.value = true;
	identityToDelete.value = identity;
}

function showDeleteClientDialog(clientID) {
	if (isLoading.value) {
		return;
	}

	showDeleteClientDialogModel.value = true;
	clientToDelete.value = clientID;
}

function showDeleteSessionDialog(session) {
	if (isLoading.value) {
		return;
	}

	showDeleteSessionDialogModel.value = true;
	sessionToDelete.value = session;
}

function showDeleteConsentDialog(consent) {
	if (isLoading.value) {
		return;
	}

	showDeleteConsentDialogModel.value = true;
	consentToDelete.value = consent;
}

function showPropertyDialog(identity) {
	if (isLoading.value) {
		return;
	}

	showIdentityPropertyDialogModel.value = true;
	identityPropertyDialogData.value = JSON.stringify(identity, null, '\t');
}

async function deleteIdentity(identityID) {
	isLoading.value = true;

	try {
		await kratosAdmin.identity.identitiesUidDelete({uid: identityID});
		await refreshUsers();
	} catch (error) {
		setErrorMessage(error);
	}

	isLoading.value = false;
	identityToDelete.value = null;
}

async function deleteClient(clientId) {
	isLoading.value = true;
	errorMsg.value = '';
	try {
		await kratosAdmin.oauth.clientsClientIdDelete({clientId});
		await refreshUsers();
	} catch (error) {
		setErrorMessage(error);
	}

	isLoading.value = false;
	clientToDelete.value = null;
}

function showEditClientDialog(item) {
	if (isLoading.value) {
		return;
	}

	updateClient.value = true;
	updateClientData.value = item;
	showCreateClientDialogModel.value = true;
}

async function deleteSession(sessionId) {
	isLoading.value = true;
	errorMsg.value = '';
	try {
		await kratosAdmin.identity.sessionsSessionIdDelete({sessionId});
		await refreshUsers();
	} catch (error) {
		setErrorMessage(error);
	}

	isLoading.value = false;
	sessionToDelete.value = null;
}

async function deleteConsent(consentId) {
	isLoading.value = true;
	errorMsg.value = '';
	try {
		await kratosAdmin.oauth.consentsConsentIdDelete({consentId});
		await refreshUsers();
	} catch (error) {
		setErrorMessage(error);
	}

	isLoading.value = false;
	consentToDelete.value = null;
}

</script>

<style scoped>

</style>
