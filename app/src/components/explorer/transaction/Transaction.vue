<template>
  <v-card :variant="embed?undefined:'text'">
    <icon-title
      v-if="showTitleBar"
      :title="`Transaction ${tx.txhash}`"
      :icon="mdiTransfer"
      :to="showTitleLink?{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: tx.txhash }}:null"
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
                <router-link :to="{ name: ROUTE_NAME_BLOCK_PAGE, params: { id: tx.bid }}">
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
                <router-link :to="{ name: ROUTE_NAME_BLOCK_PAGE, params: { id: tx.bhash }}">
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
      <v-btn
        variant="text"
        block
        size="small"
        style="margin-top:-16px;"
        @click="showTransactionDetails = !showTransactionDetails"
      >
        <v-icon>{{ showTransactionDetails ? mdiChevronUp : mdiChevronDown }}</v-icon>
      </v-btn>
      <v-row class="outputContainer">
        <v-col v-if="tx.inputs && getInputs.length > 0">
          <p class="text-center">
            {{ `${tx.inputs.length} ${plural('Input',tx.inputs.length)}` }}
          </p>
          <template
            v-for="(i,y) in getInputs"
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
              v-if="y+1 < getInputs.length"
              :thickness="2"
            />
          </template>
          <!-- split in two for nicer transition -->
          <v-expand-transition>
            <div v-if="showAllOutputs">
              <v-divider :thickness="2" />
              <template
                v-for="(i,y) in getResidualInputs"
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
                  v-if="y+1 < getResidualInputs.length"
                  :thickness="2"
                />
              </template>
            </div>
          </v-expand-transition>
        </v-col>
        <!-- empty col if no inputs exist -->
        <v-col
          v-else
          class="emptyCol"
        />
        <v-col v-if="tx.outputs && getOutputs.length > 0">
          <p class="text-center">
            {{ `${tx.outputs.length} ${plural('Output',tx.outputs.length)}` }}
          </p>
          <template
            v-for="(i,y) in getOutputs"
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
              v-if="y+1<getOutputs.length"
              :thickness="2"
            />
          </template>

          <!-- split in two for nicer transition -->
          <v-expand-transition>
            <div v-if="showAllOutputs">
              <v-divider :thickness="2" />
              <template
                v-for="(i,y) in getResidualOutputs"
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
                  v-if="y+1<getResidualOutputs.length"
                  :thickness="2"
                />
              </template>
            </div>
          </v-expand-transition>
        </v-col>
        <!-- empty col if no outputs exist -->
        <v-col
          v-else
          class="emptyCol"
        />
      </v-row>
    </v-card-text>
    <v-btn
      v-if="areItemsLimited"
      variant="text"
      block
      size="x-small"
      @click="showAllOutputs = !showAllOutputs"
    >
      <v-icon>{{ showAllOutputs ? mdiChevronUp : mdiChevronDown }}</v-icon>
    </v-btn>
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
	mdiPickaxe, mdiSigma,
	mdiTransfer,
} from '@mdi/js';
import OutputItem from './OutputItem.vue';
import {
	convertAmount, getColorMap, isDestination, plural, shortenHash,
} from '@/utilities';
import {ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import IconItem from '../../common/IconItem.vue';
import {
	computed, isProxy, ref, toRaw, toRef, isRef, onUpdated, onMounted, watch, nextTick,
} from 'vue';
import PrivacyChip from '@/components/common/PrivacyChip.vue';
import IconTitle from '@/components/common/IconTitle.vue';
import FingerprintChip from '@/components/explorer/transaction/FingerprintChip.vue';
import BarChart from '@/d3Documents/barChart.js';

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

const showTransactionDetails = toRef(props.showDetails);
const showAllOutputs = ref(false);
const maxOutputs = ref(3);

let svgInputGraph = null;
let svgOutputGraph = null;
const colorMap = getColorMap();
// Set color for non-privacy transaction
colorMap.set(undefined, '#607D8B');
const enoughDataForInputGraph = ref(true);
const enoughDataForOutputGraph = ref(true);

// Computed
const filteredInputs = computed(() => props.tx.inputs
	.filter(i => Boolean(i.highlight) || (Boolean(props.highlightTransaction) && props.highlightTransaction === i.txhash)));
const filteredOutputs = computed(() => props.tx.outputs
	.filter(i => Boolean(i.highlight) || (Boolean(props.highlightTransaction) && props.highlightTransaction === i.txhash)));
const getInputs = computed(() => getLimitedItems(sortByTimestamp(props.filterHighlightedOutputs ? filteredInputs : props.tx.inputs)));
const getResidualInputs = computed(() => getResidualItems(sortByTimestamp(props.filterHighlightedOutputs ? filteredInputs : props.tx.inputs)));
const getOutputs = computed(() => getLimitedItems(sortByTimestamp(props.filterHighlightedOutputs ? filteredOutputs : props.tx.outputs)));
const getResidualOutputs = computed(() => getResidualItems(sortByTimestamp(props.filterHighlightedOutputs ? filteredOutputs : props.tx.outputs)));
const areItemsLimited = computed(() => {
	if (!props.tx) {
		return false;
	}

	if (props.tx.inputs && props.tx.inputs.length > maxOutputs.value) {
		return true;
	}

	return Boolean(props.tx.outputs && props.tx.outputs.length > maxOutputs.value);
});
const inputSum = computed(() => props.tx.inputs?.reduce((sum, input) => sum + input.amount, 0) || 0);
const outputSum = computed(() => props.tx.outputs?.reduce((sum, input) => sum + input.amount, 0) || 0);

// Hooks
onUpdated(() => {
	init();
});

onMounted(() => {
	init();
});

watch(showTransactionDetails, newVal => {
	if (newVal) {
		// Wait until DOM is updated
		nextTick(() => init());
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

function getLimitedItems(items) {
	if (!items) {
		return [];
	}

	return items.slice(0, maxOutputs.value);
}

function getResidualItems(items) {
	if (!items) {
		return [];
	}

	if (items.length <= maxOutputs.value) {
		return [];
	}

	if (showAllOutputs.value) {
		return items.slice(maxOutputs.value);
	}

	return [];
}

</script>

<style scoped>

.outputContainer {
  container-type: inline-size;
  container-name: outputContainer;
}

/* don't show col if parent is sm or smaller */
@container outputContainer (width < 960px) {
  .emptyCol {
    display: none;
  }
}

/* css for d3 graph  */
:deep(.bar) {
  fill: rgb(var(--v-theme-primary));
}

:deep(.hide) {
  display: none;
  height: 0;
}

</style>
