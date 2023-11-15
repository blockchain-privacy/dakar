<template>
  <v-container
    v-if="resultItems && resultItems.length > 0"
    :fluid="true"
  >
    <div
      v-if="resultItems.length > 1"
      class="d-flex align-center flex-wrap"
    >
      <v-text-field
        v-model="search"
        class="me-3 mb-3"
        :clearable="true"
        :flat="true"
        hide-details
        style="min-width: 200px"
        :append-inner-icon="icons.mdiMagnify"
        label="Filter by tag and transaction hash"
      />
      <div class="d-flex align-center flex-wrap">
        <v-select
          v-model="sortKey"
          class="me-3 mb-3"
          :flat="true"
          hide-details
          :items="keys"
          label="Sort by"
          item-title="title"
          item-value="key"
        />
        <v-btn-toggle
          v-model="sortDesc"
          class="mb-3"
          mandatory
          variant="outlined"
          density="default"
        >
          <v-btn :value="false">
            <v-icon>{{ icons.mdiArrowUp }}</v-icon>
          </v-btn>
          <v-btn :value="true">
            <v-icon>{{ icons.mdiArrowDown }}</v-icon>
          </v-btn>
        </v-btn-toggle>
      </div>
    </div>
    <v-data-iterator
      v-model:items-per-page="itemsPerPage"
      v-model:page="page"
      :items="clusters"
      item-key="id"
      hide-default-footer
    >
      <template #default="{ items }">
        <div
          v-for="(cluster,i) in items"
          :key="cluster.raw.id"
        >
          <v-card
            class="my-2"
            variant="flat"
          >
            <AttributionTag
              v-for="(attribution, y) in cluster.raw.attributions"
              :key="y"
              :attribution="attribution"
              class="ms-2"
            />
            <ResultItem
              :max-items="5"
              :items="cluster.raw.txs"
            />
          </v-card>
          <v-divider
            v-if="i + 1 < items.length"
            thickness="2"
          />
        </div>
      </template>
    </v-data-iterator>
    <div class="d-flex align-center mt-2">
      <span class="ms-auto text-grey">
        Page {{ page }} of {{ numberOfPages }}
      </span>
      <v-btn
        icon
        variant="text"
        :disabled="numberOfPages === 1"
        @click="formerPage"
      >
        <v-icon :icon="icons.mdiChevronLeft" />
      </v-btn>
      <v-btn
        icon
        variant="text"
        :disabled="numberOfPages === 1"
        @click="nextPage"
      >
        <v-icon :icon="icons.mdiChevronRight" />
      </v-btn>
    </div>
  </v-container>
</template>

<script>
import {
	mdiChevronLeft, mdiChevronRight, mdiMagnify,
	mdiArrowUp, mdiArrowDown,	mdiChevronDown,
} from '@mdi/js';
import AttributionTag from '../tools/attributions/AttributionTag.vue';
import ResultItem from '@/components/heuristic/ResultItem.vue';

export default {
	name: 'Results',
	components: {ResultItem, AttributionTag},
	props: {
		resultItems: {type: Array, required: true},
	},
	data() {
		return {
			results: [],
			sortKey: 'transactionCount',
			sortDesc: true,
			itemsPerPage: 15,
			itemsPerPageArray: [4, 8, 12],
			search: '',
			page: 1,
			keys: [
				{title: 'Number of Transactions', key: 'transactionCount'},
				{title: 'Number of Attributions', key: 'attributionCount'},
			],
			icons: {
				mdiChevronLeft,
				mdiChevronRight,
				mdiMagnify,
				mdiArrowUp,
				mdiArrowDown,
				mdiChevronDown,
			},
		};
	},
	computed: {
		sortBy() {
			return [{key: this.sortKey, order: this.sortDesc}];
		},
		numberOfPages() {
			return Math.ceil(this.resultItems.length / this.itemsPerPage);
		},
		clusters() {
			if (!this.results) {
				return [];
			}

			const sortedResults = this.results.toSorted((a, b) => {
				let valA;
				let valB;

				if (this.sortKey === 'transactionCount') {
					valA = a.transactionCount;
					valB = b.transactionCount;
				} else {
					valA = a.attributionCount;
					valB = b.attributionCount;
				}

				if (this.sortDesc) {
					return valB - valA;
				}

				return valA - valB;
			});

			if (!this.search) {
				return sortedResults;
			}

			const query = this.search.trim();

			// Filter items based on search query and set counts
			return sortedResults.filter(cluster => {
				// Check if any transaction hash contains the search query
				for (const tx of cluster.txs) {
					if (tx.txhash.includes(query)) {
						return true;
					}
				}

				// Check if any attribution contains the search query
				if (cluster.attributions) {
					for (const attribution of cluster.attributions) {
						if (attribution.tag.includes(query)) {
							return true;
						}
					}
				}

				return false;
			});
		},
	},
	mounted() {
		this.setAdditionalAttributes();
	},
	updated() {
		this.setAdditionalAttributes();
	},
	methods: {
		setAdditionalAttributes() {
			let clusterCounter = 0;
			this.results = this.resultItems.map(cluster => {
				cluster.id = clusterCounter;
				clusterCounter += 1;

				// Set transaction count
				cluster.transactionCount = 0;
				if (cluster.txs) {
					cluster.transactionCount = cluster.txs.length;
				}

				// Set attribution count
				cluster.attributionCount = 0;
				if (cluster.attributions) {
					cluster.attributionCount = cluster.attributions.length;
				}

				return cluster;
			});
		},
		nextPage() {
			if (this.page + 1 <= this.numberOfPages) {
				this.page += 1;
			}
		},
		formerPage() {
			if (this.page - 1 >= 1) {
				this.page -= 1;
			}
		},
	},
};
</script>

<style scoped>

</style>
