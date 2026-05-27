<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <side-bar
    v-model="model"
    :title="title"
    :icon="icon"
    max-width="330px"
    disable-full-screen
    title-one-line
  >
    <template #body>
      <v-form
        validate-on="submit"
        @submit.prevent="addNewSelectorAction"
      >
        <v-card-text>
          <v-alert
            v-if="isBatchMode"
            variant="tonal"
            type="info"
            title="Batch Mode"
            class="mb-2"
          >
            Adding {{ plural(selectorType, 2) }} to compatible nodes. There are {{ parentNodes.length }} nodes selected.
          </v-alert>
          <template v-if="selectorType === SELECTOR_TYPE_HEURISTIC">
            <div class="text-title-small mb-3">
              Find senders and receivers of
              <wiki-tooltip description-url="coinjoin.md">
                CoinJoin
              </wiki-tooltip>
              transactions.
              <br>
              <wiki-tooltip description-url="workspaces/coinJoinHeuristics.md">
                Learn more
              </wiki-tooltip>
            </div>
            <v-select
              v-model="heuristicTypeModel"
              class="mb-3"
              :items="heuristicTypes"
              label="Heuristic Type"
              mandatory
              return-object
              hide-details
              @update:model-value="heuristicOptions.type = heuristicTypeModel.type"
            >
              <template #selection="{ item }">
                <div class="d-flex flex-column">
                  <span>{{ item.title }}</span>
                  <span
                    v-if="item.subtitle"
                    class="text-body-small"
                  >
                    {{ item.subtitle }}
                  </span>
                </div>
              </template>
              <template #subheader="i">
                <named-divider
                  :title="i.props.title"
                  :vertical-margin="0"
                  title-class="text-body-small"
                />
              </template>
              <template #item="{ props: itemProps, item }">
                <v-list-item
                  v-bind="itemProps"
                  :subtitle="item.subtitle"
                />
              </template>
            </v-select>
            <template v-if="heuristicTypeModel">
              <div class="text-title-small my-3">
                {{ heuristicTypeModel.description }}
              </div>
              <v-text-field
                v-if="heuristicTypeModel.parameter !== undefined"
                v-model="heuristicOptions.parameter"
                :rules="heuristicRules"
                :label="heuristicTypeModel.parameter.description"
                required
                :placeholder="heuristicTypeModel.parameter.value"
              />
              <v-checkbox
                v-model="heuristicOptions.clusterTypes"
                label="Use custom clusters"
                hide-details
              />
              <v-checkbox
                v-model="heuristicOptions.excludeSpendingGaps"
                label="Exclude spending gaps"
                hide-details
              />
            </template>
          </template>
          <template v-else-if="selectorType === SELECTOR_TYPE_TX_PROP">
            <div class="text-title-small mb-3">
              Select transactions based on their properties.
              <wiki-tooltip description-url="workspaces/propertySelector.md">
                Learn more
              </wiki-tooltip>
            </div>
            <slider-option
              v-model="txPropOptions.maxItems"
              label="Maximum Stored Results"
              :max="SELECTOR_MAX_ITEMS"
            />
            <named-divider title="Select" />
            <template v-if="parentNodes.length === 0">
              <div class="d-flex justify-center mt-2 text-body-large">
                Time Range
              </div>
              <div class="d-flex justify-center mb-2 text-body-small">
                <v-icon
                  start
                  :icon="mdiInformationOutline"
                /> Maximum Range: 60 days
              </div>
              <div class="d-flex align-center mb-5">
                <date-input
                  v-model="txPropOptions.startDate"
                  :rules="parameterRules.get('date')"
                  :error="startDateError"
                  label="From"
                  @update:model-value="handleDateChange"
                />
                <date-input
                  v-model="txPropOptions.endDate"
                  :rules="parameterRules.get('date')"
                  :error="endDateError"
                  label="To"
                  @update:model-value="handleDateChange"
                />
              </div>
            </template>
            <div
              v-else
              class="text-body-large"
            >
              Using stored transactions of the parent node.
            </div>
            <named-divider title="Filter by Type" />
            <v-checkbox
              v-model="txPropOptions.excludePrivacyTransactions"
              label="Exclude Classified Transactions"
            />
            <v-select
              v-model="txPropOptions.txTypes"
              :disabled="txPropOptions.excludePrivacyTransactions"
              max-width="330px"
              multiple
              label="Transaction Types"
              hide-details
              :items="transactionTypeItems"
            >
              <template #selection="{ item }">
                <color-chip
                  :title="item.title"
                  :color="item.color"
                />
              </template>
              <template #item="i">
                <v-list-item v-bind="i.props">
                  <template #prepend="{isSelected}">
                    <v-checkbox-btn :model-value="isSelected" />
                    <color-sheet
                      :color="i.item.color"
                      class="me-2"
                    />
                  </template>
                </v-list-item>
              </template>
            </v-select>
            <named-divider title="Filter by Amount" />
            <range-option
              v-model="txPropOptions.inputSum"
              placeholder="12.23456"
              label="Transaction Input Sum"
              :rules="parameterRules.get('float')"
            />
            <range-option
              v-model="txPropOptions.outputSum"
              placeholder="12.23456"
              label="Transaction Output Sum"
              :rules="parameterRules.get('float')"
            />
            <range-option
              v-model="txPropOptions.inputRange"
              placeholder="12.23456"
              label="Transaction Inputs"
              :rules="parameterRules.get('float')"
            />
            <range-option
              v-model="txPropOptions.outputRange"
              placeholder="12.23456"
              label="Transaction Outputs"
              :rules="parameterRules.get('float')"
            />
          </template>
          <template v-else-if="selectorType === SELECTOR_TYPE_TX_GRAPH">
            <div class="text-title-small mb-3">
              Select transactions based on their distance to the starting node.
              <br>
              <wiki-tooltip description-url="workspaces/graphSelector.md">
                Learn more
              </wiki-tooltip>
            </div>
            <slider-option
              v-model="txGraphOptions.maxItems"
              class="mb-2"
              label="Maximum Stored Results"
              :max="SELECTOR_MAX_ITEMS"
            />
            <v-divider thickness="2" />
            <div class="text-center text-body-large my-2">
              <label for="traversal_direction">Traversal Direction</label>
            </div>
            <div class="d-flex justify-center">
              <v-btn-toggle
                v-model="traversalDirection"
                rounded="lg"
                mandatory
                variant="text"
                color="primary"
              >
                <v-btn
                  id="traversal_direction"
                  size="small"
                >
                  Backward
                </v-btn>
                <v-btn size="small">
                  Forward
                </v-btn>
              </v-btn-toggle>
            </div>
            <slider-option
              v-model="txGraphOptions.depth"
              label="Traversal Depth"
              :max="5"
            />
            <v-checkbox
              v-model="txGraphOptions.excludePrivacyTransactions"
              label="Exclude Classified Transactions"
            />
          </template>
        </v-card-text>
        <alert :text="errorMsg" />
        <v-card-actions>
          <v-btn
            class="ms-auto"
            variant="outlined"
            type="submit"
          >
            Add
          </v-btn>
        </v-card-actions>
      </v-form>
    </template>
  </side-bar>
