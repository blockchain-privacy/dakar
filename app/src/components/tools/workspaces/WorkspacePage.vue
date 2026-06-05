<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
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
      <wiki-tooltip
        description-url="workspaces/workspaces.md"
        :icon="mdiHelpCircleOutline"
        icon-color="primary"
      />
    </icon-title>
    <alert
      :text="errorMsg"
      closable
    />
    <alert
      :text="infoMsg"
      type="info"
    />
    <v-data-iterator
      v-if="workspaceList.length > 0"
      v-model:page="iteratorPageModel"
      :items="workspaceList"
      :items-per-page="8"
      :search="search"
      :sort-by="sortBy"
    >
      <template #header>
        <div class="d-flex justify-space-between flex-wrap">
          <v-text-field
            v-model="search"
            placeholder="Search workspaces"
            :prepend-inner-icon="mdiMagnify"
            max-width="300px"
            min-width="200px"
            variant="outlined"
            clearable
            hide-details
            class="me-2 mb-2"
          />
          <sort-select
            v-model:sort="sort"
            v-model:direction="direction"
            class="mb-2"
            :items="sortItems"
            style="max-width: 300px; min-width: 200px;"
            @update:sort="handleSort"
            @update:direction="handleSort"
          />
        </div>
      </template>
      <template #default="{ items }">
        <div
          class="d-flex flex-wrap mt-2 align-center mb-5 justify-center"
          style="gap: 15px"
        >
          <workspace-card
            v-for="item in items"
            :key="item.raw.uid"
            :to="{ name: ROUTE_NAME_WORKSPACE_PAGE, params: { id: item.raw.uid, blockchainMode: item.raw.mode }}"
            :uid="item.raw.uid"
            :mode="item.raw.mode"
            :title="item.raw.name"
            :import-status="item.raw.importStatus"
            :import-time="new Date(item.raw.importTs)"
            :created="new Date(item.raw.modTimeUnix)"
          >
            <v-btn-group density="compact">
              <v-btn
                :disabled="item.raw.importStatus === WORKSPACE_IMPORT_STATUS_WAITING"
                :icon="mdiRename"
                variant="text"
                @click.stop="e => showRenameDialog(e, item.raw)"
              />
              <v-btn
                :disabled="item.raw.importStatus === WORKSPACE_IMPORT_STATUS_WAITING"
                :icon="mdiExport"
                variant="text"
                @click.stop="e => showExportDialog(e, item.raw)"
              />
              <v-btn
                :disabled="item.raw.importStatus === WORKSPACE_IMPORT_STATUS_WAITING"
                :icon="mdiDelete"
                variant="text"
                @click.stop="e => showDeleteWorkspaceDialog(e, item.raw)"
              />
            </v-btn-group>
          </workspace-card>
        </div>
      </template>
      <template #footer="{ page, pageCount, prevPage, nextPage }">
        <div class="d-flex align-center justify-center pa-4">
          <v-btn
            :disabled="page === 1"
            density="comfortable"
            :icon="mdiArrowLeft"
            variant="tonal"
            rounded
            @click="prevPage"
          />
          <div class="mx-2 text-body-small">
            Page {{ page }} of {{ pageCount }}
          </div>
          <v-btn
            :disabled="page >= pageCount"
            density="comfortable"
            :icon="mdiArrowRight"
            variant="tonal"
            rounded
            @click="nextPage"
          />
        </div>
      </template>
    </v-data-iterator>
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
      <p class="text-body-large">
        Workspace <code>{{ workspaceToDelete.name }}</code>
        will be deleted. Continue?
      </p>
    </confirm-dialog>
    <add-workspace-dialog
      v-if="showAddWorkspaceDialogModel"
      v-model="showAddWorkspaceDialogModel"
      :maxlength="maxWorkspaceNameLength"
      @added="addWorkspace"
      @imported="importWorkspace"
    />
    <export-dialog
      v-if="showExportWorkspaceDialogModel"
      v-model="showExportWorkspaceDialogModel"
      :workspace="workspaceToExport"
      @submit="exportWorkspaceHandler"
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
	mdiDelete,
	mdiMagnify,
	mdiPlus,
	mdiRename,
	mdiHelpCircleOutline,
	mdiArrowLeft,
	mdiArrowRight,
	mdiExport,
} from '@mdi/js';
import {
	computed,
	onMounted,
	onUnmounted,
	ref,
	toRaw,
} from 'vue';
import {useDisplay} from 'vuetify';
import {storeToRefs} from 'pinia';
import {
	BLOCKCHAIN_ATTRIBUTES,
	BLOCKCHAIN_BTC,
	BLOCKCHAIN_DASH,
	PAGE_TITLE,
	ROUTE_NAME_WORKSPACE_PAGE,
	WORKSPACE_IMPORT_STATUS_WAITING,
} from '@/constants/index.js';
import {
	getCurrentDate,
	getDakarClients,
	isAdminIdentity,
	isPrivilegedIdentity,
} from '@/utilities/index.js';
import IconTitle from '@/components/common/IconTitle.vue';
import TextDialog from '@/components/common/TextDialog.vue';
import ConfirmDialog from '@/components/common/ConfirmDialog.vue';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';
import AddWorkspaceDialog from '@/components/tools/workspaces/AddWorkspaceDialog.vue';
import {useLocalStore} from '@/pinia/local.js';
import Alert from '@/components/common/Alert.vue';
import WorkspaceCard from '@/components/tools/workspaces/WorkspaceCard.vue';
import SortSelect from '@/components/common/SortSelect.vue';
import ExportDialog from '@/components/tools/workspaces/ExportDialog.vue';

