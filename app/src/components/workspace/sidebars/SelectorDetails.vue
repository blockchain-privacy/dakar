<template>
  <v-card variant="text">
    <v-card-text>
      <v-row>
        <v-col>
          <icon-item
            title="Timestamp"
            :icon="mdiCalendar"
          >
            {{ selectorData.selectorTimestamp.toLocaleString() }}
          </icon-item>
        </v-col>
      </v-row>
      <v-row>
        <v-col
          v-if="selectorData.startDate"
          cols="12"
          xs="12"
          sm="6"
        >
          <icon-item
            title="Start date"
            :icon="mdiCalendarStart"
          >
            {{ new Date(selectorData.startDate).toLocaleDateString() }}
          </icon-item>
        </v-col>
        <v-col v-if="selectorData.endDate">
          <icon-item
            title="End date"
            :icon="mdiCalendarEnd"
          >
            {{ new Date(selectorData.endDate).toLocaleDateString() }}
          </icon-item>
        </v-col>
      </v-row>
      <v-row>
        <v-col
          v-if="selectorData.inputSum?.min || selectorData.inputSum?.max"
          cols="12"
          xs="12"
          sm="6"
        >
          <icon-item
            title="Input Sum"
            :icon="mdiSigma"
          >
            {{ selectorData.inputSum?.min? convertAmount(selectorData.inputSum.min):0 }} - {{ selectorData.inputSum?.max?convertAmount(selectorData.inputSum.max):'*' }}
          </icon-item>
        </v-col>
        <v-col v-if="selectorData.outputSum?.min || selectorData.outputSum?.max">
          <icon-item
            title="Output Sum"
            :icon="mdiSigma"
          >
            {{ selectorData.outputSum?.min? convertAmount(selectorData.outputSum.min):0 }} - {{ selectorData.outputSum?.max?convertAmount(selectorData.outputSum.max):'*' }}
          </icon-item>
        </v-col>
      </v-row>
      <v-row>
        <v-col
          v-if="selectorData.inputRange?.min || selectorData.inputRange?.max"
          cols="12"
          xs="12"
          sm="6"
        >
          <icon-item
            title="Input Range"
            :icon="mdiCurrencyUsd"
          >
            {{ selectorData.inputRange?.min? convertAmount(selectorData.inputRange.min):0 }} - {{ selectorData.inputRange?.max?convertAmount(selectorData.inputRange.max):'*' }}
          </icon-item>
        </v-col>
        <v-col v-if="selectorData.outputRange?.min || selectorData.outputRange?.max">
          <icon-item
            title="Output Range"
            :icon="mdiCurrencyUsd"
          >
            {{ selectorData.outputRange?.min? convertAmount(selectorData.outputRange.min):0 }} - {{ selectorData.outputRange?.max?convertAmount(selectorData.outputRange.max):'*' }}
          </icon-item>
        </v-col>
      </v-row>
      <v-row>
        <v-col
          v-if="selectorData.excludePrivacyTransactions"
          cols="12"
          xs="12"
          sm="6"
        >
          <icon-item
            title="Exclude Priv. Transactions"
            :icon="mdiIncognito"
          >
            {{ selectorData.excludePrivacyTransactions }}
          </icon-item>
        </v-col>
        <v-col v-else-if="selectorData.privacyTypes">
          <p class="text-subtitle-1 mb-2 text-center">
            Privacy Type Filter
          </p>
          <div class="d-flex justify-center flex-wrap">
            <v-chip
              v-for="p in selectorData.privacyTypes"
              :key="p"
              rounded
              class="me-2 mb-4"
            >
              <template #prepend>
                <v-sheet
                  style="width:15px; height:15px"
                  rounded
                  :color="colorMap.get(p)"
                  class="me-2"
                />
              </template>
              {{ capitalize(p) }}
            </v-chip>
          </div>
        </v-col>
      </v-row>
      <v-row v-if="selectorData.transactions?.length > 0">
        <v-col
          cols="12"
          xs="12"
          sm="6"
        >
          <icon-item
            title="Total transaction count"
            :icon="mdiPoundBoxOutline"
          >
            {{ selectorData.selectorTotalResultCount.toLocaleString() }}
          </icon-item>
        </v-col>
        <v-col>
          <icon-item
            title="Stored transaction count"
            :icon="mdiPoundBoxOutline"
          >
            {{ transactionCount.toLocaleString() }}
          </icon-item>
        </v-col>
      </v-row>
      <v-row v-else>
        <v-col>
          <v-card-title class="text-h5">
            No results
          </v-card-title>
          <v-card-text>
            This selector returned no results. Try different parameters or a larger time range.
          </v-card-text>
        </v-col>
      </v-row>
      <v-card
        v-show="selectorData.transactions?.length > 0"
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
          id="selector_details_canvas"
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
      <template v-if="selectorData.transactions?.length > 0">
        <v-divider />
        <v-list>
          <v-list-item
            v-for="t in selectorData.transactions"
            :key="t.txhash"
          >
            <v-list-item-title>
              <workspace-link
                :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,params: { id: t.txhash }}"
                class="shorten"
              >
                {{ t.txhash }}
              </workspace-link>
            </v-list-item-title>
          </v-list-item>
        </v-list>
      </template>
    </v-card-text>
  </v-card>
</template>

<script setup>
import IconItem from '@/components/common/IconItem.vue';
import {
	mdiCalendar, mdiCalendarEnd, mdiCalendarStart, mdiCurrencyUsd, mdiIncognito, mdiPoundBoxOutline, mdiSigma,
} from '@mdi/js';
import Histogram from '@/d3Documents/histogram.js';
import NamedDivider from '@/components/common/NamedDivider.vue';
import {
	computed, onMounted, onUpdated, ref,
} from 'vue';
import {
	capitalize, convertAmount, getColorMap, plural,
} from '@/utilities/index.js';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants/index.js';

const props = defineProps({selectorData: {type: Object, required: true}});

const colorMap = getColorMap();

let svgHistogram = null;
const enoughDataForGraph = ref(true);
const durationInMinutes = ref(0);

// Computed
const transactionCount = computed(() => {
	if (!props.selectorData.transactions) {
		return 0;
	}

	return props.selectorData.transactions.length;
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
	if (!props.selectorData) {
		return;
	}

	svgHistogram = new Histogram('selector_details_canvas', 600, 300, false);
	svgHistogram.draw(props.selectorData.transactions);
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
