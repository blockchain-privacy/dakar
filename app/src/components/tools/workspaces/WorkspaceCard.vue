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
      <div class="text-caption ms-auto v-card-subtitle">
        {{ subtitle }}
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
	subtitle: {type: String, required: true},
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

</script>

<style scoped>

</style>