</template>

<script setup>
import {useRoute} from 'vue-router';
import {mdiFilterPlus, mdiInformationOutline, mdiShapeCirclePlus} from '@mdi/js';
import {
	computed,
	onMounted,
	onUpdated,
	ref,
	toRaw,
} from 'vue';
import SideBar from '@/components/common/SideBar.vue';
import {
	CLUSTER_TYPE_CUSTOM,
	SELECTOR_MAX_ITEMS,
	SELECTOR_TYPE_HEURISTIC,
	SELECTOR_TYPE_TX_GRAPH,
	SELECTOR_TYPE_TX_PROP,
} from '@/constants/index.js';
import NamedDivider from '@/components/common/NamedDivider.vue';
import DateInput from '@/components/workspace/sidebars/DateInput.vue';
import {
	amountToIntegers,
	capitalize,
	filterDescriptors,
	getCoinJoinTypeCaption,
	getTransactionColorMap,
	plural,
	isMaxLargerThanMin,
} from '@/utilities/index.js';
import ColorChip from '@/components/common/ColorChip.vue';
import ColorSheet from '@/components/common/ColorSheet.vue';
import {blenderPlus, graphPlus} from '@/customIcons/index.js';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';
import SliderOption from '@/components/workspace/sidebars/SliderOption.vue';
import RangeOption from '@/components/workspace/sidebars/RangeOption.vue';
import Alert from '@/components/common/Alert.vue';

