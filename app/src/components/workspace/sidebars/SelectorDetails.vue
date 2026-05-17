<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-card variant="text">
    <v-card-text>
      <template v-if="selectorData.selectorStatus === SELECTOR_STATUS_ERROR">
        <div class="d-flex flex-column align-center">
          <v-icon
            class="text-grey"
            :icon="mdiAlertCircle"
            size="90"
          />
          <div style="max-width: 400px">
            <template v-if="selectorData.selectorErrorCode === SELECTOR_ERROR_CODE_RESULT_LIMIT_EXCEEDED">
              <p class="text-title-large text-center mt-2">
                This heuristic exceeded the limit of {{ SELECTOR_RESULT_LIMIT.toLocaleString() }} results.
              </p>
              <p class="text-h7 text-center mt-2">
                Try running this heuristic again with a stronger filter.
              </p>
            </template>
            <template v-else>
              <p class="text-title-large text-center mt-2">
                An error occurred while running this CoinJoin heuristic.
              </p>
              <p class="text-h7 text-center mt-2">
                Try running this heuristic again. Consider using different parameters. If the issue is persists, please report the error with any details.
              </p>
            </template>
          </div>
        </div>
      </template>
      <div
        v-else
        class="d-flex align-center flex-wrap justify-center"
        style="gap: 16px"
      >
        <v-card
          v-if="selectorType === SELECTOR_TYPE_HEURISTIC"
          color="primary"
          variant="flat"
          min-width="150px"
        >
          <v-card-text>
            <div class="text-headline-large">
              {{ selectorData.clusterCount?.toLocaleString() }}
            </div>
            <div class="text-body-large">
              {{ plural('Cluster', selectorData.clusterCount) }}
            </div>
          </v-card-text>
        </v-card>
        <v-card
          v-else
          color="primary"
          variant="flat"
          min-width="150px"
        >
          <v-card-text>
            <div class="text-headline-large">
              {{ selectorData.selectorTotalResultCount?.toLocaleString() }}
            </div>
            <div class="text-body-large">
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
            <div class="text-headline-large">
              {{ transactionCount.toLocaleString() }}
            </div>
            <div class="text-body-large">
              {{ selectorType === SELECTOR_TYPE_HEURISTIC?'Transactions':'Stored Transactions' }}
            </div>
          </v-card-text>
        </v-card>
      </div>
      <named-divider
        title="Properties"
        title-class="text-body-large"
      />
      <div class="d-flex align-center flex-wrap itemContainer justify-center">
        <small-icon-item
          v-if="selectorData.heuristicTypeTitle"
          :title="selectorData.heuristicTypeTitle"
          :icon="mdiApplicationVariableOutline"
          tooltip="Type"
        />
        <small-icon-item
          v-if="selectorData.heuristicParameter"
          :title="selectorData.heuristicParameter"
          :icon="mdiTune"
          :tooltip="selectorData.heuristicParameterTitle?selectorData.heuristicParameterTitle:'Parameter'"
        />
        <small-icon-item
          v-if="selectorData.heuristicCustomClusters"
          :icon="mdiMerge"
          tooltip="Custom clusters"
        />
        <small-icon-item
          v-if="selectorData.heuristicExcludeAddresses"
          :icon="mdiPlaylistRemove"
          tooltip="Exclude addresses"
        />
        <small-icon-item
          v-if="selectorData.heuristicExcludeSpendingGaps"
          :icon="mdiClockAlertOutline"
          tooltip="Exclude spending gaps"
        />
        <small-icon-item
          v-if="selectorData.heuristicTimestamp"
          :title="new Date(selectorData.heuristicTimestamp).toLocaleDateString()"
          :icon="mdiCalendar"
          :tooltip="`Created ${new Date(selectorData.heuristicTimestamp).toLocaleString()}`"
        />
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
          v-if="selectorData.selectorTimestamp"
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
          title-class="text-body-large"
          :vertical-margin="0"
        />
        <div class="d-flex align-center justify-center">
          <svg
            id="selector_details_canvas"
            class="mt-3"
            style="max-width: 900px"
            :class="{'hide':!enoughDataForGraph}"
          />
        </div>
        <v-card-title v-if="!enoughDataForGraph">
          Not enough data to display diagram
        </v-card-title>
        <v-card-text v-if="!enoughDataForGraph && durationInMinutes > 0">
          {{ `Only ${durationInMinutes} ${plural('minute', durationInMinutes)} between oldest and most recent origin transaction.` }}
        </v-card-text>
        <v-card-text v-if="!enoughDataForGraph && durationInMinutes === 0">
          All origins occur at the same point of time.
        </v-card-text>
      </v-card>
      <template v-if="selectorData.transactions?.length > 0">
        <v-text-field
          v-model="tableSearchModel"
          label="Filter table"
          hide-details
        />
        <v-data-table
          :items="tableItems"
          :headers="tableHeaders"
          multi-sort
          :search="tableSearchModel"
          items-per-page="25"
        >
          <template #item.txhash="{item}">
            <td>
              <workspace-link
                style="max-width: 200px"
                :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
                       params: { id: item.txhash, blockchainMode: route.params.blockchainMode }}"
              >
                {{ item.txhash }}
              </workspace-link>
            </td>
          </template>
          <template #item.txtype="{item}">
            <td>
              <color-chip
                v-if="item.txtype"
                :title="item.txtype"
                :color="colorMap.get(item.txtype)"
                size="small"
              />
            </td>
          </template>
          <template #item.ts="{item}">
            <td v-tooltip="{'text': new Date(item.ts).toLocaleString(), 'location':'top', 'open-delay': 400}">
              {{ new Date(item.ts).toLocaleDateString() }}
            </td>
          </template>
          <template #item.actions="{item}">
            <v-btn-group density="compact">
              <v-btn
                v-tooltip="{'text': 'Select all transactions belonging to this cluster', 'location':'top', 'open-delay': 400}"
                variant="text"
                :icon="mdiSelectAll"
                @click="handleClusterSelected(item.cluster)"
              />
              <v-btn
                v-tooltip="{'text': 'Deselect all transactions belonging to this cluster', 'location':'top', 'open-delay': 400}"
                variant="text"
                :icon="mdiSelectRemove"
                @click="handleClusterDeselected(item.cluster)"
              />
            </v-btn-group>
          </template>
        </v-data-table>
      </template>
    </v-card-text>
  </v-card>
