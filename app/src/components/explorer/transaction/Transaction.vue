<template>
  <v-card :variant="embed?undefined:'text'">
    <icon-title
      v-if="showTitleBar"
      class="ma-2"
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
      <template
        v-if="showTitleLink"
        #title
      >
        <!-- use slot so link does not span the word 'Transaction' and the actual transaction hash -->
        Transaction
        <router-link
          class="shorten ms-1"
          style="color: inherit;"
          :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: tx.txhash, blockchainMode: getSettings.blockchainMode }}"
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
                <router-link :to="{ name: ROUTE_NAME_BLOCK_PAGE, params: { id: tx.bid, blockchainMode: getSettings.blockchainMode }}">
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
                <router-link :to="{ name: ROUTE_NAME_BLOCK_PAGE, params: { id: tx.bhash, blockchainMode: getSettings.blockchainMode }}">
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
            <div
              v-show="enoughDataForInputGraph"
              style="flex: 1 1 500px"
            >
              <p class="text-subtitle-1 text-center">
                Input Distribution
              </p>
              <svg :id="`transaction_inputs_canvas_${tx.txhash}`" />
            </div>
            <!-- empty element in case input graph is hidden but output graph is not -->
            <div
              v-if="!enoughDataForInputGraph"
              style="flex: 1 1 500px"
            />
            <div
              v-show="enoughDataForOutputGraph"
              style="flex: 1 1 500px"
            >
              <p class="text-subtitle-1 text-center">
                Output Distribution
              </p>
              <svg :id="`transaction_outputs_canvas_${tx.txhash}`" />
            </div>
            <!-- empty element in case output graph is hidden but input graph is not -->
            <div
              v-if="!enoughDataForOutputGraph"
              style="flex: 1 1 500px"
            />
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
          <wiki-tooltip
            class="me-1"
            description-url="wasabi/denominations.md"
          >
            Uncommon Wasabi 2.0 denomination
          </wiki-tooltip> detected. Highlight all Wasabi 2.0 denominations?
          <v-spacer />
          <v-switch
            v-model="highlightWasabi2Denominations"
            class="ml-2"
            inset
            density="compact"
            hide-details
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
        <v-tab
          :disabled="!allInputs?.length"
          :text="`${tx.inputs?tx.inputs.length:0} ${plural('Input',tx.inputs?tx.inputs.length:0)}`"
          value="inputs"
        />
        <v-tab
          :text="`${tx.outputs.length} ${plural('Output',tx.outputs.length)}`"
          value="outputs"
        />
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
            margin="100"
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
                :sig-asm="i.sigasm"
                :key-asm="i.keyasm"
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
            <template #empty>
              <!-- needed so no scrollbars appear -->
              <p style="height: 3px" />
            </template>
          </v-infinite-scroll>
        </component>
        <!-- empty col if no inputs exist -->
        <v-col
          v-else
          class="emptyCol"
        />
        <component
          :is="outputFrameComponentColumn"
          value="outputs"
        >
          <v-infinite-scroll
            margin="100"
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
                :sig-asm="i.sigasm"
                :key-asm="i.keyasm"
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
            <template #empty>
              <!-- needed so no scrollbars appear -->
              <p style="height: 3px" />
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
	mdiChevronUp, mdiFormatHeaderPound,
	mdiFormatListNumbered,
	mdiPickaxe, mdiSigma,
	mdiTransfer,
} from '@mdi/js';
import OutputItem from './OutputItem.vue';
import {
	convertAmount,
	getColorMap,
	isDestination,
	isModeBTC, isUncommonWasabi2Denomination,
	plural,
	shortenHash,
} from '@/utilities';
import {ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import IconItem from '../../common/IconItem.vue';
import {
	computed, isProxy, ref, toRaw, toRef, isRef, onUpdated, onMounted, watch, nextTick, useTemplateRef, onUnmounted,
} from 'vue';
import PrivacyChip from '@/components/common/PrivacyChip.vue';
import IconTitle from '@/components/common/IconTitle.vue';
import FingerprintChip from '@/components/explorer/transaction/FingerprintChip.vue';
import BarChart from '@/d3Documents/barChart.js';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';
import {useExplorerStore} from '@/pinia/explorer.js';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';
import {
	VRow, VCol, VTabsWindowItem, VTabsWindow,
} from 'vuetify/components';

const props = defineProps({
	tx: {type: Object, required: true},
	showHeuristicEditorLink: {type: Boolean, required: true},
	showFingerprintLink: {type: Boolean, required: true},
	showDetails: {type: Boolean, required: false},
	showTitleLink: {type: Boolean, required: false},
	embed: {type: Boolean, required: false},
	showTitleBar: {type: Boolean, required: false},
	highlightTransaction: {type: String, required: false, default: ''},
	filterHighlightedOutputs: {type: Boolean, required: false},
});

const {getSettings} = storeToRefs(useLocalStore());
const {highlightWasabi2Denominations} = storeToRefs(useExplorerStore());

const showTransactionDetails = toRef(props.showDetails);

let svgInputGraph = null;
let svgOutputGraph = null;
const colorMap = getColorMap(getSettings.value.blockchainMode);
// Set color for non-privacy transaction
colorMap.set(undefined, '#607D8B');
const enoughDataForInputGraph = ref(true);
const enoughDataForOutputGraph = ref(true);

const showMaxInputs = ref(3);
const showMaxOutputs = ref(3);

const outputContainerRef = useTemplateRef('outputContainer');

const isTabMode = ref(false);
const tabs = ref('inputs');
let resizeObserver;
// Computed
const filteredInputs = computed(() => props.tx.inputs
	.filter(i => Boolean(i.highlight) || (Boolean(props.highlightTransaction) && props.highlightTransaction === i.txhash)));
const filteredOutputs = computed(() => props.tx.outputs
	.filter(i => Boolean(i.highlight) || (Boolean(props.highlightTransaction) && props.highlightTransaction === i.txhash)));

const allInputs = computed(() => props.filterHighlightedOutputs ? filteredInputs.value : props.tx.inputs);
const allOutputs = computed(() => props.filterHighlightedOutputs ? filteredOutputs.value : props.tx.outputs);

const displayedInputs = computed(() => sortByTimestamp(allInputs).slice(0, showMaxInputs.value));
const displayedOutputs = computed(() => sortByTimestamp(allOutputs).slice(0, showMaxOutputs.value));

const inputSum = computed(() => props.tx.inputs?.reduce((sum, input) => sum + input.amount, 0) || 0);
const outputSum = computed(() => props.tx.outputs?.reduce((sum, input) => sum + input.amount, 0) || 0);
const isBTC = computed(() => isModeBTC(getSettings.value.blockchainMode));
const hasUncommonWasabi2Denomination = computed(() => isBTC.value
	&& (props.tx.inputs?.some(i => isUncommonWasabi2Denomination(i.amount))
	|| props.tx.outputs?.some(o => isUncommonWasabi2Denomination(o.amount))));

const outputFrameComponent = computed(() => isTabMode.value ? VTabsWindow : VRow);
const outputFrameComponentColumn = computed(() => isTabMode.value ? VTabsWindowItem : VCol);
// Hooks
onUpdated(() => {
	init();
});

onMounted(() => {
	resizeObserver = new ResizeObserver(entries => {
		const {width} = entries[0].contentRect;

		if (width < 1000) {
			isTabMode.value = true;

			if (allInputs.value?.length > 0) {
				tabs.value = 'inputs';
			} else {
				tabs.value = 'outputs';
			}
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

// Functions
function init() {
	updateInputGraph();
	updateOutputGraph();
}

function updateInputGraph() {
	if (!props.tx.inputs) {
		enoughDataForInputGraph.value = false;
		return;
	}

	svgInputGraph = new BarChart(`transaction_inputs_canvas_${props.tx.txhash}`, 600, 150, false);
	svgInputGraph.drawStacked(props.tx.inputs, colorMap);
	enoughDataForInputGraph.value = !svgInputGraph.empty;
}

function updateOutputGraph() {
	if (!props.tx.outputs) {
		enoughDataForOutputGraph.value = false;
		return;
	}

	svgOutputGraph = new BarChart(`transaction_outputs_canvas_${props.tx.txhash}`, 600, 150, false);
	svgOutputGraph.drawStacked(props.tx.outputs, colorMap);
	enoughDataForOutputGraph.value = !svgOutputGraph.empty;
}

function sortByTimestamp(outputs) {
	if (!outputs) {
		return [];
	}

	let copiedOutputs;

	if (isProxy(outputs)) {
		copiedOutputs = structuredClone(toRaw(outputs));
	} else if (isRef(outputs)) {
		// If 'outputs' is a computed value we need to convert each array element from proxy to raw
		copiedOutputs = structuredClone(outputs.value.map(toRaw));
	} else {
		copiedOutputs = structuredClone(outputs);
	}

	return copiedOutputs.sort((a, b) => {
		if (!a.ts || !b.ts) {
			return true;
		}

		return new Date(a.ts) - new Date(b.ts);
	});
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
