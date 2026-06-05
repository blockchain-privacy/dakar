<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-card
    width="288px"
    :to="importStatus === WORKSPACE_IMPORT_STATUS_WAITING?undefined:to"
  >
    <div
      style="display: grid"
      class="mb-2"
    >
      <svg
        :id="svgID"
        style="width: 100%; grid-area: 1/1"
      />
      <v-skeleton-loader
        v-if="loading"
        type="image"
        style="grid-area: 1/1"
      />
      <v-chip
        v-if="loading && importStatus === WORKSPACE_IMPORT_STATUS_WAITING"
        style="position: absolute; align-self: center; justify-self: center"
        rounded
        color="warning"
      >
        Importing workspace
      </v-chip>
      <v-icon
        v-if="mode"
        style="position: absolute; right: 5px; top: 5px"
        :icon="BLOCKCHAIN_ATTRIBUTES[mode].icon"
        :color="BLOCKCHAIN_ATTRIBUTES[mode].color"
        size="x-large"
      />
    </div>
    <div class="d-flex flex-column">
      <div class="d-flex">
        <v-chip
          v-if="displayImportResult"
          v-tooltip="{'text': `Error while importing the workspace. Some nodes may be missing.`, 'location':'top', 'open-delay': 400}"
          rounded
          size="small"
          :color="importChipColor"
          :prepend-icon="mdiImport"
          class="ms-2 me-auto"
        >
          {{ importResultLabel }}
        </v-chip>
        <v-chip
          v-tooltip="{'text': `Modified ${created.toLocaleString()}`, 'location':'top', 'open-delay': 400}"
          rounded
          size="small"
          :prepend-icon="mdiCalendar"
          class="me-2 ms-auto"
        >
          {{ relativeTime }}
        </v-chip>
      </div>

      <div class="d-flex justify-space-between align-center justify-center px-2">
        <div
          style="text-overflow: ellipsis; overflow: hidden; white-space: nowrap"
          class="me-2 text-title-large"
        >
          {{ title }}
        </div>
        <div class="flex-shrink-0">
          <slot />
        </div>
      </div>
    </div>
    <alert :text="errorMsg" />
  </v-card>
</template>

<script setup>

import {
	computed,
	onMounted,
	onUnmounted,
	onUpdated,
	ref,
	useId,
} from 'vue';
import {mdiCalendar, mdiImport} from '@mdi/js';
import {
	getDakarClients,
	getGraphColorMap,
} from '@/utilities/index.js';
import Alert from '@/components/common/Alert.vue';
import {
	BLOCKCHAIN_ATTRIBUTES,
	WORKSPACE_IMPORT_STATUS_ERROR,
	WORKSPACE_IMPORT_STATUS_SUCCESS,
	WORKSPACE_IMPORT_STATUS_WAITING,
} from '@/constants/index.js';
import NodeGraph from '@/d3Documents/nodeGraph.js';
import {useCacheStore} from '@/pinia/cache.js';

const componentID = useId();

const props = defineProps({
	mode: {type: String, required: true},
	uid: {type: String, required: true},
	title: {type: String, required: true},
	created: {type: Date, required: true},
	to: {type: Object, required: true},
	importStatus: {type: String, required: false, default: ''},
	importTime: {type: Date, required: false, default: null},
});

const cacheStore = useCacheStore();
const dakarClients = getDakarClients();
const workspaceData = ref(null);
const errorMsg = ref('');
const loading = ref(true);
const computeUpdate = ref(0);
let oldUID = '';
let oldImportStatus = '';
let intervalHandle = null;
const nodeGraph = new NodeGraph(getGraphColorMap(props.mode));

// Computed
const svgID = computed(() => `svg_workspace_card_${componentID}`);
const relativeTime = computed(() => {
	// See mounted hook
	const _ = computeUpdate.value;
	return getRelativeTime(props.created);
});

const importChipColor = computed(() => {
	if (props.importStatus === WORKSPACE_IMPORT_STATUS_SUCCESS) {
		return 'success';
	}

	if (props.importStatus === WORKSPACE_IMPORT_STATUS_ERROR) {
		return 'error';
	}

	return 'grey';
});

