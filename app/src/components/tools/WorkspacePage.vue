<template>
  <v-card
    class="mx-auto"
    variant="text"
    max-width="1200"
  >
    <icon-title
      title="Workspaces"
      icon="$graphIcon"
      one-line
    >
      <v-btn
        v-model="showSearchField"
        variant="text"
        :icon="display.xs.value"
        @click="showAddWorkspaceDialogModel = true"
      >
        <v-icon :icon="mdiPlus" />
        <div class="hidden-xs">
          New
        </div>
      </v-btn>
      <v-btn
        v-model="showSearchField"
        variant="text"
        icon
        @click="showSearchField = !showSearchField"
      >
        <v-icon>{{ mdiMagnify }}</v-icon>
      </v-btn>
      <wiki-tooltip
        description-url="workspaces/workspaces.md"
        :icon="mdiHelpCircleOutline"
        icon-color="primary"
      />
      <v-menu location="bottom">
        <template #activator="{ props }">
          <v-btn
            v-bind="props"
            variant="text"
            icon
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
          single-line
          hide-details
          autofocus
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
        <div class="d-flex">
          <v-icon
            class="me-2 ms-auto"
            @click="showRenameDialog(item)"
          >
            {{ mdiRename }}
          </v-icon>
          <v-icon @click="showDeleteWorkspaceDialog(item)">
            {{ mdiDelete }}
          </v-icon>
        </div>
      </template>
    </v-data-table>
    <confirm-dialog
      v-if="showDeleteAllDialog"
      v-model="showDeleteAllDialog"
      title="Delete All Workspaces"
      confirm-label="Delete All"
      confirm-color="red"
      @confirm="deleteWorkspace(true)"
    >
      <p class="text-subtitle-1">
        Are you sure you want to delete all of your workspaces?
      </p>
    </confirm-dialog>
    <confirm-dialog
      v-if="showDeleteWorkspaceDialogModel"
      v-model="showDeleteWorkspaceDialogModel"
      title="Delete Workspace"
      confirm-label="Delete"
      confirm-color="red"
      @confirm="deleteWorkspace(false)"
    >
      <p class="text-subtitle-1">
        Workspace <code>{{ workspaceToDelete.name }}</code>
        will be deleted. Continue?
      </p>
    </confirm-dialog>
    <text-dialog
      v-if="showAddWorkspaceDialogModel"
      v-model="showAddWorkspaceDialogModel"
      title="New Workspace"
      submit-label="Create"
      input-label="Name of the new workspace"
      :maxlength="maxWorkspaceNameLength"
      @submit="addWorkspace"
    />
    <text-dialog
      v-if="showRenameWorkspaceDialogModel"
      v-model="showRenameWorkspaceDialogModel"
      title="Rename Workspace"
      submit-label="Rename"
      input-label="New workspace name"
      :input-value="renamedWorkspace?.name"
      :maxlength="maxWorkspaceNameLength"
      @submit="renameWorkspace"
    />
  </v-card>
</template>

<script setup>
import {
	mdiRefresh, mdiDelete, mdiMagnify, mdiDotsVertical, mdiPlus, mdiRename, mdiHelpCircleOutline,
} from '@mdi/js';
import {PAGE_TITLE, ROUTE_NAME_WORKSPACE_PAGE} from '@/constants';
import {handleError} from '@/utilities';
import IconTitle from '@/components/common/IconTitle.vue';
import FadeTransition from '@/components/common/FadeTransition.vue';
import {
	inject, onMounted, ref, toRaw,
} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import TextDialog from '@/components/common/TextDialog.vue';
import ConfirmDialog from '@/components/common/ConfirmDialog.vue';
import {useDisplay} from 'vuetify';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();
const display = useDisplay();
const context = {addMessage: msgStore.addMessage, $route: route};

const workspaceList = ref([]);
const showDeleteAllDialog = ref(false);
const showDeleteWorkspaceDialogModel = ref(false);
const showAddWorkspaceDialogModel = ref(false);
const showRenameWorkspaceDialogModel = ref(false);
const workspaceToDelete = ref(null);
const renamedWorkspace = ref(null);
const isLoading = ref(false);
const showSearchField = ref(false);
const search = ref('');
const sortBy = ref([{key: 'modTimeUnix', order: 'desc'}]);
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

const maxWorkspaceNameLength = 50;

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

		if (response.workspaces) {
			workspaceList.value = response.workspaces;
		} else {
			workspaceList.value = [];
		}

		msgStore.resetMessages();
	} catch (e) {
		handleError(context, e);
	}

	isLoading.value = false;
}

async function renameWorkspace(workspace) {
	const workspaceName = workspace;
	const workspaceUID = renamedWorkspace.value.uid;
	if (workspaceName === '') {
		setErrorMessage('workspace name must not be empty');
		return;
	}

	if (workspaceName.length > maxWorkspaceNameLength) {
		setErrorMessage(`workspace name is longer than the maximum of ${maxWorkspaceNameLength} characters`);
		return;
	}

	if (workspaceUID === '') {
		setErrorMessage('workspace UID is not set');
		return;
	}

	isLoading.value = true;

	try {
		await dakar.workspace.workspacesRenamePost({
			workspace: {name: workspaceName, workspaceUID},
		});
		msgStore.resetMessages();
		await refreshWorkspaceList();
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

	if (workspaceName.length > maxWorkspaceNameLength) {
		setErrorMessage(`workspace name is longer than the maximum of ${maxWorkspaceNameLength} characters`);
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
}

function showRenameDialog(workspace) {
	if (isLoading.value) {
		return;
	}

	// Workspace is a ref -> ned to convert and clone it
	renamedWorkspace.value = structuredClone(toRaw(workspace));
	showRenameWorkspaceDialogModel.value = true;
}

function showDeleteWorkspaceDialog(workspace) {
	if (isLoading.value) {
		return;
	}

	showDeleteWorkspaceDialogModel.value = true;
	workspaceToDelete.value = workspace;
}

function showDeleteAllWorkspacesDialog() {
	if (isLoading.value) {
		return;
	}

	showDeleteAllDialog.value = true;
}

</script>

<style scoped>

</style>
