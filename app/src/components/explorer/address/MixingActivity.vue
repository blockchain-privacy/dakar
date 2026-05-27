<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div>
    <v-card variant="text">
      <v-card-text>
        Classified transactions which are directly connected to this address show partially the
        <wiki-tooltip description-url="mixingActivity.md">
          mixing activity
        </wiki-tooltip>.
        <v-row
          v-if="activities?.length > 0"
          class="mt-2"
        >
          <v-col class="d-flex align-center justify-center flex-wrap">
            <p class="v-label me-2">
              Filter by Transaction type
            </p>
            <chip-filter
              v-model="chipFilterModel"
              mandatory
              :items="transactionTypes"
              @changed="handleChipFilterChanged"
            />
          </v-col>
        </v-row>
        <v-row class="mt-2">
          <v-col
            class="d-flex align-center"
            cols="12"
            lg="3"
          >
            <v-switch
              v-model="includeCusterAddresses"
              label="Search all addresses of cluster"
              hide-details
              :disabled="isLoading"
              @update:model-value="updateSvgData(true)"
            />
          </v-col>
          <v-col
            v-if="activities?.length > 1"
            class="d-flex align-center"
            cols="12"
            lg="9"
          >
            <v-range-slider
              v-model="rangePicker.model"
              :ticks="rangePicker.events"
              class="mr-8"
              :min="rangePicker.min"
              :max="rangePicker.max"
              label="Filter by time"
              hide-details
              thumb-label="always"
              show-ticks="always"
              track-size="1"
              tick-size="9"
              @end="updateSvgData(false)"
            >
              <template
                v-if="isSameDay"
                #thumb-label="{ modelValue }"
              >
                {{ new Date(modelValue).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) }}
              </template>
              <template
                v-else
                #thumb-label="{ modelValue }"
              >
                {{ new Date(modelValue).toLocaleDateString() }}
              </template>
            </v-range-slider>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
    <alert :text="errorMsg" />
    <template v-if="hasLoaded && !isLoading">
      <v-card
        v-if="showEmptyResponseMsg && !showTooManyAddressesMsg"
        class="my-3"
        variant="text"
      >
        <v-card-text
          class="text-title-large"
          style="text-align:center"
        >
          No mixing activity detected
        </v-card-text>
      </v-card>
      <v-card
        v-if="showTooManyAddressesMsg"
        class="my-3"
        variant="text"
      >
        <v-card-text class="text-title-large text-center">
          Mixing activity lookup is not possible because the cluster
          of this address is connected to too many addresses
        </v-card-text>
      </v-card>
    </template>
    <div v-show="activities && activities.length > 0">
      <v-tabs
        v-model="graphTabs"
        grow
        @update:model-value="onTabChange"
      >
        <v-tab key="barChart">
          Bar Chart
        </v-tab>
        <v-tab key="graph">
          Graph
        </v-tab>
      </v-tabs>
      <v-window
        v-model="graphTabs"
        :touch="false"
        style="line-height: 0"
      >
        <v-window-item
          key="barChart"
          eager
        >
          <v-card variant="text">
            <v-card-text>
              <p
                v-if="showNotEnoughDataMsg && !isLoading"
                class="text-title-large"
                style="text-align: center"
              >
                Not enough data available to draw chart
              </p>
              <transaction-table-dialog
                v-model="barTable.show"
                :headers="barTable.headers"
                :transactions="barTable.transactions"
                :start-date="barTable.startDate"
                :end-date="barTable.endDate"
              />
              <div style="overflow: auto">
                <svg
                  v-show="showBarChart"
                  id="mixing_activity_barchart"
                  style="min-width: 1100px"
                />
              </div>
            </v-card-text>
          </v-card>
        </v-window-item>
        <v-window-item
          key="graph"
          eager
        >
          <v-card
            v-if="!overrideTooManyTransactionsWarning && showTooManyTransactionsMsg"
            variant="text"
          >
            <v-card-text>
              <div style="text-align:center">
                <v-alert
                  variant="text"
                  prominent
                  type="warning"
                >
                  The mixing activity results have more than
                  {{ tooManyTransactionsThreshold }} transactions.
                  Displaying a large number of items in a force graph may severely degrade
                  the performance of your browser.
                  Consider filtering the results by time or transaction type.
                </v-alert>
                <v-btn
                  color="primary"
                  @click="showForceGraphDespiteWarning"
                >
                  Display force graph anyway
                </v-btn>
              </div>
            </v-card-text>
          </v-card>
          <v-card
            v-if="!showGraph && !isLoading && !showTooManyTransactionsMsg"
            variant="text"
          >
            <v-card-text>
              <p
                class="text-title-large"
                style="text-align: center"
              >
                No data available
              </p>
            </v-card-text>
          </v-card>
          <transaction-dialog
            v-if="clickedNode"
            v-model="showNodeDialog"
            :input-txs="clickedNode.input_txs?clickedNode.input_txs:[]"
            :date-time="clickedNode.dateTime"
            :transaction-type="clickedNode.txtype"
            :tx-hash="clickedNode.txhash"
          />
          <div
            v-if="showGraph"
            class="d-flex align-center justify-center my-2"
          >
            <v-card variant="flat">
              <graph-toolbar
                one-line
                disable-filter
                @is-selection-enabled="(flag) => nodeGraph.setLassoEnabled(flag)"
                @center="nodeGraph.centerGraph()"
                @rearrange="nodeGraph.reorderNodes()"
              />
            </v-card>
          </div>
          <svg
            v-show="showGraph"
            id="mixing_activity_force_graph"
            style="width:100%; height:500px;"
          />
        </v-window-item>
      </v-window>
    </div>
    <v-progress-linear
      v-if="isLoading"
      class="mt-10"
      indeterminate
    />
  </div>
