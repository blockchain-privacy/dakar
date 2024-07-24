<template>
  <v-card variant="text">
    <v-card-text>
      <v-row>
        <v-col
          cols="12"
          xs="12"
          sm="6"
        >
          <icon-item
            title="Type"
            :icon="mdiApplicationVariableOutline"
          >
            {{ heuristicData.heuristicTypeTitle }}
          </icon-item>
        </v-col>
        <v-col>
          <icon-item
            title="Timestamp"
            :icon="mdiCalendar"
          >
            {{ heuristicData.heuristicTimestamp.toLocaleString() }}
          </icon-item>
        </v-col>
      </v-row>
      <v-row>
        <v-col
          v-if="heuristicData.heuristicParameter"
          cols="12"
          xs="12"
          sm="6"
        >
          <icon-item
            title="Parameter"
            :icon="mdiTune"
          >
            {{ heuristicData.heuristicParameter }}
          </icon-item>
        </v-col>
        <v-col v-if="heuristicData.heuristicCustomClusters">
          <icon-item
            title="Custom clusters"
            :icon="mdiMerge"
          >
            yes
          </icon-item>
        </v-col>
      </v-row>
      <v-row>
        <v-col
          v-if="heuristicData.heuristicExcludeAddresses"
          cols="12"
          xs="12"
          sm="6"
        >
          <icon-item
            title="Exclude addresses"
            :icon="mdiPlaylistRemove"
          >
            yes
          </icon-item>
        </v-col>
        <v-col v-if="heuristicData.heuristicExcludeSpendingGaps">
          <icon-item
            title="Exclude spending gaps"
            :icon="mdiClockAlertOutline"
          >
            yes
          </icon-item>
        </v-col>
      </v-row>
      <v-row v-if="heuristicData.clusters?.length > 0">
        <v-col
          cols="12"
          xs="12"
          sm="6"
        >
          <icon-item
            title="Number of origins"
            :icon="mdiPoundBoxOutline"
          >
            {{ transactionCount }}
          </icon-item>
        </v-col>
        <v-col>
          <icon-item
            title="Number of clusters"
            :icon="mdiPoundBoxOutline"
          >
            {{ heuristicData.clusterCount ? heuristicData.clusterCount : 0 }}
          </icon-item>
        </v-col>
      </v-row>
      <v-row v-else>
        <v-col>
          <v-card-title class="text-h5">
            No results
          </v-card-title>
          <v-card-text>
            This heuristic returned no results. Try different parameters,
            other heuristics or a different combination of heuristics.
          </v-card-text>
        </v-col>
      </v-row>
      <v-card
        v-show="heuristicData.clusters?.length > 0"
        variant="text"
        class="me-auto my-4"
      >
        <named-divider
          v-if="enoughDataForGraph"
          title="Origin Transactions"
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
import Results from '@/components/workspace/Results.vue';
import IconItem from '@/components/common/IconItem.vue';
import {
	mdiApplicationVariableOutline,
	mdiCalendar,
	mdiClockAlertOutline,
	mdiMerge,
	mdiPlaylistRemove,
	mdiPoundBoxOutline,
	mdiTune,
} from '@mdi/js';
import Histogram from '@/d3Documents/histogram';
import NamedDivider from '@/components/common/NamedDivider.vue';
import {
	computed, onMounted, onUpdated, ref,
} from 'vue';
import {plural} from '@/utilities';

const props = defineProps({
	heuristicData: {type: Object, required: true},
});

let svgHistogram = null;
const enoughDataForGraph = ref(true);
const durationInMinutes = ref(0);

// Computed
const transactionCount = computed(() => {
	if (!props.heuristicData.clusters) {
		return 0;
	}

	let count = 0;

	props.heuristicData.clusters.forEach(cluster => {
		count += cluster.txs.length;
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

	svgHistogram = new Histogram('heuristic_details_canvas', 600, 300, false);
	updateData(props.heuristicData.clusters);
}

function updateData(graphData) {
	// Flatten
	const detailArray = [];
	graphData.forEach(d => {
		detailArray.push(...d.txs);
	});

	svgHistogram.draw(detailArray);
	enoughDataForGraph.value = !svgHistogram.empty;
	durationInMinutes.value = svgHistogram.getDurationInMinutes;
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
