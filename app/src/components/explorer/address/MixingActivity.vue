<template>
  <div>
    <v-card variant="text">
      <v-card-text>
        <wiki-tooltip description-url="transactionTypes/privacyTransactions.md">
          Privacy transactions
        </wiki-tooltip>
        which are directly connected to this address show partially the
        <wiki-tooltip description-url="mixingActivity.md">
          mixing activity
        </wiki-tooltip>.
        <v-row class="mt-2">
          <v-col class="d-flex align-center justify-center flex-wrap">
            <p class="v-label me-2">
              Filter by Privacy Type
            </p>
            <!-- selected-class="" is intentionally left blank to avoid a shadow over the chip elements -->
            <v-chip-group
              v-model="selectedPrivacyID"
              :column="true"
              :multiple="true"
              :filter="true"
              :mandatory="true"
              :disabled="!activities || activities.length === 0"
              selected-class=""
              color="primary"
              @update:model-value="updateSvgData(false)"
            >
              <v-chip
                v-for="label in privacyLabels"
                :key="label.id"
              >
                <template #prepend>
                  <v-sheet
                    style="width:25px; height:15px"
                    rounded
                    :color="label.color?label.color:'black'"
                    class="me-2"
                  />
                </template>
                {{ label.text }}
              </v-chip>
            </v-chip-group>
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
              label="Include cluster addresses"
              hide-details
              :disabled="isLoading"
              @update:model-value="updateSvgData(true)"
            />
          </v-col>
          <v-col
            class="d-flex align-center"
            cols="12"
            lg="9"
          >
            <!-- vuetify lint plugin has some errors, disable for now -->
            <!-- eslint-disable vuetify/no-deprecated-props vuetify/no-deprecated-events -->
            <v-range-slider
              v-model="rangePicker.model"
              :disabled="!activities || activities.length === 0"
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
    <template v-if="hasLoaded && !isLoading">
      <v-card
        v-if="showEmptyResponseMsg && !showTooManyAddressesMsg"
        class="my-3"
        variant="text"
      >
        <v-card-text
          class="text-h6"
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
        <v-card-text class="text-h6 text-center">
          Mixing activity lookup is not possible because the cluster
          of this address is connected to too many addresses
        </v-card-text>
      </v-card>
    </template>
    <v-card
      v-show="activities && activities.length > 0"
      class="my-3"
      variant="text"
    >
      <v-card-text>
        <v-tabs
          v-model="graphTabs"
          :grow="true"
          @update:model-value="onTabChange"
        >
          <v-tab key="histogram">
            Histogram
          </v-tab>
          <v-tab key="graph">
            Force Graph
          </v-tab>
        </v-tabs>
        <v-window
          v-model="graphTabs"
          :touch="false"
        >
          <v-window-item
            key="histogram"
            :eager="true"
          >
            <v-card variant="text">
              <v-card-text>
                <p
                  v-if="showNotEnoughDataMsg && !isLoading"
                  class="text-h6"
                  style="text-align: center"
                >
                  Not enough data available to draw chart
                </p>
                <div
                  v-if="showHistogram"
                  class="text-subtitle-1"
                  style="text-align: center"
                >
                  {{
                    selectedPrivacyLabel.length === 5 ? 'All Privacy'
                    : selectedPrivacyLabel.map(capitalize).join(', ')
                  }}
                  Transactions
                </div>
                <transaction-table-dialog
                  v-model="barTable.show"
                  :headers="barTable.headers"
                  :transactions="barTable.transactions"
                  :start-date="barTable.startDate"
                  :end-date="barTable.endDate"
                />
              </v-card-text>
              <div style="overflow: auto">
                <svg
                  v-show="showHistogram"
                  id="mixing_activity_histogram"
                  style="min-width: 1100px"
                />
              </div>
            </v-card>
          </v-window-item>
          <v-window-item
            key="graph"
            :eager="true"
          >
            <v-card variant="text">
              <v-card-text>
                <div
                  v-if="!overrideTooManyTransactionsWarning && showTooManyTransactionsMsg"
                  style="text-align:center"
                >
                  <v-alert
                    variant="text"
                    :prominent="true"
                    type="warning"
                  >
                    The mixing activity results have more than
                    {{ tooManyTransactionsThreshold }} transactions.
                    Displaying a large number of items in a force graph may severely degrade
                    the performance of your browser.
                    Consider filtering the results by time or privacy type.
                  </v-alert>
                  <v-btn
                    color="primary"
                    @click="showForceGraphDespiteWarning"
                  >
                    Display force graph anyway
                  </v-btn>
                </div>

                <p
                  v-if="!showGraph && !isLoading && !showTooManyTransactionsMsg"
                  class="text-h6"
                  style="text-align: center"
                >
                  No data available
                </p>
                <transaction-dialog
                  v-if="clickedNode"
                  v-model="showNodeDialog"
                  :input-txs="clickedNode.input_txs?clickedNode.input_txs:[]"
                  :date-time="clickedNode.dateTime"
                  :privacy-type="clickedNode.privacytype"
                  :tx-hash="clickedNode.txhash"
                />
              </v-card-text>
              <svg
                v-show="showGraph"
                id="mixing_activity_force_graph"
                style="width:100%; height:500px"
              />
            </v-card>
          </v-window-item>
        </v-window>
      </v-card-text>
    </v-card>
    <v-progress-linear
      v-if="isLoading"
      class="mt-10"
      :indeterminate="true"
    />
  </div>