</template>

<script setup>
import {
	computed,
	nextTick,
	onBeforeMount,
	onMounted,
	ref,
	toRaw,
	watch,
} from 'vue';
import {useRoute} from 'vue-router';
import {useHotkey} from 'vuetify';
import BarChart from '@/d3Documents/barChart.js';
import {getTransactionColorMap, getDakarClient} from '@/utilities';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';
import TransactionTableDialog from '@/components/explorer/address/TransactionTableDialog.vue';
import TransactionDialog from '@/components/explorer/address/TransactionDialog.vue';
import NodeGraph from '@/d3Documents/nodeGraph.js';
import {WORKSPACE_NODE_TYPE_TRANSACTION} from '@/constants/index.js';
import {useWorkspaceStore} from '@/pinia/workspace.js';
import GraphToolbar from '@/components/common/GraphToolbar.vue';
import ChipFilter from '@/components/explorer/address/ChipFilter.vue';
import {setNodesDisplayAttributes} from '@/d3Documents/nodeDisplay.js';
import Alert from '@/components/common/Alert.vue';

const route = useRoute();
const workspaceStore = useWorkspaceStore();
const props = defineProps({addressHash: {type: String, required: true}});
const dakar = getDakarClient(route.params.blockchainMode);

useHotkey('cmd+a', () => nodeGraph.selectAllNodes());
useHotkey('esc', () => nodeGraph.resetLasso());

const colorMap = getTransactionColorMap(route.params.blockchainMode);
let svgBarChart = null;
const nodeGraph = new NodeGraph(colorMap);
const tooManyTransactionsThreshold = 500;
let initialLoadDone = false;
let graphMode = false;

// Select all labels by default
const showBarChart = ref(false);
const showGraph = ref(false);
const isLoading = ref(false);
const showEmptyResponseMsg = ref(false);
const showTooManyAddressesMsg = ref(false);
const showNotEnoughDataMsg = ref(false);
const showTooManyTransactionsMsg = ref(false);
const overrideTooManyTransactionsWarning = ref(false);
const activities = ref(null);
const includeCusterAddresses = ref(false);
const rangePicker = ref({
	model: null,
	min: 0,
	max: 0,
	events: [],
});
const graphTabs = ref(null);
const barTable = ref({
	headers: [{
		title: 'Transaction', align: 'start', key: 'txhash',
	},
	{title: 'Timestamp', key: 'dateTime'},
	{title: 'Type', key: 'txtype'}],
	transactions: [],
	startDate: '',
	endDate: '',
	show: false,
});
const clickedNode = ref({
	// eslint-disable-next-line camelcase
	input_txs: [],
	dateTime: null,
	txtype: '',
	txhash: '',
});
const showNodeDialog = ref(false);
const hasLoaded = ref(false);
const transactionTypes = [...colorMap.entries()].map(d => ({text: d[0], color: d[1]}));
const chipFilterModel = ref([...transactionTypes.keys()]);
const errorMsg = ref('');

watch(() => props.addressHash, () => {
	// Prop was changed -> pull new data
	updateSvgData(true);
});

