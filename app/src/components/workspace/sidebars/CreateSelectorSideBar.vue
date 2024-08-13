<template>
  <side-bar
    v-model="model"
    :title="title"
    :icon="mdiFilterPlus"
    max-width="648px"
  >
    <template #body>
      <v-form
        validate-on="submit"
        @submit.prevent="addNewSelectorAction"
      >
        <v-card-text>
          <template v-if="selectorType === SELECTOR_TYPE_HEURISTIC">
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
              <template #item="i">
                <named-divider
                  v-if="i.item.raw.divider !== undefined"
                  :title="i.item.raw.title"
                  :vertical-margin="0"
                  title-class="text-caption"
                />
                <v-list-item
                  v-else
                  v-bind="i.props"
                />
              </template>
            </v-select>
            <template v-if="heuristicTypeModel">
              <div
                class="text-subtitle-2 mb-3"
                style="max-width:260px"
              >
                {{ heuristicTypeModel.description }}
              </div>
              <v-text-field
                v-if="heuristicTypeModel.parameter !== undefined"
                v-model="heuristicOptions.parameter"
                :rules="parameterRules.get(heuristicTypeModel.parameter.type)"
                :label="heuristicTypeModel.parameter.description"
                required
                :placeholder="heuristicTypeModel.parameter.value"
                hide-details
              />
              <v-checkbox
                v-model="heuristicOptions.clusterTypes"
                label="Use custom clusters"
                hide-details
              />
              <v-checkbox
                v-model="heuristicOptions.excludeAddresses"
                label="Use address exclusion list"
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
            <div
              class="text-subtitle-2 mb-3"
              style="max-width:260px"
            >
              Select transactions based on their properties. Results are limited to 50 transactions.
            </div>

            <named-divider title="Select" />

            <template v-if="!hasParent">
              <div class="d-flex justify-center my-2 text-subtitle-1">
                Time Range
              </div>
              <div class="d-flex align-center mb-5">
                <date-input
                  v-model="selectorOptions.startDate"
                  :rules="parameterRules.get('date')"
                  :error="startDateError"
                  label="From"
                  @update:model-value="handleDateChange"
                />
                <date-input
                  v-model="selectorOptions.endDate"
                  :rules="parameterRules.get('date')"
                  :error="endDateError"
                  label="To"
                  @update:model-value="handleDateChange"
                />
              </div>
            </template>
            <div
              v-else
              class="text-subtitle-1"
            >
              Transactions of the parent node will be used
            </div>
            <named-divider title="Filter" />
            <div class="d-flex justify-center my-2 text-subtitle-1">
              Transaction Input Sum
            </div>
            <div class="d-flex align-center mb-5">
              <v-text-field
                v-model="selectorOptions.inputSum.min"
                min-width="100px"
                :rules="parameterRules.get('float')"
                label="From"
                placeholder="12.3456"
                hide-details
                class="me-2"
              />
              <v-text-field
                v-model="selectorOptions.inputSum.max"
                min-width="100px"
                :rules="parameterRules.get('float')"
                label="To"
                placeholder="12.3456"
                hide-details
              />
            </div>
            <div class="d-flex justify-center my-2 text-subtitle-1">
              Transaction Output Sum
            </div>
            <div class="d-flex align-center mb-5">
              <v-text-field
                v-model="selectorOptions.outputSum.min"
                min-width="100px"
                :rules="parameterRules.get('float')"
                label="From"
                placeholder="12.3456"
                hide-details
                class="me-2"
              />
              <v-text-field
                v-model="selectorOptions.outputSum.max"
                min-width="100px"
                :rules="parameterRules.get('float')"
                label="To"
                placeholder="12.3456"
                hide-details
              />
            </div>
            <div class="d-flex justify-center my-2 text-subtitle-1">
              Transaction Inputs
            </div>
            <div class="d-flex align-center mb-5">
              <v-text-field
                v-model="selectorOptions.inputRange.min"
                min-width="100px"
                :rules="parameterRules.get('float')"
                label="From"
                placeholder="12.3456"
                hide-details
                class="me-2"
              />
              <v-text-field
                v-model="selectorOptions.inputRange.max"
                min-width="100px"
                :rules="parameterRules.get('float')"
                label="To"
                placeholder="12.3456"
                hide-details
              />
            </div>
            <div class="d-flex justify-center my-2 text-subtitle-1">
              Transaction Outputs
            </div>
            <div class="d-flex align-center">
              <v-text-field
                v-model="selectorOptions.outputRange.min"
                min-width="100px"
                :rules="parameterRules.get('float')"
                label="From"
                placeholder="12.3456"
                hide-details
                class="me-2"
              />
              <v-text-field
                v-model="selectorOptions.outputRange.max"
                min-width="100px"
                :rules="parameterRules.get('float')"
                label="To"
                placeholder="12.3456"
                hide-details
              />
            </div>
          </template>
        </v-card-text>
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
import {mdiFilterPlus} from '@mdi/js';
import {useMsgStore} from '@/pinia/msg';
import SideBar from '@/components/common/SideBar.vue';
import {
	computed, onUpdated, ref, toRaw,
} from 'vue';
import {CLUSTER_TYPE_CUSTOM, SELECTOR_TYPE_HEURISTIC, SELECTOR_TYPE_TX_PROP} from '@/constants/index.js';
import NamedDivider from '@/components/common/NamedDivider.vue';
import DateInput from '@/components/workspace/sidebars/DateInput.vue';
import {amountToIntegers} from '@/utilities/index.js';

