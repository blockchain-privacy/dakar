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
            :icon="sigmaLeft"
          >
            {{ selectorData.inputSum?.min? convertAmount(selectorData.inputSum.min):0 }} - {{ selectorData.inputSum?.max?convertAmount(selectorData.inputSum.max):'*' }}
          </icon-item>
        </v-col>
        <v-col v-if="selectorData.outputSum?.min || selectorData.outputSum?.max">
          <icon-item
            title="Output Sum"
            :icon="sigmaRight"
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
            :icon="cashLeft"
          >
            {{ selectorData.inputRange?.min? convertAmount(selectorData.inputRange.min):0 }} - {{ selectorData.inputRange?.max?convertAmount(selectorData.inputRange.max):'*' }}
          </icon-item>
        </v-col>
        <v-col v-if="selectorData.outputRange?.min || selectorData.outputRange?.max">
          <icon-item
            title="Output Range"
            :icon="cashRight"
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
            title="Exclude Transaction Types"
            :icon="mdiIncognitoOff"
          >
            {{ selectorData.excludePrivacyTransactions }}
          </icon-item>
        </v-col>
        <v-col v-else-if="selectorData.txTypes">
          <icon-item
            title=" Privacy Type Filter"
            :icon="incognitoFilter"
          >
            <div class="d-flex flex-wrap">
              <color-chip
                v-for="p in selectorData.txTypes"
                :key="p"
                class="me-2 mb-4"
                :title="p"
                :color="colorMap.get(p)"
              />
            </div>
          </icon-item>
        </v-col>
      </v-row>
      <v-row>
        <v-col
          v-if="selectorData.depth"
          cols="12"
          xs="12"
          sm="6"
        >
          <icon-item
            title="Traversal Depth"
            :icon="mdiArrowCollapseDown"
          >
            {{ selectorData.depth }}
          </icon-item>
        </v-col>
        <v-col v-if="selectorData.isForward !== undefined">
          <icon-item
            title="Traversal Direction"
            :icon="selectorData.isForward?mdiArrowRight:mdiArrowLeft"
          >
            {{ selectorData.isForward?'forward':'backward' }}
          </icon-item>
        </v-col>
      </v-row>
      <v-row v-if="selectorData.transactions?.length > 0">
        <v-col
          cols="12"
          xs="12"
          sm="6"
        >
          <icon-item
            title="Total Transaction Count"
            :icon="mdiPoundBoxOutline"
          >
            {{ selectorData.selectorTotalResultCount.toLocaleString() }}
          </icon-item>
        </v-col>
        <v-col>
          <icon-item
            title="Stored Transaction Count"
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
        <v-data-table
          :items="tableItems"
          :headers="tableHeaders"
        >
          <template #item.txhash="{item}">
            <td>
              <workspace-link
                :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
                       params: { id: item.txhash, blockchainMode: getSettings.blockchainMode }}"
              >
                {{ item.txhash }}
              </workspace-link>
            </td>
          </template>
          <template #item.ts="{item}">
            <td>{{ new Date(item.ts).toLocaleString() }}</td>
          </template>
        </v-data-table>
      </template>
    </v-card-text>
  </v-card>
</template>

<script setup>
import IconItem from '@/components/common/IconItem.vue';
import {
	mdiArrowCollapseDown,
	mdiArrowLeft, mdiArrowRight,
	mdiCalendar, mdiCalendarEnd, mdiCalendarStart, mdiIncognitoOff, mdiPoundBoxOutline,
} from '@mdi/js';
import BarChart from '@/d3Documents/barChart.js';
import NamedDivider from '@/components/common/NamedDivider.vue';
import {
	computed, onMounted, onUpdated, ref,
} from 'vue';
import {
	convertAmount, getColorMap, plural,
} from '@/utilities/index.js';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants/index.js';
import {
	cashLeft, cashRight, sigmaLeft, sigmaRight, incognitoFilter,
} from '@/customIcons/index.js';
import ColorChip from '@/components/common/ColorChip.vue';
import {useLocalStore} from '@/pinia/local.js';
import {storeToRefs} from 'pinia';

const props = defineProps({selectorData: {type: Object, required: true}});

const {getSettings} = storeToRefs(useLocalStore());

const colorMap = getColorMap();
let svgBarChart = null;
const tableHeaders = [
	{
		key: 'txhash', title: 'Tranasaction Hash', sortable: false, align: 'left',
	},
	{
		key: 'ts', title: 'Timestamp', align: 'right',
	},
];

const enoughDataForGraph = ref(true);
const durationInMinutes = ref(0);

// Computed
const transactionCount = computed(() => {
	if (!props.selectorData.transactions) {
		return 0;
	}

	return props.selectorData.transactions.length;
});

const tableItems = computed(() => {
	if (!props.selectorData.transactions) {
		return [];
	}

	return props.selectorData.transactions.map(d => {
		d.ts = new Date(d.ts).getTime();

		return d;
	});
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

	svgBarChart = new BarChart('selector_details_canvas', 600, 300, false);
	svgBarChart.draw(props.selectorData.transactions);
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