// Computed

const selectedTransactionType = computed(() => chipFilterModel.value.map(d => transactionTypes[d].text));

// Returns truer if min and max or on the same calendar day
const isSameDay = computed(() => {
	const day1 = new Date(rangePicker.value.min);
	const day2 = new Date(rangePicker.value.max);
	// Cut off time
	day1.setHours(0, 0, 0, 0);
	day2.setHours(0, 0, 0, 0);
	// Compare numbers
	return day1.getTime() === day2.getTime();
});

// Hooks
onBeforeMount(() => {
	includeCusterAddresses.value = false;

	svgBarChart = new BarChart('mixing_activity_barchart', 1200, 300);
	svgBarChart.setClickHandler(onBarClick);
});

onMounted(() => {
	nodeGraph.initSvg('mixing_activity_force_graph', 1200, 500);

	if (!nodeGraph.setNodeClickCallback(onNodeClick)) {
		setErrorMessage('error setting node click handler');
		return;
	}

	if (workspaceStore.getIsWorkspaceActive) {
		if (!nodeGraph.setLassoSelectionCallback(handleLassoSelection)) {
			setErrorMessage('error setting lasso selection handler');
			return;
		}

		if (!nodeGraph.setLassoResetCallback(handleLassoReset)) {
			setErrorMessage('error setting lasso reset handler');
		}
	}
});

// Functions
function setErrorMessage(msg) {
	errorMsg.value = msg;
}

function onBarClick(data) {
	if (data.x0.getHours() === data.x1.getHours()
		&& data.x0.getMinutes() === data.x1.getMinutes()) {
		barTable.value.startDate = data.x0.toLocaleDateString();
		barTable.value.endDate = data.x1.toLocaleDateString();
	} else {
		barTable.value.startDate = data.x0.toLocaleString();
		barTable.value.endDate = data.x1.toLocaleString();
	}

	barTable.value.transactions = data;
	barTable.value.show = true;
}

function onNodeClick(data) {
	clickedNode.value = data;
	showNodeDialog.value = true;
}

async function getMixingActivity() {
	const response = {ok: false, data: null, msg: null};
	try {
		response.data = await dakar.tools.mixingActivityPost({
			activity: {
				addressHash: props.addressHash,
				isClusterLookup: includeCusterAddresses.value,
			},
		});
		response.ok = true;
	} catch (error) {
		if (error.message === 'too_many_addresses') {
			response.msg = error.message;
		}
	}

	return response;
}

function getFilteredData(withGraphData) {
	let fromDate = null;
	let toDate = null;
	let considerDate = true;

	if (rangePicker.value.model === null) {
		considerDate = false;
	} else {
		fromDate = new Date(rangePicker.value.model[0]);
		toDate = new Date(rangePicker.value.model[1]);
	}

	const events = new Set();
	const numActivities = activities.value.length;

	const items = activities.value.map(d => toRaw(d)).filter(d => {
		if (selectedTransactionType.value.length < transactionTypes.length
			&& !selectedTransactionType.value.includes(d.txtype)) {
			return false;
		}

		if (!considerDate) {
			return true;
		}

		const eventTime = d.dateTime;

		// Decrease accuracy of range picker ticks when many activities exist
		if (numActivities > 500) {
			eventTime.setHours(0, 0, 0, 0);
		} else if (numActivities > 200) {
			eventTime.setMinutes(0, 0, 0);
		}

		events.add(eventTime.getTime());

		return d.dateTime <= toDate && d.dateTime >= fromDate;
	});

	// Construct event objet
	const eventObj = {};
	[...events].forEach(val => {
		eventObj[val] = '';
	});
	rangePicker.value.events = eventObj;

	if (withGraphData) {
		const itemMap = new Map(items.map(d => [d.txhash, d]));

		items.forEach(d => {
			if (!d.input_txs) {
				return;
			}

			d.input_txs.forEach(it => {
				const currentItem = itemMap.get(it.txhash);
				if (!currentItem) {
					return;
				}

				if (currentItem.children === undefined) {
					currentItem.children = [];
				}

				currentItem.children.push(d.txhash);
			});
		});

		// Set children for each item
		items.forEach(d => {
			d.uid = d.txhash;
			d.transactionHash = d.txhash;
			d.type = WORKSPACE_NODE_TYPE_TRANSACTION;

			const currentItem = itemMap.get(d.txhash);
			if (currentItem === undefined || !currentItem.children) {
				return;
			}

			d.children = currentItem.children;
		});
	}

	return items;
}

