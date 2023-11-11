<template>
  <div>
    <v-card variant="text">
      <v-card-text>
        <WikiTooltip description-url="transactionTypes/privacyTransactions.md">
          Privacy transactions
        </WikiTooltip>
        which are directly connected to this address show partially the
        <WikiTooltip description-url="mixingActivity.md">
          mixing activity
        </WikiTooltip>.
        <v-row class="mt-2">
          <v-col class="d-flex align-center justify-center flex-wrap">
            <p class="v-label me-2">
              Filter by Privacy Type
            </p>
            <!-- selected-class="" is intentionally left blank to avoid a shadow over the chip elements -->
            <v-chip-group
              v-model="selectedPrivacyID"
              :column="true"
              :multiple="true"
              :filter="true"
              :mandatory="true"
              :disabled="!activities || activities.length === 0"
              selected-class=""
              color="primary"
              @update:model-value="updateSvgData(false)"
            >
              <v-chip
                v-for="label in privacyLabels"
                :key="label.id"
              >
                <template #prepend>
                  <v-sheet
                    style="width:25px; height:15px"
                    rounded
                    :color="label.color?label.color:'black'"
                    class="me-2"
                  />
                </template>
                {{ label.text }}
              </v-chip>
            </v-chip-group>
          </v-col>
        </v-row>
        <v-row class="mt-2">
          <v-col
            class="d-flex align-center"
            cols="12"
            lg="3"
          >
            <v-switch
              v-model="includeCusterAddresses"
              label="Include cluster addresses"
              hide-details
              :disabled="isLoading"
              @update:model-value="updateSvgData(true)"
            />
          </v-col>
          <v-col
            class="d-flex align-center"
            cols="12"
            lg="9"
          >
            <v-range-slider
              v-model="rangePicker.model"
              :disabled="!activities || activities.length === 0"
              :ticks="rangePicker.events"
              class="mr-8"
              :min="rangePicker.min"
              :max="rangePicker.max"
              label="Filter by time"
              hide-details
              thumb-label="always"
              show-ticks="always"
              track-size="1"
              tick-size="9"
              @end="updateSvgData(false)"
            >
              <template
                v-if="isSameDay"
                #thumb-label="{ modelValue }"
              >
                {{ new Date(modelValue).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}) }}
              </template>
              <template
                v-else
                #thumb-label="{ modelValue }"
              >
                {{ new Date(modelValue).toLocaleDateString() }}
              </template>
            </v-range-slider>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
    <template v-if="hasLoaded && !isLoading">
      <v-card
        v-if="showEmptyResponseMsg"
        class="my-3"
        variant="text"
      >
        <v-card-text
          class="text-h6"
          style="text-align:center"
        >
          No mixing activity detected
        </v-card-text>
      </v-card>
      <v-card
        v-if="showTooManyAddressesMsg"
        class="my-3"
        variant="text"
      >
        <v-card-text class="text-h6 text-center">
          Mixing activity lookup is not possible because the cluster
          of this address is connected to too many addresses
        </v-card-text>
      </v-card>
    </template>
    <v-card
      v-show="activities && activities.length > 0"
      class="my-3"
      variant="text"
    >
      <v-card-text>
        <v-tabs
          v-model="graphTabs"
          :grow="true"
          @update:model-value="onTabChange"
        >
          <v-tab key="histogram">
            Histogram
          </v-tab>
          <v-tab key="graph">
            Force Graph
          </v-tab>
        </v-tabs>
        <v-window
          v-model="graphTabs"
          :touch="false"
        >
          <v-window-item
            key="histogram"
            :eager="true"
          >
            <v-card variant="text">
              <v-card-text>
                <p
                  v-if="showNotEnoughDataMsg && !isLoading"
                  class="text-h6"
                  style="text-align: center"
                >
                  Not enough data available to draw chart
                </p>
                <div
                  v-if="showHistogram"
                  class="text-subtitle-1"
                  style="text-align: center"
                >
                  {{
                    selectedPrivacyLabel.length === 5 ? 'All Privacy'
                    : selectedPrivacyLabel.map(capitalize).join(', ')
                  }}
                  Transactions
                </div>
                <transaction-table-dialog
                  v-model="barTable.show"
                  :headers="barTable.headers"
                  :transactions="barTable.transactions"
                  :start-date="barTable.startDate"
                  :end-date="barTable.endDate"
                />
              </v-card-text>
              <div style="overflow:scroll">
                <svg
                  v-show="showHistogram"
                  id="mixing_activity_histogram"
                  style="min-width: 1100px"
                />
              </div>
            </v-card>
          </v-window-item>
          <v-window-item
            key="graph"
            :eager="true"
          >
            <v-card variant="text">
              <v-card-text>
                <div
                  v-if="!overrideTooManyTransactionsWarning && showTooManyTransactionsMsg"
                  style="text-align:center"
                >
                  <v-alert
                    variant="text"
                    :prominent="true"
                    type="warning"
                  >
                    The mixing activity results have more than
                    {{ tooManyTransactionsThreshold }} transactions.
                    Displaying a large number of items in a force graph may severely degrade
                    the performance of your browser.
                    Consider filtering the results by time or privacy type.
                  </v-alert>
                  <v-btn
                    color="primary"
                    @click="showForceGraphDespiteWarning"
                  >
                    Display force graph anyway
                  </v-btn>
                </div>

                <p
                  v-if="!showGraph && !isLoading && !showTooManyTransactionsMsg"
                  class="text-h6"
                  style="text-align: center"
                >
                  No data available
                </p>
                <transaction-dialog
                  v-if="clickedNode"
                  v-model="showNodeDialog"
                  :input-txs="clickedNode.input_txs?clickedNode.input_txs:[]"
                  :date-time="clickedNode.dateTime"
                  :privacy-type="clickedNode.privacytype"
                  :tx-hash="clickedNode.txhash"
                />
              </v-card-text>
              <svg
                v-show="showGraph"
                id="mixing_activity_force_graph"
                style="width:100%; height:500px"
              />
            </v-card>
          </v-window-item>
        </v-window>
      </v-card-text>
    </v-card>
    <v-progress-linear
      v-if="isLoading"
      class="mt-10"
      :indeterminate="true"
    />
  </div>