const model = defineModel({type: Boolean});
const emit = defineEmits(['add-selectors']);
const route = useRoute();

const props = defineProps({
	selectorType: {type: String, required: true},
	descriptors: {type: Array, required: true},
	parentNodes: {type: Array, required: false, default: () => []},
});

const errorMsg = ref('');
// Heuristic select model
const heuristicTypeModel = ref(null);
// Heuristic select items
const heuristicTypes = ref([]);

// Set direction to backward by default
const traversalDirection = ref(0);

const txPropOptions = ref({
	maxItems: 50,
	startDate: null,
	endDate: null,
	excludePrivacyTransactions: false,
	txTypes: [],
	inputSum: {min: undefined, max: undefined},
	outputSum: {min: undefined, max: undefined},
	inputRange: {min: undefined, max: undefined},
	outputRange: {min: undefined, max: undefined},
});

const txGraphOptions = ref({
	maxItems: 50,
	depth: 2,
	isForward: false,
});

const heuristicOptions = ref({
	clusterTypes: [],
	excludeSpendingGaps: false,
	type: null,
	parameter: '',
});

const startDateError = ref(false);
const endDateError = ref(false);
const maxResultsError = ref(false);
const parameterRules = new Map([
	['int', [v => {
		if (!/^\d+$/v.test(v)) {
			return 'must be a number';
		}

		const num = Number.parseInt(v, 10);
		if (Number.isNaN(num) || !Number.isInteger(num)) {
			return 'must be a number';
		}

		return num > 0 || 'must be at least 1';
	}]],
	['float', [v => {
		if (v === undefined || v === '') {
			return true;
		}

		const num = Number.parseFloat(v, 10);
		if (Number.isNaN(num)) {
			return 'must be a number';
		}

		return num > 0 || 'must be higher than 0';
	}]],
	// String rule is not implemented yet
	['string', null],
	['date', [Boolean]],
]);

const transactionTypeItems = [];

// Hooks
onMounted(() => {
	getTransactionColorMap(route.params.blockchainMode).forEach((v, k) => {
		transactionTypeItems.push({title: capitalize(k), value: k, color: v});
	});
});

onUpdated(() => {
	if (props.selectorType === SELECTOR_TYPE_HEURISTIC) {
		heuristicTypes.value = getHeuristicTypes();
		if (heuristicTypes.value.length > 0) {
			heuristicTypeModel.value = getInitialHeuristicTypeModel(heuristicTypeModel.value, heuristicTypes.value);
			heuristicOptions.value.type = heuristicTypeModel.value?.type;
		} else {
			heuristicTypeModel.value = null;
			heuristicOptions.value.type = null;
		}

		heuristicOptions.value.parameter = '';
	}

	startDateError.value = false;
	endDateError.value = false;
	maxResultsError.value = false;
});

// Computed
const title = computed(() => {
	switch (props.selectorType) {
		case SELECTOR_TYPE_HEURISTIC:
			return 'Add CoinJoin Heuristic';
		case SELECTOR_TYPE_TX_PROP:
			return 'Add Property Selector';
		case SELECTOR_TYPE_TX_GRAPH:
			return 'Add Graph Selector';
		default:
			return 'Add Selector';
	}
});

const icon = computed(() => {
	switch (props.selectorType) {
		case SELECTOR_TYPE_HEURISTIC: return blenderPlus;
		case SELECTOR_TYPE_TX_PROP: return mdiFilterPlus;
		case SELECTOR_TYPE_TX_GRAPH: return graphPlus;
		default: return mdiShapeCirclePlus;
	}
});