const model = defineModel({type: Boolean});
const emit = defineEmits(['add-selector']);
const msgStore = useMsgStore();
const route = useRoute();

const props = defineProps({
	selectorType: {type: String, required: true},
	descriptors: {type: Array, required: true},
	hasParent: {type: Boolean, required: false, default: false},
});

// Heuristic select model
const heuristicTypeModel = ref([]);
// Heuristic select items
const heuristicTypes = ref([]);

const selectorOptions = ref({
	startDate: null,
	endDate: null,
	inputSum: {min: undefined, max: undefined},
	outputSum: {min: undefined, max: undefined},
	inputRange: {min: undefined, max: undefined},
	outputRange: {min: undefined, max: undefined},
});

const heuristicOptions = ref({
	clusterTypes: [],
	excludeAddresses: false,
	excludeSpendingGaps: false,
	type: null,
});

const startDateError = ref(false);
const endDateError = ref(false);
const parameterRules = new Map([
	['int', [v => {
		if (!/^\d+$/.test(v)) {
			return false;
		}

		const num = parseInt(v, 10);
		return !isNaN(num) && Number.isInteger(num) && num > 0;
	}]],
	['float', [v => {
		if (v === undefined || v === '') {
			return true;
		}

		const num = parseFloat(v, 10);
		return !isNaN(num) && num > 0;
	}]],
	// String rule is not implemented yet
	['string', null],
	['date', [v => Boolean(v)]],
]);

// Hooks
onUpdated(() => {
	heuristicTypes.value = getHeuristicTypes();
	if (heuristicTypes.value.length > 0) {
		heuristicTypeModel.value = heuristicTypes.value.find(d => !d.divider);
		heuristicOptions.value.type = heuristicTypeModel.value.type;
	}

	startDateError.value = false;
	endDateError.value = false;
});

// Computed
const title = computed(() => {
	switch (props.selectorType) {
		case SELECTOR_TYPE_HEURISTIC:
			return 'Add Heuristic';
		case SELECTOR_TYPE_TX_PROP:
			return 'Add Selector';
		default:
			return 'Add Selector';
	}
});

// Functions
function getHeuristicTypes() {
	if (!props.descriptors) {
		return [];
	}

	const selectorItems = [];
	let lastCategory = '';
	props.descriptors
		.map(d => {
			d.category ||= 'Other';

			return d;
		})
		.sort((a, b) => {
			const comparedCategory = a.category.localeCompare(b.category);

			if (comparedCategory === 0) {
				return a.title.localeCompare(b.title);
			}

			return comparedCategory;
		})
		.forEach(d => {
		// Insert dividers
			if (d.category !== lastCategory) {
				lastCategory = d.category;
				selectorItems.push({divider: true, title: d.category});
			}

			selectorItems.push(d);
		});
	return selectorItems;
}

function isAmountRangeEmpty(obj) {
	return obj.min === undefined && obj.max === undefined;
}

function handleDateChange() {
	if (selectorOptions.value.startDate === null) {
		selectorOptions.value.startDate = selectorOptions.value.endDate;
	}

	if (selectorOptions.value.endDate === null) {
		selectorOptions.value.endDate = selectorOptions.value.startDate;
	}
}

// Converts the string to a blockchain amount, if the string is empty returns undefined
function getAmount(amount) {
	if (!amount) {
		return undefined;
	}

	return amountToIntegers(parseFloat(amount));
}

async function addNewSelectorAction(event) {
	const res = await event;
	if (!res.valid) {
		return;
	}

	let options = null;

	switch (props.selectorType) {
		case SELECTOR_TYPE_HEURISTIC:
			if (!heuristicOptions.value.type) {
				setErrorMessage('invalid heuristic type');
				return;
			}

			options = structuredClone(toRaw(heuristicOptions.value));
			options.clusterTypes = heuristicOptions.value.clusterTypes?.length > 0 ? [CLUSTER_TYPE_CUSTOM] : [];

			// Int to string
			options.paramter &&= `${options.paramter}`;
			break;
		case SELECTOR_TYPE_TX_PROP:
			options = structuredClone(toRaw(selectorOptions.value));

			if (props.hasParent) {
				delete options.startDate;
				delete options.endDate;
			} else {
				if (!options.startDate || !options.endDate || options.startDate > options.endDate) {
					startDateError.value = true;
					endDateError.value = true;
					return;
				}

				options.startDate = options.startDate.toISOString();
				options.endDate = options.endDate.toISOString();
			}

			startDateError.value = false;
			endDateError.value = false;

			options.inputSum.min = getAmount(options.inputSum.min);
			options.inputSum.max = getAmount(options.inputSum.max);
			options.outputSum.min = getAmount(options.outputSum.min);
			options.outputSum.max = getAmount(options.outputSum.max);

			options.inputRange.min = getAmount(options.inputRange.min);
			options.inputRange.max = getAmount(options.inputRange.max);
			options.outputRange.min = getAmount(options.outputRange.min);
			options.outputRange.max = getAmount(options.outputRange.max);

			if (isAmountRangeEmpty(options.inputSum) && isAmountRangeEmpty(options.outputSum)
				&& isAmountRangeEmpty(options.inputRange) && isAmountRangeEmpty(options.outputRange)) {
				setErrorMessage('at least one filter must be set');
				return;
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

			break;
		default:
			setErrorMessage('invalid selector type');
			return;
	}

	emit('add-selector', props.selectorType, options);
	model.value = false;
}

function setErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: true, category: route.name,
	});
}

</script>

<style scoped>

</style>
