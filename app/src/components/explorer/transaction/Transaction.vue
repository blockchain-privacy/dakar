<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-card :variant="embed?undefined:'text'">
    <icon-title
      v-if="showTitleBar"
      class="pa-2"
      :title="`Transaction ${tx.txhash}`"
      :icon="mdiTransfer"
    >
      <privacy-chip
        v-if="tx.txtype"
        :transaction-type="tx.txtype"
        class="ms-2"
      />
      <fingerprint-chip
        v-if="showFingerprintLink && isDestination(tx.txtype)"
        :transaction-hash="tx.txhash"
        class="ms-2"
      />
      <mode-chip
        v-if="showMode"
        class="ms-2"
        :blockchain-mode="route.params.blockchainMode"
      />
      <template
        v-if="showTitleLink"
        #title
      >
        <!-- use slot so link does not span the word 'Transaction' and the actual transaction hash -->
        Transaction
        <router-link
          class="ms-1"
          style="color: inherit;"
          :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: tx.txhash, blockchainMode: route.params.blockchainMode }}"
        >
          {{ tx.txhash }}
        </router-link>
      </template>
    </icon-title>
    <v-card-text>
      <v-expand-transition>
        <div v-if="showTransactionDetails">
          <v-row>
            <v-col
              cols="12"
              sm="6"
            >
              <icon-item
                :icon="mdiFormatListNumbered"
                title="Block Height"
              >
                <router-link :to="{ name: ROUTE_NAME_BLOCK_PAGE, params: { id: tx.bid, blockchainMode: route.params.blockchainMode }}">
                  {{ tx.bid.toLocaleString() }}
                </router-link>
              </icon-item>
            </v-col>
            <v-col>
              <icon-item
                :icon="mdiCalendar"
                title="Timestamp"
              >
                {{ new Date(tx.bts).toLocaleString() }}
              </icon-item>
            </v-col>
          </v-row>
          <v-row>
            <v-col
              v-if="(tx.fee || tx.fee === 0) && tx.fee >= 0"
              cols="12"
              sm="6"
            >
              <icon-item
                :icon="mdiCash"
                title="Fee"
              >
                {{ convertAmount(tx.fee) }}
              </icon-item>
            </v-col>
            <v-col>
              <icon-item
                :icon="mdiFormatHeaderPound"
                title="Block"
              >
                <router-link :to="{ name: ROUTE_NAME_BLOCK_PAGE, params: { id: tx.bhash, blockchainMode: route.params.blockchainMode }}">
                  {{ shortenHash(tx.bhash) }}
                </router-link>
              </icon-item>
            </v-col>
          </v-row>
          <v-row>
            <v-col
              cols="12"
              sm="6"
            >
              <icon-item
                title="Input Sum"
                :icon="mdiSigma"
              >
                {{ convertAmount(inputSum) }}
              </icon-item>
            </v-col>
            <v-col>
              <icon-item
                title="Output Sum"
                :icon="mdiSigma"
              >
                {{ convertAmount(outputSum) }}
              </icon-item>
            </v-col>
          </v-row>
          <v-row v-if="isCoinBaseTx(tx)">
            <v-col>
              <icon-item
                :icon="mdiPickaxe"
                title="Coinbase"
              >
                yes
              </icon-item>
            </v-col>
          </v-row>
          <div class="d-flex flex-wrap">
            <div style="flex: 1 1 500px">
              <div
                v-if="tx.inputs?.some(d => d.amount)"
                class="mx-4 mb-2"
              >
                <p class="text-body-large text-center">
                  Input Amount Distribution
                </p>
                <amount-chart :outputs="tx.inputs" />
              </div>
              <div v-show="enoughDataForInputGraph">
                <p class="text-body-large text-center">
                  Input Timeline
                </p>
                <svg :id="`transaction_inputs_canvas_${tx.txhash}_${componentID}`" />
              </div>
            </div>
            <div style="flex: 1 1 500px">
              <div
                v-if="tx.outputs?.some(d => d.amount)"
                class="mx-4 mb-2"
              >
                <p class="text-body-large text-center">
                  Output Amount Distribution
                </p>
                <amount-chart :outputs="tx.outputs" />
              </div>
              <div v-show="enoughDataForOutputGraph">
                <p class="text-body-large text-center">
                  Output Timeline
                </p>
                <svg :id="`transaction_outputs_canvas_${tx.txhash}_${componentID}`" />
              </div>
            </div>
          </div>
          <!-- bottom spacer for transition -->
          <div style="height: 10px" />
        </div>
      </v-expand-transition>
      <!-- use btn as width reference because it always exists -->
      <v-btn
        ref="outputContainer"
        variant="text"
        block
        size="small"
        style="margin-top:-16px;"
        @click="showTransactionDetails = !showTransactionDetails"
      >
        <v-icon>{{ showTransactionDetails ? mdiChevronUp : mdiChevronDown }}</v-icon>
      </v-btn>
      <v-alert
        v-if="hasUncommonWasabi2Denomination"
        color="info"
        variant="tonal"
        density="compact"
        class="mb-5"
      >
        <div class="d-flex align-center">
          <div>
            <wiki-tooltip
              class="me-1"
              description-url="wasabi/denominations.md"
            >
              Uncommon Wasabi 2.0 denomination
            </wiki-tooltip> detected. Highlight all Wasabi 2.0 denominations?
          </div>
          <v-spacer />
          <!-- need to set min-width so the switch does not shrink when less space is available -->
          <v-switch
            v-model="highlightWasabi2Denominations"
            class="ms-2"
            inset
            density="compact"
            hide-details
            min-width="55px"
          />
        </div>
      </v-alert>
      <v-tabs
        v-model="tabs"
        grow
        :disabled="!isTabMode"
        :hide-slider="!isTabMode"
        mandatory
      >
        <v-row>
          <v-col class="d-flex">
            <output-sort
              v-if="tx.inputs?.length > 1"
              v-model="inputSortAndFilterModel"
              :transaction-types="inputTransactionTypes"
            />
            <v-tab
              class="flex-grow-1"
              :disabled="!allInputs?.length"
              :text="inputTabTitle"
              value="inputs"
            />
          </v-col>
          <v-col class="d-flex">
            <output-sort
              v-if="tx.outputs?.length > 1"
              v-model="outputSortAndFilterModel"
              :transaction-types="outputTransactionTypes"
            />
            <v-tab
              class="flex-grow-1"
              :text="outputTabTitle"
              value="outputs"
            />
          </v-col>
        </v-row>
      </v-tabs>
      <component
        :is="outputFrameComponent"
        v-model="tabs"
      >
        <component
          :is="outputFrameComponentColumn"
          v-if="tx.inputs && allInputs.length > 0"
          value="inputs"
        >
          <v-infinite-scroll
            ref="inputScroll"
            margin="100"
            empty-text=""
            @load="showMoreInputs"
          >
            <template
              v-for="(i,y) in displayedInputs"
              :key="i.addresshash + i.inputindex"
            >
              <output-item
                is-input
                :amount="i.amount"
                :address-hash="i.addresshash"
                :tx-hash="i.txhash"
                :output-index="i.outputindex"
                :input-index="i.inputindex"
                :timestamp="i.ts"
                :transaction-type="i.txtype"
                :highlight="Boolean(i.highlight) || (Boolean(highlightTransaction) && highlightTransaction === i.txhash)"
              />
              <v-divider
                v-if="y+1 < displayedInputs.length"
                :thickness="2"
              />
            </template>
            <template #loading>
              <!-- empty -->
            </template>
          </v-infinite-scroll>
        </component>
        <!-- empty col if no inputs exist -->
        <v-col v-else />
        <component
          :is="outputFrameComponentColumn"
          value="outputs"
        >
          <v-infinite-scroll
            ref="outputScroll"
            margin="100"
            empty-text=""
            @load="showMoreOutputs"
          >
            <template
              v-for="(i,y) in displayedOutputs"
              :key="i.addresshash + i.outputindex"
            >
              <output-item
                :is-input="false"
                :amount="i.amount"
                :address-hash="i.addresshash"
                :tx-hash="i.txhash"
                :output-index="i.outputindex"
                :input-index="i.inputindex"
                :timestamp="i.ts"
                :transaction-type="i.txtype"
                :highlight="Boolean(i.highlight) || (Boolean(highlightTransaction) && highlightTransaction === i.txhash)"
              />
              <v-divider
                v-if="y+1<displayedOutputs.length"
                :thickness="2"
              />
            </template>
            <template #loading>
              <!-- empty -->
            </template>
          </v-infinite-scroll>
        </component>
      </component>
    </v-card-text>
  </v-card>
