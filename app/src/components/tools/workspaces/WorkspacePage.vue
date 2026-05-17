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
      <wiki-tooltip
        description-url="workspaces/workspaces.md"
        :icon="mdiHelpCircleOutline"
        icon-color="primary"
      />
    </icon-title>
    <alert :text="errorMsg" />
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
            :created="new Date(item.raw.modTimeUnix)"
          >
            <v-btn-group density="compact">
              <v-btn
                :icon="mdiRename"
                variant="text"
                @click.stop="e => showRenameDialog(e,item.raw)"
              />
              <v-btn
                :icon="mdiDelete"
                variant="text"
                @click.stop="e => showDeleteWorkspaceDialog(e,item.raw)"
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
	mdiDelete,
	mdiMagnify,
	mdiPlus,
	mdiRename,
	mdiHelpCircleOutline,
	mdiArrowLeft,
	mdiArrowRight,
} from '@mdi/js';
import {
	computed,
	onMounted,
	ref,
	toRaw,
} from 'vue';
import {useDisplay} from 'vuetify';
import {storeToRefs} from 'pinia';
import {
	BLOCKCHAIN_ATTRIBUTES,
	PAGE_TITLE,
	ROUTE_NAME_WORKSPACE_PAGE,
} from '@/constants/index.js';
import {
	getDakarClients,
	isAdminIdentity,
	isPrivilegedIdentity,
} from '@/utilities/index.js';
import IconTitle from '@/components/common/IconTitle.vue';
import TextDialog from '@/components/common/TextDialog.vue';
import ConfirmDialog from '@/components/common/ConfirmDialog.vue';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';
import BlockchainModeTextDialog from '@/components/tools/workspaces/BlockchainModeTextDialog.vue';
import {useLocalStore} from '@/pinia/local.js';
import Alert from '@/components/common/Alert.vue';
import WorkspaceCard from '@/components/tools/workspaces/WorkspaceCard.vue';
import SortSelect from '@/components/common/SortSelect.vue';

const display = useDisplay();
const {session} = storeToRefs(useLocalStore());
const dakarClients = getDakarClients();

const workspaceList = ref([]);
const showDeleteWorkspaceDialogModel = ref(false);
const showAddWorkspaceDialogModel = ref(false);
const showRenameWorkspaceDialogModel = ref(false);
const workspaceToDelete = ref(null);
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

// Computed
const authPerMode = computed(() => Object.values(BLOCKCHAIN_ATTRIBUTES).filter(m => isPrivilegedIdentity(session.value, m.mode)
	|| isAdminIdentity(session.value, m.mode)));

// Hooks
onMounted(() => {
	document.title = `Workspaces - ${PAGE_TITLE}`;
	refreshWorkspaceList();
	handleSort();
});

// Functions
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