const importResultLabel = computed(() => {
	if (props.importStatus === WORKSPACE_IMPORT_STATUS_SUCCESS) {
		return 'Import ok';
	}

	if (props.importStatus === WORKSPACE_IMPORT_STATUS_ERROR) {
		return 'Import error';
	}

	return '';
});

const displayImportResult = computed(() => {
	if (props.importStatus !== WORKSPACE_IMPORT_STATUS_SUCCESS && props.importStatus !== WORKSPACE_IMPORT_STATUS_ERROR) {
		return false;
	}

	if (props.importTime === null) {
		return false;
	}

	// 1 day
	const cutoff = 60 * 60 * 24 * 1000;
	// If the workspace was imported over later than the cutoff, then don't display the result
	return Date.now() - props.importTime.getTime() < cutoff;
});

// Hooks
onMounted(() => {
	intervalHandle = setInterval(() => {
		// Change ref so computed value gets updated
		computeUpdate.value += 1;
	}, 1000);
	init();
});

onUpdated(() => {
	if (oldUID === props.uid && props.importStatus === oldImportStatus) {
		return;
	}

	init();
});

onUnmounted(() => {
	if (intervalHandle !== null) {
		clearInterval(intervalHandle);
	}
});

// Functions
async function init() {
	oldUID = props.uid;
	oldImportStatus = props.importStatus;

	if (props.importStatus === WORKSPACE_IMPORT_STATUS_WAITING) {
		return;
	}

	const svgElement = document.getElementById(svgID.value);
	const cacheValue = cacheStore.getWithMetadata(props.uid);
	// Fetch the workspace, if it is not in the cache or if the workspace is newer than the cache item
	if (cacheValue === undefined || cacheValue.ts < props.created) {
		workspaceData.value = await getWorkspaceData();
		nodeGraph.setEnableInteractions(false);
		nodeGraph.setEnableThumbnailMode(true);
		nodeGraph.initSvg(svgID.value);
		nodeGraph.addNodes(workspaceData.value);
		nodeGraph.centerGraph();
		cacheStore.set(props.uid, svgElement.innerHTML);
		return;
	}

	svgElement.innerHTML = cacheValue.value;
	loading.value = false;
}

async function getWorkspaceData() {
	loading.value = true;
	let data = [];
	try {
		const response = await dakarClients[props.mode].workspace.workspacesStateUidGet({uid: props.uid});

		if (response.state) {
			data = JSON.parse(response.state);
		}
	} catch (error) {
		errorMsg.value = error.message;
	}

	loading.value = false;

	return data;
}

// Returns the relative time to the current date.
function getRelativeTime(targetDate) {
	const diffInMilliseconds = targetDate - Date.now();
	const diffInSeconds = Math.floor(diffInMilliseconds / 1000);
	const secondsInMinute = 60;
	const secondsInHour = 3600;
	const secondsInDay = 86_400;
	const secondsInMonth = 2_592_000; // Approximation of seconds in 30 days
	const secondsInYear = 31_536_000; // Approximation of seconds in 365 days

	let timeUnit;
	let timeValue;

	if (Math.abs(diffInSeconds) < secondsInMinute) {
		timeUnit = 'second';
		timeValue = diffInSeconds;
	} else if (Math.abs(diffInSeconds) < secondsInHour) {
		timeUnit = 'minute';
		timeValue = Math.round(diffInSeconds / secondsInMinute);
	} else if (Math.abs(diffInSeconds) < secondsInDay) {
		timeUnit = 'hour';
		timeValue = Math.round(diffInSeconds / secondsInHour);
	} else if (Math.abs(diffInSeconds) < secondsInMonth) {
		timeUnit = 'day';
		timeValue = Math.round(diffInSeconds / secondsInDay);
	} else if (Math.abs(diffInSeconds) < secondsInYear) {
		timeUnit = 'month';
		timeValue = Math.round(diffInSeconds / secondsInMonth);
	} else {
		timeUnit = 'year';
		timeValue = Math.round(diffInSeconds / secondsInYear);
	}

	return new Intl.RelativeTimeFormat('en').format(timeValue, timeUnit);
}

</script>

<style scoped>

</style>