</template>

<script setup>
import {
	mdiAlertCircle,
	mdiApplicationVariableOutline,
	mdiArrowCollapseDown,
	mdiArrowLeft,
	mdiArrowRight,
	mdiCalendar,
	mdiCalendarEnd,
	mdiCalendarStart,
	mdiClockAlertOutline,
	mdiIncognitoOff,
	mdiMerge,
	mdiPlaylistRemove,
	mdiSelectAll,
	mdiSelectRemove,
	mdiTune,
} from '@mdi/js';
import {
	computed,
	onMounted,
	onUpdated,
	ref,
} from 'vue';
import {useRoute} from 'vue-router';
import BarChart from '@/d3Documents/barChart.js';
import NamedDivider from '@/components/common/NamedDivider.vue';
import {
	convertAmount,
	getTransactionColorMap,
	plural,
	setUndefinedTransactionColor,
} from '@/utilities/index.js';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';
import {
	ROUTE_NAME_TRANSACTION_PAGE,
	SELECTOR_ERROR_CODE_RESULT_LIMIT_EXCEEDED,
	SELECTOR_RESULT_LIMIT,
	SELECTOR_STATUS_ERROR,
	SELECTOR_TYPE_HEURISTIC,
} from '@/constants/index.js';
import {
	cashLeft,
	cashRight,
	sigmaLeft,
	sigmaRight,
	incognitoFilter,
} from '@/customIcons/index.js';
import ColorChip from '@/components/common/ColorChip.vue';
import SmallIconItem from '@/components/common/SmallIconItem.vue';

const props = defineProps({
	selectorType: {type: String, required: true},
	selectorData: {type: Object, required: true},
});

const route = useRoute();

const emit = defineEmits(['clusterSelected', 'clusterDeselected']);

const colorMap = getTransactionColorMap(route.params.blockchainMode);
setUndefinedTransactionColor(colorMap, undefined);
let svgBarChart = null;
const tableHeadersWithoutCluster = [
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

const tableHeadersWithCluster = [
	{
		key: 'txhash', title: 'Transaction', sortable: false,
	},
	{key: 'cluster', title: 'Cluster'},
	{key: 'txtype', title: 'Type'},
	{key: 'ts', title: 'Timestamp'},
	{
		key: 'actions', title: 'Actions', align: 'end', sortable: false,
	},
];

const enoughDataForGraph = ref(true);
const durationInMinutes = ref(0);
const tableSearchModel = ref('');
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

const tableHeaders = computed(() => tableItems.value.length > 0 && tableItems.value[0].cluster >= 0
	? tableHeadersWithCluster
	: tableHeadersWithoutCluster);

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

function handleClusterSelected(clusterID) {
	emit('clusterSelected', clusterID);
}

function handleClusterDeselected(clusterID) {
	emit('clusterDeselected', clusterID);
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