const display = useDisplay();
const {session} = storeToRefs(useLocalStore());
const dakarClients = getDakarClients();

const workspaceList = ref([]);
const showDeleteWorkspaceDialogModel = ref(false);
const showAddWorkspaceDialogModel = ref(false);
const showRenameWorkspaceDialogModel = ref(false);
const showExportWorkspaceDialogModel = ref(false);
const workspaceToDelete = ref(null);
const workspaceToExport = ref(null);
const renamedWorkspace = ref(null);
const isLoading = ref(false);
const search = ref('');
const sortBy = ref([]);
const errorMsg = ref('');
const infoMsg = ref('');
const iteratorPageModel = ref(1);

const sortItems = [
	{value: 'name', title: 'Name'},
	{value: 'mode', title: 'Blockchain'},
	{value: 'modTimeUnix', title: 'Last modification'},
];

const sort = ref(sortItems[2]);
const direction = ref(true);

const maxWorkspaceNameLength = 50;
let reloadTimer = null;

// Computed
const authPerMode = computed(() => Object.values(BLOCKCHAIN_ATTRIBUTES).filter(m => isPrivilegedIdentity(session.value, m.mode)
	|| isAdminIdentity(session.value, m.mode)));

// Hooks
onMounted(async () => {
	document.title = `Workspaces - ${PAGE_TITLE}`;
	await refreshWorkspaceList();
	handleSort();
	startWaitingForReload(workspaceList.value);
});

onUnmounted(() => {
	if (reloadTimer !== null) {
		clearTimeout(reloadTimer);
		reloadTimer = null;
	}
});

// Functions
function startWaitingForReload(workspaces) {
	if (reloadTimer !== null) {
		// Already checking for reload
		return;
	}

	const waitingWorkspace = workspaces.find(w => w.importStatus === WORKSPACE_IMPORT_STATUS_WAITING);
	if (waitingWorkspace) {
		reloadTimer = setTimeout(refreshWorkspaceListHidden, 3000);
	}
}

async function renameWorkspace(workspace) {
	errorMsg.value = '';
	const workspaceName = workspace;
	if (workspaceName === '') {
		errorMsg.value = 'workspace name must not be empty';
		return;
	}

	if (workspaceName.length > maxWorkspaceNameLength) {
		errorMsg.value = `workspace name is longer than the maximum of ${maxWorkspaceNameLength} characters`;
		return;
	}

	const {mode} = renamedWorkspace.value;

	if (mode === '') {
		errorMsg.value = 'workspace mode is empty';
		return;
	}

	const workspaceUID = renamedWorkspace.value.uid;

	if (workspaceUID === '') {
		errorMsg.value = 'workspace UID is not set';
		return;
	}

	isLoading.value = true;

	try {
		await dakarClients[mode].workspace.workspacesRenamePost({
			workspace: {name: workspaceName, workspaceUID},
		});

		await refreshWorkspaceList();
	} catch (error) {
		errorMsg.value = error.message;
	}

	isLoading.value = false;
}

async function exportWorkspaceHandler(workspace, option) {
	errorMsg.value = '';
	showExportWorkspaceDialogModel.value = false;

	if (!workspace) {
		errorMsg.value = 'workspace empty';
		return;
	}

	if (!workspace.mode) {
		errorMsg.value = 'workspace mode is empty';
		return;
	}

	isLoading.value = true;

	switch (option) {
		case 'workspace': {
			await exportWorkspace(workspace);
			break;
		}

		case 'entities': {
			await exportEntities(workspace);
			break;
		}

		default: {
			errorMsg.value = 'invalid export option';
		}
	}

	isLoading.value = false;
}

async function exportWorkspace(workspace) {
	try {
		const response = await dakarClients[workspace.mode].workspace.workspacesExportPost({workspace: {workspaceUID: workspace.uid}});

		const a = document.createElement('a');
		a.href = URL.createObjectURL(response);

		a.setAttribute(
			'download',
			`workspace_export_${getCurrentDate()}_${workspace.name}.json`,
		);
		a.click();
		a.remove();
	} catch (error) {
		errorMsg.value = error.message;
	}
}

