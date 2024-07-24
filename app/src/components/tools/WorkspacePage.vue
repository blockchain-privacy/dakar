<template>
  <v-card
    class="mx-auto"
    variant="text"
    max-width="1200"
  >
    <icon-title
      title="Workspaces"
      icon="$graphIcon"
      :one-line="true"
    >
      <v-btn
        v-model="showSearchField"
        variant="text"
        @click="showAddWorkspaceDialogModel = true"
      >
        <v-icon :icon="mdiPlus" />
        Add Workspace
      </v-btn>
      <v-btn
        v-model="showSearchField"
        variant="text"
        :icon="true"
        @click="showSearchField = !showSearchField"
      >
        <v-icon>{{ mdiMagnify }}</v-icon>
      </v-btn>
      <v-menu location="bottom">
        <template #activator="{ props }">
          <v-btn
            variant="text"
            :icon="true"
            v-bind="props"
          >
            <v-icon>{{ mdiDotsVertical }}</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item
            :disabled="isLoading"
            @click="refreshWorkspaceList"
          >
            <template #prepend>
              <v-icon>{{ mdiRefresh }}</v-icon>
            </template>
            <v-list-item-title>Refresh</v-list-item-title>
          </v-list-item>
          <v-list-item @click="showDeleteAllWorkspacesDialog">
            <template #prepend>
              <v-icon>{{ mdiDelete }}</v-icon>
            </template>
            <v-list-item-title>Delete All Workspaces</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </icon-title>
    <fade-transition>
      <div
        v-if="showSearchField"
        class="d-flex align-center justify-center mb-4"
      >
        <v-text-field
          v-model="search"
          :append-inner-icon="mdiMagnify"
          label="Filter items"
          :single-line="true"
          hide-details
          :autofocus="true"
          style="max-width:800px"
          @keydown.esc="showSearchField = false"
        />
      </div>
    </fade-transition>
    <v-data-table
      v-model:sort-by="sortBy"
      :search="search"
      :loading="isLoading"
      :headers="headers"
      :items="workspaceList"
    >
      <template #item.name="{ item }">
        <router-link
          :to="{ name: ROUTE_NAME_WORKSPACE_PAGE,
                 params: { id: item.uid }}"
        >
          {{ item.name }}
        </router-link>
      </template>
      <template #item.modTimeUnix="{ item }">
        <span>{{ new Date(item.modTimeUnix).toLocaleString() }}</span>
      </template>
      <template #[`item.actions`]="{ item }">
        <v-icon @click="showDeleteWorkspaceDialog(item)">
          {{ mdiDelete }}
        </v-icon>
      </template>
    </v-data-table>
    <v-dialog
      v-model="showDeleteAllDialog"
      max-width="500px"
    >
      <v-card>
        <v-card-title>
          <span class="text-h5">Delete All Workspaces</span>
        </v-card-title>
        <v-card-text>
          <p class="text-subtitle-1">
            Are you sure you want to delete all of your workspaces?
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            variant="text"
            @click="closeDeleteAllWorkspacesDialog"
          >
            Cancel
          </v-btn>
          <v-btn
            color="red"
            variant="text"
            @click="deleteWorkspace(true)"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog
      v-if="workspaceToDelete"
      v-model="showDeleteWorkspaceDialogModel"
      max-width="500px"
    >
      <v-card>
        <v-card-title>
          <span class="text-h5">Delete Workspace</span>
        </v-card-title>
        <v-card-text>
          <p class="text-subtitle-1">
            Workspace <code>{{ workspaceToDelete.name }}</code>
            will be deleted. Continue?
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            variant="text"
            @click="closeDeleteWorkspaceDialog"
          >
            Cancel
          </v-btn>
          <v-btn
            color="red"
            variant="text"
            @click="deleteWorkspace(false)"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog
      v-model="showAddWorkspaceDialogModel"
      max-width="500px"
    >
      <v-card>
        <v-card-title>
          <span class="text-h5">Add Workspace</span>
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="newWorkspaceName"
            label="Name of the new workspace"
            :autofocus="true"
            @keydown.enter="addWorkspace(newWorkspaceName)"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            variant="text"
            color="red"
            @click="closeAddWorkspaceDialog"
          >
            Cancel
          </v-btn>
          <v-btn
            variant="text"
            @click="addWorkspace(newWorkspaceName)"
          >
            Add
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script setup>
import {
	mdiRefresh, mdiDelete, mdiMagnify, mdiDotsVertical, mdiPlus,
} from '@mdi/js';
import {PAGE_TITLE, ROUTE_NAME_WORKSPACE_PAGE} from '@/constants';
import {handleError} from '@/utilities';
import IconTitle from '@/components/common/IconTitle.vue';
import FadeTransition from '@/components/common/FadeTransition.vue';
import {inject, onMounted, ref} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();
const context = {addMessage: msgStore.addMessage, $route: route};