const heuristicRules = computed(() => {
	if (!heuristicTypeModel.value?.parameter?.type) {
		return [];
	}

	const rules = parameterRules.get(heuristicTypeModel.value.parameter.type);
	if (heuristicTypeModel.value.parameter.type === 'int') {
		if (heuristicTypeModel.value.parameter.minimum) {
			rules.push(v => Number.parseInt(v, 10) >= heuristicTypeModel.value.parameter.minimum
				|| `Minimum: ${heuristicTypeModel.value.parameter.minimum}`);
		}

		if (heuristicTypeModel.value.parameter.maximum) {
			rules.push(v => Number.parseInt(v, 10) <= heuristicTypeModel.value.parameter.maximum
				|| `Maximum: ${heuristicTypeModel.value.parameter.maximum}`);
		}
	}

	return rules;
});

const isBatchMode = computed(() => props.parentNodes.length > 1);

// Functions

// returns true if the descriptors contain at least two items with different types (CoinJoin implementations)
function hasMultipleTypes(descriptors) {
	let lastCoinJoinImplementation = '';

	for (const d of descriptors) {
		const caption = getCoinJoinTypeCaption(d.type);

		if (lastCoinJoinImplementation && caption !== lastCoinJoinImplementation) {
			return true;
		}

		lastCoinJoinImplementation = caption;
	}

	return false;
}

function getHeuristicTypes() {
	if (!props.descriptors || props.parentNodes.length === 0) {
		return [];
	}

	const descriptors = filterDescriptors(props.descriptors, props.parentNodes, !isBatchMode.value);
	const multipleTypes = hasMultipleTypes(descriptors);

	const selectorItems = [];
	let lastCategory = '';
	descriptors.map(d => {
		d.category ||= 'Other';
		return d;
	})
		.toSorted((a, b) => {
			const comparedCategory = b.category.localeCompare(a.category);

			if (comparedCategory === 0) {
				return a.title.localeCompare(b.title);
			}

			return comparedCategory;
		})
		.forEach(d => {
		// Insert subheaders
			if (d.category !== lastCategory) {
				lastCategory = d.category;
				selectorItems.push({title: d.category, type: 'subheader'});
			}

			d.subtitle = multipleTypes ? getCoinJoinTypeCaption(d.type) : '';

			selectorItems.push(d);
		});

	return selectorItems;
}

function isAmountRangeEmpty(obj) {
	return obj.min === undefined && obj.max === undefined;
}

function handleDateChange() {
	if (txPropOptions.value.startDate === null) {
		txPropOptions.value.startDate = txPropOptions.value.endDate;
	}

	if (txPropOptions.value.endDate === null) {
		txPropOptions.value.endDate = txPropOptions.value.startDate;
	}
}

// Converts the string to a blockchain amount, if the string is empty returns undefined
function getAmount(amount) {
	if (!amount) {
		return undefined;
	}

	return amountToIntegers(Number.parseFloat(amount));
}

// Returns a valid heuristic type. tries to set it to oldHeuristicTypeObject if possible.
function getInitialHeuristicTypeModel(oldHeuristicTypeObject, newHeuristicTypes) {
	if (oldHeuristicTypeObject) {
		const obj = newHeuristicTypes.find(d => d.type === oldHeuristicTypeObject.type);
		if (obj) {
			return obj;
		}
	}

	return newHeuristicTypes.find(d => d.type !== 'subheader' && !d.disabled);
}

function buildHeuristicOptions() {
	if (!heuristicOptions.value.type) {
		setErrorMessage('invalid heuristic type');
		return null;
	}

	const options = structuredClone(toRaw(heuristicOptions.value));
	options.clusterTypes = heuristicOptions.value.clusterTypes?.length > 0 ? [CLUSTER_TYPE_CUSTOM] : [];

	options.parameter &&= `${options.parameter}`;
	return options;
}

// Returns true if the two dates are more than 60 days apart
function isDateRangeToBig(startDate, endDate) {
	// 24 * 60 * 60 * 1000 = 86400000
	const milliSecondsPerDay = 86_400_000;
	return Math.round(Math.abs((endDate - startDate) / milliSecondsPerDay)) > 60;
}