async function exportEntities(workspace) {
	try {
		const response = await dakarClients[workspace.mode].workspace.workspacesExportEntitiesPost({workspace: {workspaceUID: workspace.uid}});

		const a = document.createElement('a');
		a.href = URL.createObjectURL(response);

		a.setAttribute(
			'download',
			`workspace_entities_export_${getCurrentDate()}_${workspace.name}.csv`,
		);
		a.click();
		a.remove();
	} catch (error) {
		errorMsg.value = error.message;
	}
}

async function addWorkspace(name, mode) {
	errorMsg.value = '';
	showAddWorkspaceDialogModel.value = false;
	const workspaceName = name.trim();
	if (workspaceName === '') {
		errorMsg.value = 'workspace name must not be empty';
		return;
	}

	if (workspaceName.length > maxWorkspaceNameLength) {
		errorMsg.value = `workspace name is longer than the maximum of ${maxWorkspaceNameLength} characters`;
		return;
	}

	if (!mode) {
		errorMsg.value = 'workspace mode is empty';
		return;
	}

	isLoading.value = true;
	try {
		await dakarClients[mode].workspace.workspacesNamePost({name: workspaceName});
		await refreshWorkspaceList();
	} catch (error) {
		errorMsg.value = error.message;
	}

	isLoading.value = false;
}

async function importWorkspace(file) {
	errorMsg.value = '';
	showAddWorkspaceDialogModel.value = false;

	isLoading.value = true;
	try {
		const blockchainMode = JSON.parse(await file.text()).meta?.blockchainMode;

		if (blockchainMode !== BLOCKCHAIN_DASH && blockchainMode !== BLOCKCHAIN_BTC) {
			errorMsg.value = `invalid blockchain mode while importing: '${blockchainMode}'`;
			return;
		}

		await dakarClients[blockchainMode].workspace.workspacesImportPost({file});
		await refreshWorkspaceList();
	} catch (error) {
		errorMsg.value = error.message;
	} finally {
		isLoading.value = false;
	}
}

async function refreshWorkspaceList() {
	isLoading.value = true;
	errorMsg.value = '';

	workspaceList.value = [];
	const resolved = await Promise.allSettled(authPerMode.value.map(chain => dakarClients[chain.mode].workspace.workspacesGet()));

	const workspaces = [];
	for (const [index, response] of resolved.entries()) {
		if (response.status === 'rejected') {
			errorMsg.value = response.reason.message;
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
	startWaitingForReload(workspaceList.value);
}

async function refreshWorkspaceListHidden() {
	const resolved = await Promise.allSettled(authPerMode.value.map(chain => dakarClients[chain.mode].workspace.workspacesGet()));

	const workspaces = [];
	for (const [index, response] of resolved.entries()) {
		if (response.status === 'rejected') {
			errorMsg.value = response.reason.message;
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
	reloadTimer = null;
	startWaitingForReload(workspaceList.value);
}

async function deleteWorkspace() {
	errorMsg.value = '';
	infoMsg.value = '';
	if (!workspaceToDelete.value.mode) {
		errorMsg.value = 'workspace mode must not be empty';
		return;
	}

	if (!workspaceToDelete.value.uid) {
		errorMsg.value = 'workspace uid must not be empty';
		return;
	}

	isLoading.value = true;

	try {
		const response = await dakarClients[workspaceToDelete.value.mode].workspace
			.workspacesUidDelete({uid: workspaceToDelete.value.uid});
		if (response.msg) {
			infoMsg.value = response.msg;
		}

		await refreshWorkspaceList();
	} catch (error) {
		errorMsg.value = error.message;
	}

	isLoading.value = false;
}

function showRenameDialog(e, workspace) {
	e.preventDefault();
	if (isLoading.value) {
		return;
	}

	// Workspace is a ref -> need to convert and clone it
	renamedWorkspace.value = structuredClone(toRaw(workspace));
	showRenameWorkspaceDialogModel.value = true;
}

function showExportDialog(e, workspace) {
	e.preventDefault();
	if (isLoading.value) {
		return;
	}

	workspaceToExport.value = workspace;
	showExportWorkspaceDialogModel.value = true;
}

function showDeleteWorkspaceDialog(e, workspace) {
	e.preventDefault();
	if (isLoading.value) {
		return;
	}

	showDeleteWorkspaceDialogModel.value = true;
	workspaceToDelete.value = workspace;
}

function handleSort() {
	sortBy.value = [{key: sort.value.value, order: direction.value ? 'desc' : 'asc'}];
	// Show first page of data iterator
	iteratorPageModel.value = 1;
}

</script>

<style scoped>

</style>