const workspaceList = ref([]);
const showDeleteAllDialog = ref(false);
const showDeleteWorkspaceDialogModel = ref(false);
const showAddWorkspaceDialogModel = ref(false);
const workspaceToDelete = ref(null);
const isLoading = ref(false);
const showSearchField = ref(false);
const search = ref('');
const sortBy = ref([{key: 'modTimeUnix', order: 'desc'}]);
const newWorkspaceName = ref('');
const headers = [
	{
		title: 'Name', key: 'name', align: 'start', sortable: false,
	},
	{
		title: 'Last modification', key: 'modTimeUnix',
	},
	{
		title: '', key: 'actions', sortable: false, align: 'end',
	},
];

// Hooks
onMounted(() => {
	document.title = `Workspaces - ${PAGE_TITLE}`;
	refreshWorkspaceList();
});

// Functions
function setErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: false, category: route.name,
	});
}

function setInfoMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'info', temporary: true, category: route.name,
	});
}

async function loadWorkspaceList() {
	isLoading.value = true;
	try {
		const response = await dakar.workspace.workspacesGet();

		if (!response.workspaces) {
			throw new Error('received malformed response');
		}

		workspaceList.value = response.workspaces;
		msgStore.resetMessages();
	} catch (e) {
		handleError(context, e);
	}

	isLoading.value = false;
}

async function addWorkspace(name) {
	showAddWorkspaceDialogModel.value = false;
	const workspaceName = name.trim();
	if (workspaceName === '') {
		setErrorMessage('workspace name must not be empty');
		return;
	}

	isLoading.value = true;
	try {
		await dakar.workspace.workspacesNamePost({name: workspaceName});
		msgStore.resetMessages();
		await refreshWorkspaceList();
	} catch (e) {
		handleError(context, e);
	}

	isLoading.value = false;
}

async function refreshWorkspaceList() {
	await loadWorkspaceList();

	search.value = '';

	if (!workspaceList.value) {
		return;
	}

	workspaceList.value = workspaceList.value.map(d => {
		// Convert date to unix time, so it can be sorted in data table
		d.modTimeUnix = new Date(d.ts).getTime();
		return d;
	});
}

async function deleteWorkspace(all) {
	isLoading.value = true;

	if (all) {
		try {
			const response = await dakar.workspace.workspacesDelete();
			if (response.msg) {
				setInfoMessage(response.msg);
			}

			await refreshWorkspaceList();
		} catch (e) {
			setErrorMessage(e);
		}
	} else {
		try {
			const response = await dakar.workspace.workspacesUidDelete({uid: workspaceToDelete.value.uid});
			if (response.msg) {
				setInfoMessage(response.msg);
			}

			await refreshWorkspaceList();
		} catch (e) {
			setErrorMessage(e);
		}
	}

	isLoading.value = false;
	showDeleteWorkspaceDialogModel.value = false;
	showDeleteAllDialog.value = false;
}

function showDeleteWorkspaceDialog(workspace) {
	if (isLoading.value) {
		return;
	}

	showDeleteWorkspaceDialogModel.value = true;
	workspaceToDelete.value = workspace;
}

function closeDeleteWorkspaceDialog() {
	showDeleteWorkspaceDialogModel.value = false;
}

function 	showDeleteAllWorkspacesDialog() {
	showDeleteAllDialog.value = true;
}

function 	closeDeleteAllWorkspacesDialog() {
	showDeleteAllDialog.value = false;
}

function closeAddWorkspaceDialog() {
	showAddWorkspaceDialogModel.value = false;
}

</script>

<style scoped>

</style>