function showForceGraphDespiteWarning() {
	overrideTooManyTransactionsWarning.value = true;
	onTabChange(1);
}

function handleChipFilterChanged() {
	updateSvgData(false);
}

function handleLassoSelection() {
	workspaceStore.setWorkspaceNodes(nodeGraph.getLassoSelectedNodesData().map(d => ({id: d.uid, type: WORKSPACE_NODE_TYPE_TRANSACTION})));
}

function handleLassoReset() {
	workspaceStore.workspaceNodes.clear();
}

function onTabChange(tab) {
	// Tab === 0: bar chart
	// tab === 1: force graph
	const wantGraph = tab === 1;
	// Check if tab was actually changed. @changed:modelValue also fires on initial load of component
	if (graphMode === wantGraph) {
		return;
	}

	graphMode = wantGraph;

	if (graphMode) {
		if (!showTooManyTransactionsMsg.value
			|| overrideTooManyTransactionsWarning.value) {
			updateSvgData();
		}

		return;
	}

	updateSvgData();
}

async function updateSvgData(pullNewData) {
	showBarChart.value = false;
	showGraph.value = false;
	isLoading.value = true;
	showTooManyAddressesMsg.value = false;
	clickedNode.value = null;
	barTable.value.transactions = [];
	// Check if new data has to be loaded
	if (pullNewData || !initialLoadDone) {
		const mixingActivity = await getMixingActivity();
		hasLoaded.value = true;

		if (!mixingActivity.ok) {
			isLoading.value = false;
			activities.value = [];

			if (mixingActivity.msg === 'too_many_addresses') {
				showTooManyAddressesMsg.value = true;
				initialLoadDone = true;
				return;
			}

			setErrorMessage('error getting mixing activity');
			return;
		}

		if (!mixingActivity.data.activities) {
			showEmptyResponseMsg.value = true;
			isLoading.value = false;
			activities.value = [];
			return;
		}

		// Used to set boundaries for the date picker
		let maxDate = null;
		let minDate = null;

		activities.value = mixingActivity.data.activities.map(d => {
			d.dateTime = new Date(d.block[0].ts);

			if (maxDate === null || d.dateTime > maxDate) {
				maxDate = new Date(d.dateTime);
			}

			if (minDate === null || d.dateTime < minDate) {
				minDate = new Date(d.dateTime);
			}

			return d;
		});

		rangePicker.value.min = new Date(minDate).getTime();
		rangePicker.value.max = new Date(maxDate).getTime();
		rangePicker.value.model = [rangePicker.value.min, rangePicker.value.max];
		initialLoadDone = true;
	}

	showEmptyResponseMsg.value = false;

	const filteredItems = getFilteredData(graphMode);

	if (!filteredItems) {
		isLoading.value = false;
		showGraph.value = false;
		showBarChart.value = false;
		showNotEnoughDataMsg.value = true;
		return;
	}

	showTooManyTransactionsMsg.value = filteredItems.length > tooManyTransactionsThreshold;

	// Draw
	if (graphMode) {
		if (!showTooManyTransactionsMsg.value
			|| overrideTooManyTransactionsWarning.value) {
			showGraph.value = true;
			nodeGraph.removeAllNodes(false);

			// Needed so svg is not still hidden when doing force simulation
			await nextTick(() => {
				nodeGraph.addNodes(setNodesDisplayAttributes(filteredItems));
				nodeGraph.centerGraph();
			});
		}
	} else {
		svgBarChart.reset();
		svgBarChart.drawStacked(filteredItems, colorMap);
		showBarChart.value = !svgBarChart.empty;
		showNotEnoughDataMsg.value = svgBarChart.empty;
	}

	isLoading.value = false;
}

// Initial load
await updateSvgData(true);
</script>

<style scoped>
:deep( .overlay ) {
  stroke-width: 2px;
  stroke: #1976D2;
  fill: #1976D2;
  cursor: pointer;
}

:deep( .nodeMouseOver ) {
  cursor: pointer;
}

:deep( .v-slider-thumb__label ) {
  min-width: 75px;
}

:deep(.v-slider-track__tick--filled) {
  background-color: rgb(var(--v-theme-primary-lighten-2));
}

:deep(.v-slider-track__tick) {
  background-color: rgb(var(--v-theme-primary-lighten-2));
}

</style>
