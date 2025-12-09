<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-card
    width="288px"
    :to="to"
  >
    <div style="display: grid">
      <svg
        :id="svgID"
        style="width: 100%; grid-area: 1/1"
      />
      <v-skeleton-loader
        v-if="loading"
        type="image"
        style="grid-area: 1/1"
      />
      <v-icon
        v-if="mode"
        style="position: absolute; right: 5px; top: 5px"
        :icon="BLOCKCHAIN_ATTRIBUTES[mode].icon"
        :color="BLOCKCHAIN_ATTRIBUTES[mode].color"
        size="x-large"
      />
    </div>
    <div class="d-flex flex-column">
      <div
        v-tooltip="{'text': created.toLocaleString(), 'location':'top', 'open-delay': 400}"
        class="text-caption ms-auto v-card-subtitle"
      >
        {{ getRelativeTime(created) }}
      </div>
      <v-card-title class="d-flex justify-space-between align-center">
        <div
          style="text-overflow: ellipsis; overflow: hidden"
          class="me-2"
        >
          {{ title }}
        </div>
        <slot />
      </v-card-title>
    </div>
    <alert :text="errorMsg" />
  </v-card>
</template>

<script setup>

import {
	computed,
	onMounted, onUpdated, ref, useId,
} from 'vue';
import {
	getDakarClients,
	getGraphColorMap,
} from '@/utilities/index.js';
import Alert from '@/components/common/Alert.vue';
import {BLOCKCHAIN_ATTRIBUTES} from '@/constants/index.js';
import NodeGraph from '@/d3Documents/nodeGraph.js';

const componentID = useId();

const props = defineProps({
	mode: {type: String, required: true},
	uid: {type: String, required: true},
	title: {type: String, required: true},
	created: {type: Date, required: true},
	to: {type: Object, required: true},
});

const dakarClients = getDakarClients();
const workspaceData = ref(null);
const errorMsg = ref('');
const loading = ref(true);
let oldUID = '';

const nodeGraph = new NodeGraph(getGraphColorMap(props.mode));

// Computed
const svgID = computed(() => `svg_workspace_card_${componentID}`);

// Hooks
onMounted(() => {
	init();
});

onUpdated(() => {
	if (oldUID === props.uid) {
		return;
	}

	init();
});

// Functions
async function init() {
	workspaceData.value = await getWorkspaceData();
	oldUID = props.uid;

	nodeGraph.setEnableInteractions(false);
	nodeGraph.setEnableThumbnailMode(true);
	nodeGraph.initSvg(svgID.value);
	nodeGraph.addNodes(workspaceData.value);
	nodeGraph.centerGraph();
}

async function getWorkspaceData() {
	loading.value = true;
	let data = [];
	try {
		const response = await dakarClients[props.mode].workspace.workspacesStateUidGet({uid: props.uid});

		if (response.state) {
			data = JSON.parse(response.state);
		}
	} catch (e) {
		errorMsg.value = e.message;
	}

	loading.value = false;

	return data;
}

// Returns the relative time to the current date.
function getRelativeTime(targetDate) {
	const diffInMilliseconds = targetDate - new Date();
	const diffInSeconds = Math.floor(diffInMilliseconds / 1000);
	const secondsInMinute = 60;
	const secondsInHour = 3600;
	const secondsInDay = 86400;
	const secondsInMonth = 2592000; // Approximation of seconds in 30 days
	const secondsInYear = 31536000; // Approximation of seconds in 365 days

	let timeUnit;
	let timeValue;

	if (Math.abs(diffInSeconds) < secondsInMinute) {
		timeUnit = 'second';
		timeValue = diffInSeconds;
	} else if (Math.abs(diffInSeconds) < secondsInHour) {
		timeUnit = 'minute';
		timeValue = Math.floor(diffInSeconds / secondsInMinute);
	} else if (Math.abs(diffInSeconds) < secondsInDay) {
		timeUnit = 'hour';
		timeValue = Math.floor(diffInSeconds / secondsInHour);
	} else if (Math.abs(diffInSeconds) < secondsInMonth) {
		timeUnit = 'day';
		timeValue = Math.floor(diffInSeconds / secondsInDay);
	} else if (Math.abs(diffInSeconds) < secondsInYear) {
		timeUnit = 'month';
		timeValue = Math.floor(diffInSeconds / secondsInMonth);
	}

	return new Intl.RelativeTimeFormat('en').format(timeValue, timeUnit);
}

</script>

<style scoped>

</style>