</template>

<script>
import {mdiCalendarRange} from '@mdi/js';
import Histogram from '@/d3Documents/histogram';
import ForceGraph from '@/d3Documents/forceGraph';
import {getPrivacyTypeLabel} from '@/utilities';
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';
import TransactionTableDialog from '@/components/explorer/address/TransactionTableDialog.vue';
import TransactionDialog from '@/components/explorer/address/TransactionDialog.vue';

// Capitalize returns the first letter of each word (separated by a space) in str capitalized
function capitalize(str) {
	return str.split(' ').map(d => d[0].toUpperCase() + d.slice(1)).join(' ');
}

export default {
	name: 'MixingActivity',
	components: {TransactionDialog, TransactionTableDialog, WikiTooltip},
	props: {
		addressHash: {type: String, required: true},
	},
	data() {
		return {
			icons: {mdiCalendarRange},
			txRoute: ROUTE_NAME_TRANSACTION_PAGE,
			lastQuery: '',
			includeCluster: false,
			// Select all labels by default
			selectedPrivacyID: [0, 1, 2, 3, 4],
			allPrivacyLabels: ['destination', 'collateral creation', 'collateral payment', 'origin', 'mixing'],
			showHistogram: false,
			showGraph: false,
			colorMap: new Map(),
			svgHistogram: null,
			svgGraph: null,
			isLoading: false,
			showEmptyResponseMsg: false,
			showTooManyAddressesMsg: false,
			showNotEnoughDataMsg: false,
			showTooManyTransactionsMsg: false,
			overrideTooManyTransactionsWarning: false,
			tooManyTransactionsThreshold: 500,
			activities: null,
			initialLoadDone: false,
			includeCusterAddresses: false,
			rangePicker: {
				model: null,
				min: 0,
				max: 0,
				events: [],
			},
			graphTabs: null,
			barTable: {
				headers: [{
					title: 'Transaction', align: 'start', key: 'txhash',
				},
				{title: 'Timestamp', key: 'dateTime'},
				{title: 'Transaction Type', key: 'privacytype'}],
				transactions: [],
				startDate: '',
				endDate: '',
				show: false,
			},
			clickedNode: {
				input_txs: [],
				dateTime: null,
				privacytype: '',
				txhash: '',
			},
			showNodeDialog: false,
			graphMode: false,
			hasLoaded: false,
		};
	},
	computed: {
		privacyLabels() {
			const labels = [];

			this.colorMap.forEach((v, k) => {
				labels.push({text: capitalize(k), color: v, id: k});
			});

			return labels;
		},
		// Returns truer if min and max or on the same calendar day
		isSameDay() {
			const day1 = new Date(this.rangePicker.min);
			const day2 = new Date(this.rangePicker.max);
			// Cut off time
			day1.setHours(0, 0, 0, 0);
			day2.setHours(0, 0, 0, 0);
			// Compare numbers
			return day1.getTime() === day2.getTime();
		},
		selectedPrivacyLabel() {
			const selectedLabels = [];

			this.selectedPrivacyID.forEach(i => {
				selectedLabels.push(this.allPrivacyLabels[i]);
			});

			return selectedLabels;
		},
	},
	watch: {
		addressHash() {
			// Prop was changed -> pull new data
			this.updateSvgData(true);
		},
	},
	beforeMount() {
		this.includeCusterAddresses = this.includeCluster;
		this.colorMap.set('destination', '#0072B2');
		this.colorMap.set('collateral creation', '#E69F00');
		this.colorMap.set('collateral payment', '#009E73');
		this.colorMap.set('origin', '#D55E00');
		this.colorMap.set('mixing', '#56B4E9');

		this.svgHistogram = new Histogram(
			'mixing_activity_histogram',
			1200,
			300,
			false,
		);
		this.svgHistogram.setClickHandler(this.onBarClick);
	},
	mounted() {
		// Has to be called after the SVG is included in the DOM
		this.svgGraph = new ForceGraph(1200, 500, 'mixing_activity_force_graph', this.colorMap);
		this.svgGraph.setClickHandler(this.onNodeClick);
	},
	created() {
		this.updateSvgData(true);
	},
	methods: {
		capitalize,
		setErrorMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'error', temporary: true, category: this.$route.name});
		},
		onBarClick(data) {
			if (data.x0.getHours() === data.x1.getHours()
          && data.x0.getMinutes() === data.x1.getMinutes()) {
				this.barTable.startDate = data.x0.toLocaleDateString();
				this.barTable.endDate = data.x1.toLocaleDateString();
			} else {
				this.barTable.startDate = data.x0.toLocaleString();
				this.barTable.endDate = data.x1.toLocaleString();
			}

			this.barTable.transactions = data;
			this.barTable.show = true;
		},
		onNodeClick(data) {
			this.clickedNode = data;
			this.showNodeDialog = true;
		},
		async getMixingActivity() {
			this.lastQuery = this.addressHash;
			const response = {ok: false, data: null, msg: null};
			try {
				response.data = await this.dakar.tools.mixingActivityPost({
					activity: {
						addressHash: this.addressHash,
						isClusterLookup: this.includeCusterAddresses,
					},
				});
				response.ok = true;
			} catch (e) {
				if (e.message === 'too_many_addresses') {
					response.msg = e.message;
				}
			}

			return response;
		},
		getFilteredData(withLinks) {
			let fromDate = null;
			let toDate = null;
			let considerDate = true;

			if (this.rangePicker.model === null) {
				considerDate = false;
			} else {
				fromDate = new Date(this.rangePicker.model[0]);
				toDate = new Date(this.rangePicker.model[1]);
			}

			const ret = {items: [], links: []};
			const events = new Set();
			const numActivities = this.activities.length;
			ret.items = this.activities.filter(d => {
				if (this.selectedPrivacyLabel.length < 5
            && !this.selectedPrivacyLabel.includes(d.privacytype)) {
					return false;
				}

				if (!considerDate) {
					return true;
				}

				const eventTime = d.dateTime;

				// Decrease accuracy of range picker ticks when many activities exist
				if (numActivities > 500) {
					eventTime.setHours(0, 0, 0, 0);
				} else if (numActivities > 200) {
					eventTime.setMinutes(0, 0, 0);
				}

				events.add(eventTime.getTime());

				return d.dateTime <= toDate && d.dateTime >= fromDate;
			});

			// Construct event objet
			const eventObj = {};
			Array.from(events).forEach(val => {
				eventObj[val] = '';
			});
			this.rangePicker.events = eventObj;

			if (withLinks) {
				const filteredHashes = new Set(ret.items.map(d => d.txhash));

				ret.items.forEach(d => {
					if (!d.input_txs) {
						return;
					}

					d.input_txs.forEach(it => {
						if (!filteredHashes.has(it.txhash)) {
							return;
						}

						ret.links.push({source: it.txhash, target: d.txhash});
					});
				});
			}

			return ret;
		},
		getCategories(filtered) {
			if (this.selectedPrivacyLabel.length < 5) {
				return this.selectedPrivacyLabel;
			}

			return [...new Set(filtered.map(d => d.privacytype))];
		},
		showForceGraphDespiteWarning() {
			this.overrideTooManyTransactionsWarning = true;
			this.onTabChange(1);
		},
		onTabChange(tab) {
			// Tab === 0: histogram
			// tab === 1: force graph
			const wantGraph = tab === 1;
			// Check if tab was actually changed. @changed:modelValue also fires on initial load of component
			if (this.graphMode === wantGraph) {
				return;
			}

			this.graphMode = wantGraph;

			if (this.graphMode) {
				if (!this.showTooManyTransactionsMsg
            || this.overrideTooManyTransactionsWarning) {
					this.updateSvgData();
				}

				return;
			}

			this.updateSvgData();
		},
		async updateSvgData(pullNewData) {
			this.showHistogram = false;
			this.showGraph = false;
			this.isLoading = true;
			this.showTooManyAddressesMsg = false;
			this.clickedNode = null;
			this.barTable.transactions = [];
			// Check if new data has to be loaded
			if (pullNewData || !this.initialLoadDone) {
				const mixingActivity = await this.getMixingActivity();
				this.hasLoaded = true;

				if (!mixingActivity.ok) {
					this.isLoading = false;
					this.activities = [];

					if (mixingActivity.msg === 'too_many_addresses') {
						this.showTooManyAddressesMsg = true;
						this.initialLoadDone = true;
						return;
					}

					this.setErrorMessage('error getting mixing activity');
					return;
				}

				if (!mixingActivity.data.activities) {
					this.showEmptyResponseMsg = true;
					this.isLoading = false;
					this.activities = [];
					return;
				}

				// Used to set boundaries for the date picker
				let maxDate = null;
				let minDate = null;

				this.activities = mixingActivity.data.activities.map(d => {
					d.privacytype = getPrivacyTypeLabel(d.privacytype);
					d.dateTime = new Date(d.block[0].ts);

					if (maxDate === null || d.dateTime > maxDate) {
						maxDate = new Date(d.dateTime);
					}

					if (minDate === null || d.dateTime < minDate) {
						minDate = new Date(d.dateTime);
					}

					return d;
				});

				this.rangePicker.min = new Date(minDate).getTime();
				this.rangePicker.max = new Date(maxDate).getTime();
				this.rangePicker.model = [this.rangePicker.min, this.rangePicker.max];
				this.initialLoadDone = true;
			}

			this.showEmptyResponseMsg = false;

			const filtered = this.getFilteredData(this.graphMode);

			if (!filtered.items) {
				this.isLoading = false;
				this.showGraph = false;
				this.showHistogram = false;
				this.showNotEnoughDataMsg = true;
				return;
			}

			this.showTooManyTransactionsMsg = filtered.items.length > this.tooManyTransactionsThreshold;

			// Draw
			if (this.graphMode) {
				if (!this.showTooManyTransactionsMsg
            || this.overrideTooManyTransactionsWarning) {
					this.svgGraph.draw(filtered.items, filtered.links);
					this.showGraph = true;
				}
			} else {
				this.svgHistogram.reset();
				this.svgHistogram.drawStacked(
					filtered.items,
					this.getCategories(filtered.items),
					this.colorMap,
				);
				this.showHistogram = !this.svgHistogram.empty;
				this.showNotEnoughDataMsg = this.svgHistogram.empty;
			}

			this.isLoading = false;
		},
	},
};
</script>

<style scoped>
:deep( .overlay ) {
  stroke-width: 2px;
  stroke: #1976D2;
  fill: #1976D2;
  cursor: pointer;
}

:deep( .nodeMouseOver ) {
  cursor: pointer;
}

:deep( .v-slider-thumb__label ) {
  min-width: 75px;
}

:deep(.v-slider-track__tick--filled) {
  background-color: rgb(var(--v-theme-primary-lighten-2));
}

:deep(.v-slider-track__tick) {
  background-color: rgb(var(--v-theme-primary-lighten-2));
}
</style>
