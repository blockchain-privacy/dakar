<template>
  <v-card
    class="mx-auto"
    variant="text"
    max-width="1200"
  >
    <icon-title
      title="Workspaces"
      :icon="mdiGraph"
      :one-line="true"
    >
      <v-btn
        v-model="showSearchField"
        variant="text"
        @click="addWorkspace"
      >
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
          single-line
          hide-details
          style="max-width:800px"
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
  </v-card>
</template>

<script setup>
import {
	mdiGraph, mdiRefresh, mdiDelete, mdiMagnify, mdiDotsVertical,
} from '@mdi/js';
import {PAGE_TITLE} from '@/constants';
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
const workspaceToDelete = ref(null);
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

// Hooks
onMounted(() => {
	document.title = `Workspaces - ${PAGE_TITLE}`;
	refreshWorkspaceList();
});

// Functions
function setErrorMessage(msg) {
	msgStore.addMessage({text: msg, type: 'error', temporary: false, category: route.name});
}

function setInfoMessage(msg) {
	msgStore.addMessage({text: msg, type: 'info', temporary: true, category: route.name});
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

async function addWorkspace() {
	isLoading.value = true;
	try {
		await dakar.workspace.addWorkspaceNameGet({name: 'test'});
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
	let arg;
	if (all) {
		// eslint-disable-next-line camelcase
		arg = {delete_all: true};
	} else {
		arg = {uid: workspaceToDelete.value.uid};
	}

	try {
		// Todo implement
		// Const response = await dakar.heuristic.deleteHeuristicPost({heuristic: arg});
		if (response.msg) {
			setInfoMessage(response.msg);
		}

		await refreshWorkspaceList();
	} catch (e) {
		setErrorMessage(e);
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

</script>

<style scoped>

</style>