// Checks if besides 'startDate', 'endDate' and 'maxItems' another option is set
function isOptionsEmpty(options) {
	return !Object.keys(options).some(k => k !== 'endDate' && k !== 'startDate' && k !== 'maxItems');
}

function buildTxPropOptions() {
	const options = structuredClone(toRaw(txPropOptions.value));

	if (props.parentNodes.length > 0) {
		delete options.startDate;
		delete options.endDate;
	} else {
		if (!options.startDate || !options.endDate || options.startDate > options.endDate) {
			startDateError.value = true;
			endDateError.value = true;
			return;
		}

		// Set endDate to the end of the day, so the full range of the end date is included
		const endDate = new Date(options.endDate);
		options.endDate = new Date(endDate.setHours(23, 59, 59, 999));

		if (isDateRangeToBig(options.startDate, options.endDate)) {
			startDateError.value = true;
			endDateError.value = true;
			setErrorMessage('time range is larger than 60 days');
			return;
		}

		options.startDate = options.startDate.toISOString();
		options.endDate = options.endDate.toISOString();
	}

	startDateError.value = false;
	endDateError.value = false;

	options.maxItems = Number.parseInt(options.maxItems, 10);
	options.inputSum.min = getAmount(options.inputSum.min);
	options.inputSum.max = getAmount(options.inputSum.max);
	options.outputSum.min = getAmount(options.outputSum.min);
	options.outputSum.max = getAmount(options.outputSum.max);

	options.inputRange.min = getAmount(options.inputRange.min);
	options.inputRange.max = getAmount(options.inputRange.max);
	options.outputRange.min = getAmount(options.outputRange.min);
	options.outputRange.max = getAmount(options.outputRange.max);

	if (!isMaxLargerThanMin(options.inputSum)) {
		setErrorMessage('input sum: minimum of must be smaller or equal to maximum');
		return;
	}

	if (!isMaxLargerThanMin(options.outputSum)) {
		setErrorMessage('output sum: minimum of must be smaller or equal to maximum');
		return;
	}

	if (!isMaxLargerThanMin(options.inputRange)) {
		setErrorMessage('input range: minimum of must be smaller or equal to maximum');
		return;
	}

	if (!isMaxLargerThanMin(options.outputRange)) {
		setErrorMessage('output range: minimum of must be smaller or equal to maximum');
		return;
	}

	if (options.excludePrivacyTransactions) {
		delete options.txTypes;
	} else {
		delete options.excludePrivacyTransactions;
	}

	if (options.txTypes?.length === 0) {
		delete options.txTypes;
	}

	if (isAmountRangeEmpty(options.inputSum)) {
		delete options.inputSum;
	}

	if (isAmountRangeEmpty(options.outputSum)) {
		delete options.outputSum;
	}

	if (isAmountRangeEmpty(options.inputRange)) {
		delete options.inputRange;
	}

	if (isAmountRangeEmpty(options.outputRange)) {
		delete options.outputRange;
	}

	return options;
}

function buildTxGraphOptions() {
	const options = structuredClone(toRaw(txGraphOptions.value));
	options.isForward = traversalDirection.value === 1;
	return options;
}

async function addNewSelectorAction(event) {
	// Check if form is valid
	const res = await event;
	if (!res.valid) {
		return;
	}

	let options;

	switch (props.selectorType) {
		case SELECTOR_TYPE_HEURISTIC:
			options = buildHeuristicOptions();
			break;
		case SELECTOR_TYPE_TX_PROP:
			options = buildTxPropOptions();
			break;
		case SELECTOR_TYPE_TX_GRAPH:
			options = buildTxGraphOptions();
			break;
		default:
			setErrorMessage('invalid selector type');
			return;
	}

	if (!options) {
		return;
	}

	// Check for empty object
	if (props.selectorType !== SELECTOR_TYPE_HEURISTIC && isOptionsEmpty(options)) {
		setErrorMessage('at least one filter must be set');
		return;
	}

	emit('add-selectors', props.selectorType, options, props.parentNodes);
	model.value = false;
}

function setErrorMessage(msg) {
	errorMsg.value = msg;
}

</script>

<style scoped>

</style>
