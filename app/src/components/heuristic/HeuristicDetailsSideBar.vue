<template>
  <side-bar
    v-model="inputVal"
    title="Heuristic Properties"
    :icon="icon.mdiChartBar"
  >
    <template #actions>
      <v-btn
        v-if="!isHollow && heuristicData.isLoaded"
        id="heuristic_download"
        icon
        variant="text"
        class="ms-auto"
        @click="downloadSummary"
      >
        <v-icon>{{ icon.mdiFileDownloadOutline }}</v-icon>
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
              <IconItem
                title="Type"
                :icon="icon.mdiApplicationVariableOutline"
              >
                {{ heuristicData.heuristicTypeTitle }}
              </IconItem>
            </v-col>
            <v-col v-if="heuristicData.heuristicParameter">
              <IconItem
                title="Parameter"
                :icon="icon.mdiTune"
              >
                {{ heuristicData.heuristicParameter }}
              </IconItem>
            </v-col>
          </v-row>
          <v-row>
            <v-col
              v-if="heuristicData.heuristicCustomClusters"
              cols="12"
              xs="12"
              sm="6"
            >
              <IconItem
                title="Custom clusters"
                :icon="icon.mdiMerge"
              >
                yes
              </IconItem>
            </v-col>
            <v-col v-if="heuristicData.heuristicExcludeAddresses">
              <IconItem
                title="Exclude addresses"
                :icon="icon.mdiPlaylistRemove"
              >
                yes
              </IconItem>
            </v-col>
          </v-row>
          <v-row v-if="heuristicData.heuristicExcludeSpendingGaps">
            <v-col>
              <IconItem
                title="Exclude spending gaps"
                :icon="icon.mdiClockAlertOutline"
              >
                yes
              </IconItem>
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
              <IconItem
                title="Number of origins"
                :icon="icon.mdiPoundBoxOutline"
              >
                {{ transactionCount }}
              </IconItem>
            </v-col>
            <v-col>
              <IconItem
                title="Number of clusters"
                :icon="icon.mdiPoundBoxOutline"
              >
                {{ heuristicData.clusterCount ? heuristicData.clusterCount : 0 }}
              </IconItem>
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

<script>
import Results from '@/components/heuristic/Results.vue';
import IconItem from '@/components/common/IconItem.vue';
import {
	mdiApplicationVariableOutline,
	mdiChartBar, mdiClockAlertOutline, mdiFileDownloadOutline,
	mdiMerge,
	mdiPlaylistRemove,
	mdiPoundBoxOutline,
	mdiTune,
} from '@mdi/js';
import Histogram from '@/d3Documents/histogram';
import {getCurrentDate} from '@/utilities';
import NamedDivider from '@/components/common/NamedDivider.vue';
import SideBar from '@/components/heuristic/SideBar.vue';

export default {
	name: 'HeuristicDetailsSidebar',
	components: {SideBar, NamedDivider, Results, IconItem},
	props: {
		modelValue: {type: Boolean, required: true},
		heuristicData: {type: Object, required: true},
		clusters: {type: Array, required: true},
		newHeuristicPrefix: {type: String, required: true},
	},
	emits: ['update:modelValue'],
	data() {
		return {
			icon: {
				mdiApplicationVariableOutline,
				mdiTune,
				mdiPoundBoxOutline,
				mdiChartBar,
				mdiMerge,
				mdiPlaylistRemove,
				mdiFileDownloadOutline,
				mdiClockAlertOutline,
			},
			chart: null,
			svgHistogram: null,
			enoughDataForGraph: true,
			durationInMinutes: 0,
		};
	},
	computed: {
		isHollow() {
			return this.heuristicData.heuristicUid.startsWith(this.newHeuristicPrefix);
		},
		transactionCount() {
			if (!this.heuristicData.clusters) {
				return 0;
			}

			let count = 0;

			this.heuristicData.clusters.forEach(cluster => {
				count += cluster.txs.length;
			});
			return count;
		},
		inputVal: {
			get() {
				return this.modelValue;
			},
			set(val) {
				this.$emit('update:modelValue', val);
			},
		},
	},
	updated() {
		this.doUpdate();
	},
	mounted() {
		this.svgHistogram = new Histogram('heuristic_details_canvas', 600, 300, false);
	},
	methods: {
		setErrorMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'error', temporary: true, category: this.$route.name});
		},
		doUpdate() {
			// Do nothing if sheet is not open
			if (!this.modelValue || !this.clusters) {
				return;
			}

			this.svgHistogram.reset();
			this.updateData(this.clusters);
		},
		updateData(graphData) {
			// Flatten
			const detailArray = [];
			graphData.forEach(d => {
				detailArray.push(...d.txs);
			});

			this.svgHistogram.draw(detailArray);
			this.enoughDataForGraph = !this.svgHistogram.empty;
			this.durationInMinutes = this.svgHistogram.getDurationInMinutes;
		},
		async downloadSummary() {
			try {
				const response = await this.dakar.heuristic.heuristicsSummaryHeuristicUIDGet({heuristicUID: this.heuristicData.heuristicUid});
				// Looks hacky, but it is the only way with good UX
				const a = document.createElement('a');
				a.href = URL.createObjectURL(response);

				a.setAttribute(
					'download',
					`heuristic_summary_${getCurrentDate()}_${this.heuristicData.heuristicUid}.csv`,
				);
				a.click();
				a.remove();
			} catch (e) {
				this.setErrorMessage(e);
			}
		},
	},
};
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
