<template>
  <side-bar
    v-model="model"
    title="Add Selector"
    :icon="mdiFilterPlus"
    max-width="648px"
  >
    <template #body>
      <v-form
        validate-on="submit"
        @submit.prevent="addNewSelectorAction"
      >
        <v-card-text>
          <v-select
            v-model="selectorTypeModel"
            class="mb-3"
            :items="selectorTypes"
            label="Selector Type"
            mandatory
            hide-details
          />
          <template v-if="selectorTypeModel === SELECTOR_TYPE_HEURISTIC">
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
              <v-form
                v-if="heuristicTypeModel.parameter !== undefined"
                v-model="heuristicTypeModel.parameter.valid"
              >
                <v-text-field
                  v-model="heuristicOptions.parameter"
                  :rules="parameterRules.get(heuristicTypeModel.parameter.type)"
                  :label="heuristicTypeModel.parameter.description"
                  required
                  :placeholder="heuristicTypeModel.parameter.value"
                  hide-details
                />
              </v-form>
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
          <template v-else-if="selectorTypeModel === SELECTOR_TYPE_TX_PROP">
            <div class="d-flex justify-center my-2 text-subtitle-1">
              Time Range
            </div>
            <div class="d-flex align-center">
              <date-input
                v-model="selectorOptions.startDate"
                :rules="parameterRules.get('date')"
              />
              <date-input
                v-model="selectorOptions.endDate"
                :rules="parameterRules.get('date')"
              />
            </div>
            <div class="d-flex justify-center my-2 text-subtitle-1">
              Transaction Input Sum
            </div>
            <div class="d-flex align-center">
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
            <div class="d-flex align-center">
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
            <div class="d-flex align-center">
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
import {onUpdated, ref, toRaw} from 'vue';
import {CLUSTER_TYPE_CUSTOM, SELECTOR_TYPE_HEURISTIC, SELECTOR_TYPE_TX_PROP} from '@/constants/index.js';
import NamedDivider from '@/components/common/NamedDivider.vue';
import DateInput from '@/components/workspace/sidebars/DateInput.vue';
const props = defineProps({descriptors: {type: Array, required: true}});
const model = defineModel({type: Boolean});
const emit = defineEmits(['add-selector']);
const msgStore = useMsgStore();
const route = useRoute();

// Switch model
const selectorTypeModel = ref(SELECTOR_TYPE_HEURISTIC);
// Switch items
const selectorTypes = [
	{title: 'Transaction Heuristic', value: SELECTOR_TYPE_HEURISTIC},
	{title: 'Transaction Properties', value: SELECTOR_TYPE_TX_PROP},
];

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
		.sort((a, b) => a.category < b.category && a.title < b.title)
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

async function addNewSelectorAction(event) {
	const res = await event;
	if (!res.valid) {
		return;
	}

	let options = null;

	if (selectorTypeModel.value === SELECTOR_TYPE_HEURISTIC) {
		if (!heuristicOptions.value.type) {
			setErrorMessage('invalid heuristic type');
			return;
		}

		options = structuredClone(toRaw(heuristicOptions.value));
		options.clusterTypes = heuristicOptions.value.clusterTypes ? [CLUSTER_TYPE_CUSTOM] : [];

		// Int to string
		options.paramter &&= `${options.paramter}`;
	} else if (selectorTypeModel.value === SELECTOR_TYPE_TX_PROP) {
		options = structuredClone(toRaw(selectorOptions.value));

		if (!options.startDate || !options.endDate) {
			setErrorMessage('invalid date range');
			return;
		}

		options.startDate = options.startDate.toISOString();
		options.endDate = options.endDate.toISOString();

		options.inputSum.min &&= parseFloat(options.inputSum.min);
		options.inputSum.max &&= parseFloat(options.inputSum.max);
		options.outputSum.min &&= parseFloat(options.outputSum.min);
		options.outputSum.max &&= parseFloat(options.outputSum.max);

		options.inputRange.min &&= parseFloat(options.inputRange.min);
		options.inputRange.max &&= parseFloat(options.inputRange.max);
		options.outputRange.min &&= parseFloat(options.outputRange.min);
		options.outputRange.max &&= parseFloat(options.outputRange.max);
	} else {
		setErrorMessage('invalid selector type');
		return;
	}

	emit('add-selector', selectorTypeModel.value, options);
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
