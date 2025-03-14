<template>
  <v-card variant="text">
    <v-card-text>
      <div
        class="d-flex align-center flex-wrap justify-center"
        style="gap: 16px"
      >
        <v-card
          color="primary"
          variant="flat"
          min-width="150px"
        >
          <v-card-text>
            <div class="text-h4">
              {{ selectorData.selectorTotalResultCount.toLocaleString() }}
            </div>
            <div class="text-subtitle-1">
              Total Transactions
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
              {{ transactionCount.toLocaleString() }}
            </div>
            <div class="text-subtitle-1">
              Stored Transactions
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
          v-if="selectorData.startDate"
          :title="new Date(selectorData.startDate).toLocaleDateString()"
          :icon="mdiCalendarStart"
          tooltip="Start date"
        />
        <small-icon-item
          v-if="selectorData.endDate"
          :title="new Date(selectorData.endDate).toLocaleDateString()"
          :icon="mdiCalendarEnd"
          tooltip="End date"
        />
        <small-icon-item
          v-if="selectorData.inputSum?.min || selectorData.inputSum?.max"
          :title="`${selectorData.inputSum?.min? convertAmount(selectorData.inputSum.min):0} - ${selectorData.inputSum?.max?convertAmount(selectorData.inputSum.max):'*'}`"
          :icon="sigmaLeft"
          tooltip="Input sum"
        />
        <small-icon-item
          v-if="selectorData.outputSum?.min || selectorData.outputSum?.max"
          :title="`${selectorData.outputSum?.min? convertAmount(selectorData.outputSum.min):0} - ${selectorData.outputSum?.max?convertAmount(selectorData.outputSum.max):'*'}`"
          :icon="sigmaRight"
          tooltip="Output sum"
        />
        <small-icon-item
          v-if="selectorData.inputRange?.min || selectorData.inputRange?.max"
          :title="`${selectorData.inputRange?.min? convertAmount(selectorData.inputRange.min):0} - ${selectorData.inputRange?.max?convertAmount(selectorData.inputRange.max):'*'}`"
          :icon="cashLeft"
          tooltip="Input range"
        />
        <small-icon-item
          v-if="selectorData.outputRange?.min || selectorData.outputRange?.max"
          :title="`${selectorData.outputRange?.min? convertAmount(selectorData.outputRange.min):0} - ${selectorData.outputRange?.max?convertAmount(selectorData.outputRange.max):'*'}`"
          :icon="cashRight"
          tooltip="Output range"
        />
        <small-icon-item
          v-if="selectorData.excludePrivacyTransactions"
          :icon="mdiIncognitoOff"
          tooltip="Exclude transaction types"
        />
        <small-icon-item
          v-if="selectorData.depth"
          :title="`${selectorData.depth}`"
          :icon="mdiArrowCollapseDown"
          tooltip="Traversal depth"
        />
        <small-icon-item
          v-if="selectorData.isForward !== undefined"
          :title="selectorData.isForward?'forward':'backward'"
          :icon="selectorData.isForward?mdiArrowRight:mdiArrowLeft"
          tooltip="Traversal Direction"
        />
        <small-icon-item
          :title="selectorData.selectorTimestamp.toLocaleDateString()"
          :icon="mdiCalendar"
          :tooltip="`Created ${selectorData.selectorTimestamp.toLocaleString()}`"
        />
      </div>
      <div
        v-if="selectorData.txTypes"
        class="d-flex align-center flex-wrap itemContainer justify-center mt-2"
      >
        <small-icon-item
          :icon="incognitoFilter"
          tooltip="Transaction type filter"
        >
          <div class="d-flex flex-wrap">
            <color-chip
              v-for="p in selectorData.txTypes"
              :key="p"
              class="me-2"
              :title="p"
              :color="colorMap.get(p)"
              size="small"
            />
          </div>
        </small-icon-item>
      </div>
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
                style="max-width: 300px"
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
import {
	mdiArrowCollapseDown,
	mdiArrowLeft, mdiArrowRight,
	mdiCalendar, mdiCalendarEnd, mdiCalendarStart, mdiIncognitoOff,
} from '@mdi/js';
import BarChart from '@/d3Documents/barChart.js';
import NamedDivider from '@/components/common/NamedDivider.vue';
import {
	computed, onMounted, onUpdated, ref,
} from 'vue';
import {
	convertAmount, getColorMap, plural, setUndefinedTransactionColor,
} from '@/utilities/index.js';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants/index.js';
import {
	cashLeft, cashRight, sigmaLeft, sigmaRight, incognitoFilter,
} from '@/customIcons/index.js';
import ColorChip from '@/components/common/ColorChip.vue';
import {useLocalStore} from '@/pinia/local.js';
import {storeToRefs} from 'pinia';
import SmallIconItem from '@/components/common/SmallIconItem.vue';

const props = defineProps({selectorData: {type: Object, required: true}});

const {getSettings} = storeToRefs(useLocalStore());

const colorMap = getColorMap(getSettings.value.blockchainMode);
setUndefinedTransactionColor(colorMap, undefined);
let svgBarChart = null;
const tableHeaders = [
	{
		key: 'txhash', title: 'Transaction', sortable: false, align: 'left',
	},
	{
		key: 'txtype', title: 'Type', align: 'right',
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

	svgBarChart = new BarChart('selector_details_canvas', 600, 150);
	svgBarChart.drawStacked(props.selectorData.transactions, colorMap);
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
