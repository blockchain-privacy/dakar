<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

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
        v-if="workspaceList.length > 0"
        variant="text"
        :icon="display.xs.value"
        @click="showAddWorkspaceDialogModel = true"
      >
        <v-icon :icon="mdiPlus" />
        <div class="hidden-xs">
          Add Workspace
        </div>
      </v-btn>
      <v-btn
        v-if="workspaceList.length > 0"
        :active="showSearchField"
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
          @keydown.esc="search = ''; showSearchField = false"
        />
      </div>
    </fade-transition>
    <v-data-table
      v-if="workspaceList.length > 0"
      v-model:sort-by="sortBy"
      :search="search"
      :loading="isLoading"
      :headers="headers"
      :items="workspaceList"
    >
      <template #item.name="{ item }">
        <router-link
          :to="{ name: ROUTE_NAME_WORKSPACE_PAGE, params: { id: item.uid, blockchainMode: item.mode }}"
        >
          {{ item.name }}
        </router-link>
      </template>
      <template #item.mode="{ item }">
        <div class="d-flex align-center">
          <v-icon
            :icon="BLOCKCHAIN_ATTRIBUTES[item.mode].icon"
            :color="BLOCKCHAIN_ATTRIBUTES[item.mode].color"
            start
            size="x-large"
          />
          {{ BLOCKCHAIN_ATTRIBUTES[item.mode].title }}
        </div>
      </template>
      <template #item.modTimeUnix="{ item }">
        <span>{{ new Date(item.modTimeUnix).toLocaleString() }}</span>
      </template>
      <template #[`item.actions`]="{ item }">
        <div class="d-flex">
          <v-icon
            start
            class="ms-auto"
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
    <v-progress-linear
      v-else-if="isLoading"
      class="ma-5"
      indeterminate
    />
    <div
      v-else
      class="d-flex justify-center"
    >
      <v-btn
        variant="text"
        :prepend-icon="mdiPlus"
        @click="showAddWorkspaceDialogModel = true"
      >
        Add new workspace
      </v-btn>
    </div>
    <confirm-dialog
      v-if="showDeleteWorkspaceDialogModel"
      v-model="showDeleteWorkspaceDialogModel"
      title="Delete Workspace"
      confirm-label="Delete"
      confirm-color="red"
      @confirm="deleteWorkspace"
    >
      <p class="text-subtitle-1">
        Workspace <code>{{ workspaceToDelete.name }}</code>
        will be deleted. Continue?
      </p>
    </confirm-dialog>
    <blockchain-mode-text-dialog
      v-if="showAddWorkspaceDialogModel"
      v-model="showAddWorkspaceDialogModel"
      title="New Workspace"
      submit-label="Create"
      input-label="Workspace name"
      :maxlength="maxWorkspaceNameLength"
      show-mode-switch
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
	mdiDelete, mdiMagnify, mdiPlus, mdiRename, mdiHelpCircleOutline,
} from '@mdi/js';
import {
	BLOCKCHAIN_ATTRIBUTES, PAGE_TITLE, ROUTE_NAME_WORKSPACE_PAGE,
} from '@/constants/index.js';
import {
	getDakarClients, handleError, isAdminIdentity, isPrivilegedIdentity,
} from '@/utilities/index.js';
import IconTitle from '@/components/common/IconTitle.vue';
import FadeTransition from '@/components/common/FadeTransition.vue';
import {
	computed, onMounted, ref, toRaw,
} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg.js';
import TextDialog from '@/components/common/TextDialog.vue';
import ConfirmDialog from '@/components/common/ConfirmDialog.vue';
import {useDisplay} from 'vuetify';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';
import BlockchainModeTextDialog from '@/components/tools/workspaces/BlockchainModeTextDialog.vue';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';

const route = useRoute();
const msgStore = useMsgStore();
const display = useDisplay();
const {session} = storeToRefs(useLocalStore());
const context = {addMessage: msgStore.addMessage, $route: route};
const dakarClients = getDakarClients();

const workspaceList = ref([]);
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
		title: 'Blockchain', key: 'mode',
	},
	{
		title: 'Last modification', key: 'modTimeUnix',
	},
	{
		title: '', key: 'actions', sortable: false, align: 'end',
	},
];

const maxWorkspaceNameLength = 50;

// Computed
const authPerMode = computed(() => Object.values(BLOCKCHAIN_ATTRIBUTES).filter(m => isPrivilegedIdentity(session.value, m.mode)
	|| isAdminIdentity(session.value, m.mode)));

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

async function renameWorkspace(workspace) {
	const workspaceName = workspace;
	if (workspaceName === '') {
		setErrorMessage('workspace name must not be empty');
		return;
	}

	if (workspaceName.length > maxWorkspaceNameLength) {
		setErrorMessage(`workspace name is longer than the maximum of ${maxWorkspaceNameLength} characters`);
		return;
	}

	const {mode} = renamedWorkspace.value;

	if (mode === '') {
		setErrorMessage('workspace mode is empty');
		return;
	}

	const workspaceUID = renamedWorkspace.value.uid;

	if (workspaceUID === '') {
		setErrorMessage('workspace UID is not set');
		return;
	}

	isLoading.value = true;

	try {
		await dakarClients[mode].workspace.workspacesRenamePost({
			workspace: {name: workspaceName, workspaceUID},
		});
		msgStore.resetMessages();
		await refreshWorkspaceList();
	} catch (e) {
		handleError(context, e);
	}

	isLoading.value = false;
}

async function addWorkspace(name, mode) {
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

	if (!mode) {
		setErrorMessage('workspace mode is empty');
		return;
	}

	isLoading.value = true;
	try {
		await dakarClients[mode].workspace.workspacesNamePost({name: workspaceName});
		msgStore.resetMessages();
		await refreshWorkspaceList();
	} catch (e) {
		handleError(context, e);
	}

	isLoading.value = false;
}

async function refreshWorkspaceList() {
	isLoading.value = true;

	workspaceList.value = [];
	const resolved = await Promise.allSettled(authPerMode.value.map(chain => dakarClients[chain.mode].workspace.workspacesGet()));

	const workspaces = [];
	for (const [index, response] of resolved.entries()) {
		if (response.status === 'rejected') {
			handleError(context, response.reason);
			continue;
		}

		if (response.value?.workspaces) {
			workspaces.push(...response.value.workspaces.map(w => {
				w.mode = authPerMode.value[index].mode;
				w.modTimeUnix = new Date(w.ts).getTime();
				return w;
			}));
		}
	}

	workspaceList.value = workspaces;
	isLoading.value = false;
	search.value = '';
}

async function deleteWorkspace() {
	if (!workspaceToDelete.value.mode) {
		setErrorMessage('workspace mode must not be empty');
		return;
	}

	if (!workspaceToDelete.value.uid) {
		setErrorMessage('workspace iod must not be empty');
		return;
	}

	isLoading.value = true;

	try {
		const response = await dakarClients[workspaceToDelete.value.mode].workspace
			.workspacesUidDelete({uid: workspaceToDelete.value.uid});
		if (response.msg) {
			setInfoMessage(response.msg);
		}

		await refreshWorkspaceList();
	} catch (e) {
		setErrorMessage(e);
	}

	isLoading.value = false;
}

function showRenameDialog(workspace) {
	if (isLoading.value) {
		return;
	}

	// Workspace is a ref -> need to convert and clone it
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

</script>

<style scoped>

</style>