</template>

<script setup>
import {
	mdiCalendar,
	mdiCash,
	mdiChevronDown,
	mdiChevronUp,
	mdiFormatHeaderPound,
	mdiFormatListNumbered,
	mdiPickaxe,
	mdiSigma,
	mdiTransfer,
} from '@mdi/js';
import {
	computed,
	ref,
	toRef,
	onUpdated,
	onMounted,
	watch,
	nextTick,
	useTemplateRef,
	onUnmounted,
	useId,
} from 'vue';
import {storeToRefs} from 'pinia';
import {
	VRow,
	VCol,
	VTabsWindowItem,
	VTabsWindow,
} from 'vuetify/components';
import {useRoute} from 'vue-router';
import IconItem from '../../common/IconItem.vue';
import OutputItem from './OutputItem.vue';
import {
	convertAmount,
	getTransactionColorMap,
	isDestination,
	isModeBTC,
	isUncommonWasabi2Denomination,
	plural,
	setUndefinedTransactionColor,
	shortenHash,
} from '@/utilities';
import {ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import PrivacyChip from '@/components/common/PrivacyChip.vue';
import IconTitle from '@/components/common/IconTitle.vue';
import FingerprintChip from '@/components/explorer/transaction/FingerprintChip.vue';
import BarChart from '@/d3Documents/barChart.js';
import {useExplorerStore} from '@/pinia/explorer.js';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';
import AmountChart from '@/components/explorer/transaction/AmountChart.vue';
import ModeChip from '@/components/common/ModeChip.vue';
import OutputSort from '@/components/explorer/transaction/OutputSort.vue';

const props = defineProps({
	tx: {type: Object, required: true},
	showHeuristicEditorLink: {type: Boolean, required: true},
	showFingerprintLink: {type: Boolean, required: true},
	showDetails: {type: Boolean, required: false},
	showTitleLink: {type: Boolean, required: false},
	showMode: {type: Boolean, required: false},
	embed: {type: Boolean, required: false},
	showTitleBar: {type: Boolean, required: false},
	highlightTransaction: {type: String, required: false, default: ''},
	filterHighlightedOutputs: {type: Boolean, required: false},
});
const componentID = useId();
const route = useRoute();
const {highlightWasabi2Denominations} = storeToRefs(useExplorerStore());

const inputScroll = useTemplateRef('inputScroll');
const outputScroll = useTemplateRef('outputScroll');

const showTransactionDetails = toRef(props.showDetails);

let svgInputGraph = null;
let svgOutputGraph = null;

const colorMap = getTransactionColorMap(route.params.blockchainMode);
setUndefinedTransactionColor(colorMap, undefined);
const enoughDataForInputGraph = ref(true);
const enoughDataForOutputGraph = ref(true);

const maxOutputCountDefault = 3;

const showMaxInputs = ref(maxOutputCountDefault);
const showMaxOutputs = ref(maxOutputCountDefault);

const inputSortAndFilterModel = ref({});
const outputSortAndFilterModel = ref({});

const outputContainerRef = useTemplateRef('outputContainer');

const isTabMode = ref(false);
// Needs to be null, so initial resize observer can set correct value
const tabs = ref(null);
let resizeObserver;
// Computed
const inputTransactionTypes = computed(() => getTransactionFilter(props.tx.inputs));
const outputTransactionTypes = computed(() => getTransactionFilter(props.tx.outputs));

const allInputs = computed(() => filterOutputs(props.tx.inputs, inputSortAndFilterModel));
const allOutputs = computed(() => filterOutputs(props.tx.outputs, outputSortAndFilterModel));

const displayedInputs = computed(() => allInputs.value.slice(0, showMaxInputs.value));
const displayedOutputs = computed(() => allOutputs.value.slice(0, showMaxOutputs.value));

const inputSum = computed(() => props.tx.inputs?.reduce((sum, input) => sum + input.amount, 0) || 0);
const outputSum = computed(() => props.tx.outputs?.reduce((sum, input) => sum + input.amount, 0) || 0);
const isBTC = computed(() => isModeBTC(route.params.blockchainMode));
const hasUncommonWasabi2Denomination = computed(() => isBTC.value
	&& (props.tx.inputs?.some(i => isUncommonWasabi2Denomination(i.amount))
		|| props.tx.outputs?.some(o => isUncommonWasabi2Denomination(o.amount))));

const outputFrameComponent = computed(() => isTabMode.value ? VTabsWindow : VRow);
const outputFrameComponentColumn = computed(() => isTabMode.value ? VTabsWindowItem : VCol);

const inputTabTitle = computed(() => {
	const allInputsLength = props.tx.inputs ? props.tx.inputs.length : 0;

	let title = `${allInputsLength} ${plural('Input', allInputsLength)}`;

	if (allInputs.value.length < allInputsLength) {
		title = `${allInputs.value.length} of ` + title;
	}

	return title;
});

const outputTabTitle = computed(() => {
	const allOutputsLength = props.tx.outputs ? props.tx.outputs.length : 0;

	let title = `${allOutputsLength} ${plural('Output', allOutputsLength)}`;

	if (allOutputs.value.length < allOutputsLength) {
		title = `${allOutputs.value.length} of ` + title;
	}

	return title;
});

// Hooks
onUpdated(() => {
	init();
});

onMounted(() => {
	resizeObserver = new ResizeObserver(entries => {
		const {width} = entries[0].contentRect;

		if (width < 1000) {
			isTabMode.value = true;

			tabs.value ||= allInputs.value?.length > 0 ? 'inputs' : 'outputs';
		} else {
			isTabMode.value = false;
			tabs.value = null;
		}
	});

	resizeObserver.observe(outputContainerRef.value.$el);
	init();
});

onUnmounted(() => {
	resizeObserver.disconnect();
});

watch(showTransactionDetails, newVal => {
	if (newVal) {
		// Wait until DOM is updated
		nextTick(() => init());
	}
});

watch(allInputs, newVal => {
	if (newVal?.length === 0) {
		tabs.value = 'outputs';
	}
});

watch(() => props.filterHighlightedOutputs, () => {
	resetInputs();
	resetOutputs();
});

watch(() => inputSortAndFilterModel.value, () => {
	resetInputs();
});

watch(() => outputSortAndFilterModel.value, () => {
	resetOutputs();
});

// Functions
function init() {
	updateInputGraph();
	updateOutputGraph();
}

function resetInputs() {
	inputScroll.value?.reset();
	showMaxInputs.value = maxOutputCountDefault;
}

function resetOutputs() {
	outputScroll.value?.reset();
	showMaxOutputs.value = maxOutputCountDefault;
}

function filterOutputs(outputs, sortAndFilter) {
	if (!outputs) {
		return [];
	}

	let filtered = outputs;
	if (props.filterHighlightedOutputs) {
		filtered = filtered.filter(i => Boolean(i.highlight) || (Boolean(props.highlightTransaction) && props.highlightTransaction === i.txhash));
	}

	if (sortAndFilter.value.filter) {
		filtered = filtered.filter(i => sortAndFilter.value.filter.includes(i.txtype) || (!i.txtype && sortAndFilter.value.filter.includes('other')));
	}

	if (filtered.length < 2) {
		return filtered;
	}

	const sortBy = sortAndFilter.value.sortValue ? sortAndFilter.value.sortValue.value : 'time'; // Fallback to time

	if (sortBy === 'time') {
		return filtered.toSorted((a, b) => {
			if (!a.ts || !b.ts) {
				return 0;
			}

			if (sortAndFilter.value.sortDescending) {
				return new Date(b.ts) - new Date(a.ts);
			}

			return new Date(a.ts) - new Date(b.ts);
		});
	}

	if (sortBy === 'amount') {
		return filtered.toSorted((a, b) => {
			if (a.amount === undefined || b.amount === undefined) {
				return 0;
			}

			if (sortAndFilter.value.sortDescending) {
				return b.amount - a.amount;
			}

			return a.amount - b.amount;
		});
	}

	return filtered.toSorted((a, b) => {
		const txType1 = a.txtype || 'other';
		const txType2 = b.txtype || 'other';

		if (sortAndFilter.value.sortDescending) {
			return txType2.localeCompare(txType1);
		}

		return txType1.localeCompare(txType2);
	});
}

// Returns an array only containing the transaction types present in outputs
function getTransactionFilter(outputs) {
	if (!outputs) {
		return [];
	}

	const cMap = getTransactionColorMap(route.params.blockchainMode);
	const filteredColorMap = new Map();

	outputs.forEach(o => {
		if (o.txtype) {
			filteredColorMap.set(o.txtype, cMap.get(o.txtype));
		} else {
			setUndefinedTransactionColor(filteredColorMap, 'other');
		}
	});

	return [...filteredColorMap].map(d => ({text: d[0], color: d[1]}));
}

function updateInputGraph() {
	if (!props.tx.inputs) {
		enoughDataForInputGraph.value = false;
		return;
	}

	svgInputGraph = new BarChart(`transaction_inputs_canvas_${props.tx.txhash}_${componentID}`, 600, 150);
	svgInputGraph.drawStacked(props.tx.inputs, colorMap);
	enoughDataForInputGraph.value = !svgInputGraph.empty;
}

function updateOutputGraph() {
	if (!props.tx.outputs) {
		enoughDataForOutputGraph.value = false;
		return;
	}

	svgOutputGraph = new BarChart(`transaction_outputs_canvas_${props.tx.txhash}_${componentID}`, 600, 150);
	svgOutputGraph.drawStacked(props.tx.outputs, colorMap);
	enoughDataForOutputGraph.value = !svgOutputGraph.empty;
}

function isCoinBaseTx(tx) {
	if (!tx || !tx.outputs) {
		return false;
	}

	return !tx.inputs || tx.inputs.length === 0;
}

function showMoreInputs({done}) {
	if (allInputs.value.length === 0 || showMaxInputs.value >= allInputs.value.length || props.embed) {
		done('empty');
		return;
	}

	showMaxInputs.value += 15;
	done('ok');
}

function showMoreOutputs({done}) {
	if (allOutputs.value.length === 0 || showMaxOutputs.value >= allOutputs.value.length || props.embed) {
		done('empty');
		return;
	}

	showMaxOutputs.value += 15;
	done('ok');
}

</script>

<style scoped>

/* css for d3 graph  */
:deep(.bar) {
  fill: rgb(var(--v-theme-primary));
}

:deep(.hide) {
  display: none;
  height: 0;
}

:deep(.v-tab) {
  opacity: 1 !important;
}

</style>
