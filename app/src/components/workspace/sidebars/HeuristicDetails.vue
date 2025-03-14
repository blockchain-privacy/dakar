<template>
  <v-card variant="text">
    <v-card-text>
      <div class="d-flex align-center flex-wrap justify-center">
        <v-card
          color="primary"
          variant="flat"
          class="me-4"
          min-width="150px"
        >
          <v-card-text>
            <div class="text-h4">
              {{ transactionCount.toLocaleString() }}
            </div>
            <div class="text-subtitle-1">
              {{ plural('Transaction', transactionCount) }}
            </div>
          </v-card-text>
        </v-card>
        <v-card
          color="primary"
          variant="flat"
          min-width="150px"
        >
          <v-card-text>
            <div class="text-h4">
              {{ heuristicData.clusterCount.toLocaleString() }}
            </div>
            <div class="text-subtitle-1">
              {{ plural('Cluster', heuristicData.clusterCount) }}
            </div>
          </v-card-text>
        </v-card>
      </div>
      <named-divider
        title="Properties"
        title-class="text-subtitle-1"
      />
      <div class="d-flex align-center flex-wrap itemContainer justify-center">
        <small-icon-item
          :title="heuristicData.heuristicTypeTitle"
          :icon="mdiApplicationVariableOutline"
          tooltip="Type"
        />
        <small-icon-item
          v-if="heuristicData.heuristicParameter"
          :title="heuristicData.heuristicParameter"
          :icon="mdiTune"
          :tooltip="heuristicData.heuristicParameterTitle?heuristicData.heuristicParameterTitle:'Parameter'"
        />
        <small-icon-item
          v-if="heuristicData.heuristicCustomClusters"
          :icon="mdiMerge"
          tooltip="Custom clusters"
        />
        <small-icon-item
          v-if="heuristicData.heuristicExcludeAddresses"
          :icon="mdiPlaylistRemove"
          tooltip="Exclude addresses"
        />
        <small-icon-item
          v-if="heuristicData.heuristicExcludeSpendingGaps"
          :icon="mdiClockAlertOutline"
          tooltip="Exclude spending gaps"
        />
        <small-icon-item
          :title="heuristicData.heuristicTimestamp.toLocaleDateString()"
          :icon="mdiCalendar"
          :tooltip="`Created ${heuristicData.heuristicTimestamp.toLocaleString()}`"
        />
      </div>
      <v-card
        v-show="heuristicData.clusters?.length > 0"
        variant="text"
        class="me-auto my-4"
      >
        <named-divider
          v-if="enoughDataForGraph"
          title="Transactions"
          title-class="text-subtitle-1"
          :vertical-margin="0"
        />
        <svg
          id="heuristic_details_canvas"
          class="mt-3"
          :class="{'hide':!enoughDataForGraph}"
        />
        <v-card-title
          v-if="!enoughDataForGraph"
          class="text-h5"
        >
          Not enough data to display diagram
        </v-card-title>
        <v-card-text v-if="!enoughDataForGraph && durationInMinutes > 0">
          {{ `Only ${durationInMinutes} ${plural('minute', durationInMinutes)} between earliest and latest origin.` }}
        </v-card-text>
        <v-card-text v-if="!enoughDataForGraph && durationInMinutes === 0">
          All origins occur in the same point of time.
        </v-card-text>
      </v-card>
      <template v-if="heuristicData.clusters?.length > 0">
        <named-divider
          title="Clusters"
          title-class="text-subtitle-1"
          :vertical-margin="0"
        />
        <results :result-items="heuristicData.clusters" />
      </template>
    </v-card-text>
  </v-card>
</template>

<script setup>
import Results from '@/components/workspace/sidebars/Results.vue';
import {
	mdiApplicationVariableOutline,
	mdiCalendar,
	mdiClockAlertOutline,
	mdiMerge,
	mdiPlaylistRemove,
	mdiTune,
} from '@mdi/js';
import BarChart from '@/d3Documents/barChart.js';
import NamedDivider from '@/components/common/NamedDivider.vue';
import {
	computed, onMounted, onUpdated, ref,
} from 'vue';
import {getColorMap, plural, setUndefinedTransactionColor} from '@/utilities/index.js';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';
import SmallIconItem from '@/components/common/SmallIconItem.vue';

const props = defineProps({
	heuristicData: {type: Object, required: true},
});

const {getSettings} = storeToRefs(useLocalStore());

const colorMap = getColorMap(getSettings.value.blockchainMode);
setUndefinedTransactionColor(colorMap, undefined);
let svgBarChart = null;
const enoughDataForGraph = ref(true);
const durationInMinutes = ref(0);

// Computed
const transactionCount = computed(() => {
	if (!props.heuristicData.clusters) {
		return 0;
	}

	let count = 0;

	props.heuristicData.clusters.forEach(cluster => {
		count += cluster.transactions.length;
	});

	return count;
});

// Hooks
onUpdated(() => {
	init();
});

onMounted(() => {
	init();
});

// Function
function init() {
	// Do nothing if sheet is not open
	if (!props.heuristicData) {
		return;
	}

	svgBarChart = new BarChart('heuristic_details_canvas', 600, 150);
	updateData(props.heuristicData.clusters);
}

function updateData(graphData) {
	// Flatten
	const detailArray = [];
	graphData.forEach(d => {
		detailArray.push(...d.transactions);
	});

	svgBarChart.drawStacked(detailArray, colorMap);
	enoughDataForGraph.value = !svgBarChart.empty;
	durationInMinutes.value = svgBarChart.getDurationInMinutes;
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

</style>
