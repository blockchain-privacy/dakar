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
        style="min-width: 300px"
        :append-inner-icon="mdiMagnify"
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
            <v-icon>{{ mdiArrowUp }}</v-icon>
          </v-btn>
          <v-btn :value="true">
            <v-icon>{{ mdiArrowDown }}</v-icon>
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
            <attribution-tag
              v-for="(attribution, y) in cluster.raw.attributions"
              :key="y"
              :attribution="attribution"
              class="ms-2"
            />
            <result-item
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
        <v-icon :icon="mdiChevronLeft" />
      </v-btn>
      <v-btn
        icon
        variant="text"
        :disabled="numberOfPages === 1"
        @click="nextPage"
      >
        <v-icon :icon="mdiChevronRight" />
      </v-btn>
    </div>
  </v-container>
</template>

<script setup>
import {
	mdiArrowDown, mdiArrowUp, mdiChevronLeft, mdiChevronRight, mdiMagnify,
} from '@mdi/js';
import AttributionTag from '../tools/attributions/AttributionTag.vue';
import ResultItem from '@/components/workspace/ResultItem.vue';
import {
	computed, onMounted, onUpdated, ref,
} from 'vue';

const props = defineProps({resultItems: {type: Array, required: true}});

const itemsPerPage = 15;
const keys = [
	{title: 'Number of Transactions', key: 'transactionCount'},
	{title: 'Number of Attributions', key: 'attributionCount'},
];

const results = ref([]);
const sortKey = ref('transactionCount');
const sortDesc = ref(true);
const search = ref('');
const page = ref(1);

const numberOfPages = computed(() => Math.ceil(props.resultItems.length / itemsPerPage));
const clusters = computed(() => {
	if (!results.value) {
		return [];
	}

	const sortedResults = results.value.toSorted((a, b) => {
		let valA;
		let valB;

		if (sortKey.value === 'transactionCount') {
			valA = a.transactionCount;
			valB = b.transactionCount;
		} else {
			valA = a.attributionCount;
			valB = b.attributionCount;
		}

		if (sortDesc.value) {
			return valB - valA;
		}

		return valA - valB;
	});

	if (!search.value) {
		return sortedResults;
	}

	const query = search.value.trim();

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
});

// Hooks
onMounted(() => {
	setAdditionalAttributes();
});

onUpdated(() => {
	setAdditionalAttributes();
});

// Functions
function setAdditionalAttributes() {
	let clusterCounter = 0;
	results.value = props.resultItems.map(cluster => {
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
}

function nextPage() {
	if (page.value + 1 <= numberOfPages.value) {
		page.value += 1;
	}
}

function formerPage() {
	if (page.value - 1 >= 1) {
		page.value -= 1;
	}
}

</script>

<style scoped>

</style>
