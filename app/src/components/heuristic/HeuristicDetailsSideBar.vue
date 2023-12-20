<template>
  <side-bar
    v-model="inputVal"
    title="Heuristic Properties"
    :icon="mdiChartBar"
  >
    <template #actions>
      <v-btn
        v-if="!isHollow && heuristicData.isLoaded"
        id="heuristic_download"
        :icon="true"
        variant="text"
        class="ms-auto"
        @click="downloadSummary"
      >
        <v-icon>{{ mdiFileDownloadOutline }}</v-icon>
      </v-btn>
      <v-tooltip
        location="bottom"
        activator="#heuristic_download"
      >
        <span>Download heuristic summary</span>
      </v-tooltip>
    </template>
    <template #body>
      <v-card variant="text">
        <v-card-text v-show="heuristicData.isLoaded">
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
          <v-row v-if="isHollow">
            <v-col>
              <v-card-title class="text-h5">
                Not executed
              </v-card-title>
              <v-card-text>
                This heuristic has not been executed, therefore no results are available.
              </v-card-text>
            </v-col>
          </v-row>
          <v-row v-else-if="clusters.length > 0">
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
            v-show="clusters.length > 0"
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
              {{
                `Only ${durationInMinutes} minute${durationInMinutes > 1 ? 's' : ''}
                between earliest and latest origin.`
              }}
            </v-card-text>
            <v-card-text v-if="!enoughDataForGraph && durationInMinutes === 0">
              All origins occur in the same point of time.
            </v-card-text>
          </v-card>
          <named-divider
            v-if="clusters.length > 0"
            title="Clusters"
            title-class="text-subtitle-1"
            :vertical-margin="0"
          />
          <results
            v-if="clusters.length > 0"
            :result-items="clusters"
          />
        </v-card-text>
        <v-skeleton-loader
          v-if="!heuristicData.isLoaded"
          type="article,article,article,article,article"
        />
      </v-card>
    </template>
  </side-bar>
</template>

<script setup>
import Results from '@/components/heuristic/Results.vue';
import IconItem from '@/components/common/IconItem.vue';
import {
	mdiApplicationVariableOutline, mdiChartBar, mdiClockAlertOutline, mdiFileDownloadOutline,
	mdiMerge, mdiPlaylistRemove, mdiPoundBoxOutline, mdiTune, mdiCalendar,
} from '@mdi/js';
import Histogram from '@/d3Documents/histogram';
import {getCurrentDate} from '@/utilities';
import NamedDivider from '@/components/common/NamedDivider.vue';
import SideBar from '@/components/common/SideBar.vue';
import {computed, inject, onMounted, onUpdated, ref} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();

const props = defineProps({
	modelValue: {type: Boolean, required: true},
	heuristicData: {type: Object, required: true},
	clusters: {type: Array, required: true},
	newHeuristicPrefix: {type: String, required: true},
});

const emit = defineEmits(['update:modelValue']);

let svgHistogram = null;
const enoughDataForGraph = ref(true);
const durationInMinutes = ref(0);

// Computed
const isHollow = computed(() => props.heuristicData.heuristicUid.startsWith(props.newHeuristicPrefix));
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

const inputVal = computed({
	get() {
		return props.modelValue;
	},
	set(val) {
		emit('update:modelValue', val);
	},
});

// Hooks
onUpdated(() => {
	doUpdate();
});

onMounted(() => {
	svgHistogram = new Histogram('heuristic_details_canvas', 600, 300, false);
});

// Function
function setErrorMessage(msg) {
	msgStore.addMessage({text: msg, type: 'error', temporary: true, category: route.name});
}

function doUpdate() {
	// Do nothing if sheet is not open
	if (!props.modelValue || !props.clusters) {
		return;
	}

	svgHistogram.reset();
	updateData(props.clusters);
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

async function downloadSummary() {
	try {
		const response = await dakar.heuristic.heuristicsSummaryHeuristicUIDGet({heuristicUID: props.heuristicData.heuristicUid});
		// Looks hacky, but it is the only way with good UX
		const a = document.createElement('a');
		a.href = URL.createObjectURL(response);

		a.setAttribute(
			'download',
			`heuristic_summary_${getCurrentDate()}_${props.heuristicData.heuristicUid}.csv`,
		);
		a.click();
		a.remove();
	} catch (e) {
		setErrorMessage(e);
	}
}
</script>

<style scoped>
/* css for d3 graph  */
:deep(.bar) {
  fill: rgb(var(--v-theme-primary));
}

:deep(.hide){
  display: none;
  height: 0;
}
</style>