</template>

<script setup>
import Histogram from '@/d3Documents/histogram';
import ForceGraph from '@/d3Documents/forceGraph';
import {getColorMap, getPrivacyTypeLabel} from '@/utilities';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';
import TransactionTableDialog from '@/components/explorer/address/TransactionTableDialog.vue';
import TransactionDialog from '@/components/explorer/address/TransactionDialog.vue';
import {computed, inject, onBeforeMount, onMounted, ref, watch} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();
const props = defineProps({addressHash: {type: String, required: true}});

const allPrivacyLabels = ['destination', 'collateral creation', 'collateral payment', 'origin', 'mixing'];
const colorMap = getColorMap();
let svgHistogram = null;
let svgGraph = null;
const tooManyTransactionsThreshold = 500;
let initialLoadDone = false;
let graphMode = false;

// Select all labels by default
const selectedPrivacyID = ref([0, 1, 2, 3, 4]);
const showHistogram = ref(false);
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
	{title: 'Transaction Type', key: 'privacytype'}],
	transactions: [],
	startDate: '',
	endDate: '',
	show: false,
});
const clickedNode = ref({
	// eslint-disable-next-line camelcase
	input_txs: [],
	dateTime: null,
	privacytype: '',
	txhash: '',
});
const showNodeDialog = ref(false);
const hasLoaded = ref(false);

watch(() => props.addressHash, () => {
	// Prop was changed -> pull new data
	updateSvgData(true);
});

// Computed
const privacyLabels = computed(() => {
	const labels = [];

	colorMap.forEach((v, k) => {
		labels.push({text: capitalize(k), color: v, id: k});
	});

	return labels;
});

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

const selectedPrivacyLabel = computed(() => {
	const selectedLabels = [];

	selectedPrivacyID.value.forEach(i => {
		selectedLabels.push(allPrivacyLabels[i]);
	});

	return selectedLabels;
});

// Hooks
onBeforeMount(() => {
	includeCusterAddresses.value = false;

	svgHistogram = new Histogram(
		'mixing_activity_histogram',
		1200,
		300,
		false,
	);
	svgHistogram.setClickHandler(onBarClick);
});

onMounted(() => {
	// Has to be called after the SVG is included in the DOM
	svgGraph = new ForceGraph(1200, 500, 'mixing_activity_force_graph', colorMap);
	svgGraph.setClickHandler(onNodeClick);
});

// Functions
// Capitalize returns the first letter of each word (separated by a space) in str capitalized
function capitalize(str) {
	return str.split(' ').map(d => d[0].toUpperCase() + d.slice(1)).join(' ');
}

function setErrorMessage(msg) {
	msgStore.addMessage({text: msg, type: 'error', temporary: true, category: route.name});
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
	} catch (e) {
		if (e.message === 'too_many_addresses') {
			response.msg = e.message;
		}
	}

	return response;
}

function getFilteredData(withLinks) {
	let fromDate = null;
	let toDate = null;
	let considerDate = true;

	if (rangePicker.value.model === null) {
		considerDate = false;
	} else {
		fromDate = new Date(rangePicker.value.model[0]);
		toDate = new Date(rangePicker.value.model[1]);
	}

	const ret = {items: [], links: []};
	const events = new Set();
	const numActivities = activities.value.length;
	ret.items = activities.value.filter(d => {
		if (selectedPrivacyLabel.value.length < 5
        && !selectedPrivacyLabel.value.includes(d.privacytype)) {
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
	Array.from(events).forEach(val => {
		eventObj[val] = '';
	});
	rangePicker.value.events = eventObj;

	if (withLinks) {
		const filteredHashes = new Set(ret.items.map(d => d.txhash));

		ret.items.forEach(d => {
			if (!d.input_txs) {
				return;
			}

			d.input_txs.forEach(it => {
				if (!filteredHashes.has(it.txhash)) {
					return;
				}

				ret.links.push({source: it.txhash, target: d.txhash});
			});
		});
	}

	return ret;
}

function getCategories(filtered) {
	if (selectedPrivacyLabel.value.length < 5) {
		return selectedPrivacyLabel;
	}

	return [...new Set(filtered.map(d => d.privacytype))];
}

function showForceGraphDespiteWarning() {
	overrideTooManyTransactionsWarning.value = true;
	onTabChange(1);
}

function onTabChange(tab) {
	// Tab === 0: histogram
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
	showHistogram.value = false;
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
			d.privacytype = getPrivacyTypeLabel(d.privacytype);
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

	const filtered = getFilteredData(graphMode);

	if (!filtered.items) {
		isLoading.value = false;
		showGraph.value = false;
		showHistogram.value = false;
		showNotEnoughDataMsg.value = true;
		return;
	}

	showTooManyTransactionsMsg.value = filtered.items.length > tooManyTransactionsThreshold;

	// Draw
	if (graphMode) {
		if (!showTooManyTransactionsMsg.value
        || overrideTooManyTransactionsWarning.value) {
			svgGraph.draw(filtered.items, filtered.links);
			showGraph.value = true;
		}
	} else {
		svgHistogram.reset();
		svgHistogram.drawStacked(
			filtered.items,
			getCategories(filtered.items),
			colorMap,
		);
		showHistogram.value = !svgHistogram.empty;
		showNotEnoughDataMsg.value = svgHistogram.empty;
	}

	isLoading.value = false;
}

// Initial load
updateSvgData(true);
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
