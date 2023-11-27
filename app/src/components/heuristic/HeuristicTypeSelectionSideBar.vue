<template>
  <side-bar
    v-model="inputVal"
    title="Add Heuristic"
    :icon="mdiShapeSquareRoundedPlus"
    max-width="648px"
  >
    <template #body>
      <v-tabs
        v-if="tabItems"
        v-model="heuristicTabs"
        class="my-4"
        align-tabs="center"
      >
        <v-tab
          v-for="tabItem in tabItems"
          :key="tabItem"
        >
          {{ tabItem }}
        </v-tab>
      </v-tabs>
      <v-window
        v-if="tabItems"
        v-model="heuristicTabs"
        class="mt-3"
        :touch="false"
      >
        <v-window-item
          v-for="tabItem in tabItems"
          :key="tabItem"
        >
          <div class="d-flex flex-wrap justify-center mx-2">
            <v-card
              v-for="(item, index) in heuristicTypes.filter((e) => {
                if(!e.category && tabItem === 'Other')
                  return true;
                return e.category === tabItem
              })"
              :key="index"
              variant="outlined"
              class="mx-2 mb-4 d-flex flex-column"
              max-width="300"
            >
              <v-card-title>
                {{ item.title }}
              </v-card-title>
              <v-card-text>
                {{ item.description }}
              </v-card-text>
              <v-card-text>
                <v-form
                  v-if="item.parameter !== undefined"
                  v-model="item.parameter.valid"
                >
                  <v-text-field
                    v-model="item.parameter.value"
                    :rules="parameterRules.get(item.parameter.type)"
                    :label="item.parameter.description"
                    required
                  />
                </v-form>
                <v-checkbox
                  v-model="item.useCustomClusters"
                  label="Use custom clusters"
                  hide-details
                />
                <v-checkbox
                  v-model="item.useAddressExclusionList"
                  label="Use address exclusion list"
                  hide-details
                />
                <v-checkbox
                  v-model="item.excludeSpendingGaps"
                  label="Exclude spending gaps"
                  hide-details
                />
              </v-card-text>
              <v-card-actions
                class="pt-0"
              >
                <v-btn
                  class="ms-auto"
                  variant="outlined"
                  @click="addNewHeuristicAction(item)"
                >
                  Add
                </v-btn>
              </v-card-actions>
            </v-card>
          </div>
        </v-window-item>
      </v-window>
    </template>
  </side-bar>
</template>

<script setup>
import {mdiShapeSquareRoundedPlus} from '@mdi/js';
import SideBar from '@/components/heuristic/SideBar.vue';
import {computed, ref} from 'vue';
import {useDisplay} from 'vuetify';

const display = useDisplay();

const props = defineProps({
	modelValue: {type: Boolean, required: true},
	tabItems: {type: Array, required: true},
	descriptors: {type: Array, required: true},
});

const emit = defineEmits(['update:modelValue', 'add-heuristic']);

const parameterRules = new Map([
	['int', [v => {
		if (!/^\d+$/.test(v)) {
			return false;
		}

		const num = parseInt(v, 10);
		return Number.isInteger(num) && num > 0;
	}]],
	// String rule is not implemented yet
	['string', null],
]);

const heuristicTabs = ref(null);

const heuristicTypes = computed(() => props.descriptors.map(descriptor => {
	// Extend descriptor objects with default values for the switches
	descriptor.useCustomClusters = false;
	descriptor.useAddressExclusionList = false;
	descriptor.excludeSpendingGaps = false;
	return descriptor;
}));

const inputVal = computed({
	get() {
		return props.modelValue;
	},
	set(val) {
		emit('update:modelValue', val);
	},
});

// Functions
function addNewHeuristicAction(item) {
	if (item.parameter !== undefined && !item.parameter.valid) {
		return;
	}

	// Hide component on small screen devices after adding a new heuristic
	if (display.smAndDown.value) {
		inputVal.value = false;
	}

	emit('add-heuristic', item);
}

</script>

<style scoped>

</style>
