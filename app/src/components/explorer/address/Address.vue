<template>
  <div v-if="addressHash">
    <v-card variant="text">
      <icon-title
        v-if="showTitleBar"
        :title="`Address ${addressHash}`"
        :icon="mdiCardBulletedOutline"
      >
        <exclusion-chip :address-hash="addressHash" />
      </icon-title>
      <v-card-text>
        <v-container>
          <v-row>
            <v-col
              cols="12"
              sm="4"
            >
              <icon-item
                :icon="mdiScaleBalance"
                title="Balance"
              >
                {{ convertAmount(outputSum - inputSum) }}
                {{ coinUnit }}
              </icon-item>
            </v-col>
            <v-col
              cols="12"
              sm="4"
            >
              <icon-item
                :icon="mdiBankTransferIn"
                title="Total amount received"
              >
                {{ convertAmount(outputSum) }}
                {{ coinUnit }}
              </icon-item>
            </v-col>
            <v-col>
              <icon-item
                :icon="mdiBankTransferOut"
                title="Total amount spent"
              >
                {{ convertAmount(inputSum) }}
                {{ coinUnit }}
              </icon-item>
            </v-col>
          </v-row>
          <v-row>
            <v-col
              cols="12"
              sm="4"
            >
              <icon-item
                :icon="mdiPound"
                title="Outputs"
              >
                {{ outputCount }}
              </icon-item>
            </v-col>
            <v-col
              cols="12"
              sm="4"
            >
              <icon-item
                :icon="mdiPound"
                title="Unspent outputs"
              >
                {{ outputCount - inputCount }}
              </icon-item>
            </v-col>
          </v-row>
        </v-container>
      </v-card-text>
    </v-card>
    <v-tabs
      v-model="tab"
      class="mt-4"
      fixed-tabs
    >
      <v-tab>
        Outputs
      </v-tab>
      <v-tab :disabled="!showAdvanced">
        Clusters
      </v-tab>
      <v-tab :disabled="!showAdvanced">
        Mixing Activity
      </v-tab>
    </v-tabs>
    <v-window
      v-model="tab"
      :touch="false"
    >
      <v-window-item>
        <v-card variant="text">
          <v-card-text>
            <v-progress-linear
              v-if="isLoading"
              indeterminate
            />
            <v-alert
              v-else-if="!isOutputManipulationSupported"
              variant="tonal"
              type="info"
              class="mb-2"
            >
              Sorting and filtering is not available for this address because of its high number of outputs.
            </v-alert>
            <sort-and-filter
              v-else-if="outputCount > 1"
              v-model="sortAndFilterModel"
              :loading="isLoading"
              :output-count="outputCount"
              :input-count="inputCount"
              @change="handleFilterOrSortChange"
            />
            <v-sheet
              v-if="!isLoading && !emptyResponse"
              min-height="50"
              class="fill-height"
              color="transparent"
            >
              <v-data-table-server
                v-model:page="table.page"
                :headers="table.headers"
                :items="outputItems"
                :items-length="queryMaxCount"
                :items-per-page="itemsPerPage"
                :footer-props="{itemsPerPageOptions:[itemsPerPage]}"
                :loading="isLoading"
                :items-per-page-options="[{value:20, title:'20'}]"
                @update:page="getTableData"
              >
                <template #item.inputTransactionHash="{ item }">
                  <workspace-link
                    v-if="item.inputTransactionHash"
                    style="max-width:200px"
                    :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
                           params: { id: item.inputTransactionHash, blockchainMode: route.params.blockchainMode }}"
                  >
                    {{ item.inputTransactionHash }}
                  </workspace-link>
                </template>
                <template #item.outputTransactionHash="{ item }">
                  <workspace-link
                    v-if="item.outputTransactionHash"
                    style="max-width:200px"
                    :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
                           params: { id: item.outputTransactionHash, blockchainMode: route.params.blockchainMode }}"
                  >
                    {{ item.outputTransactionHash }}
                  </workspace-link>
                </template>
                <template #item.inputTimestamp="{ item }">
                  {{ item.inputTimestamp ? new Date(item.inputTimestamp).toLocaleString() : '' }}
                </template>
                <template #item.outputTimestamp="{ item }">
                  {{ item.outputTimestamp ? new Date(item.outputTimestamp).toLocaleString() : '' }}
                </template>
                <template #item.amount="{ item }">
                  {{ convertAmount(item.amount) }} {{ coinUnit }}
                </template>
              </v-data-table-server>
            </v-sheet>
            <v-row v-if="emptyResponse">
              <v-col class="d-flex justify-center">
                <p class="text-h6">
                  No outputs found
                </p>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>
      </v-window-item>
      <v-window-item>
        <cluster-lookup :address-hash="addressHash" />
      </v-window-item>
      <v-window-item>
        <mixing-activity :address-hash="addressHash" />
      </v-window-item>
    </v-window>
  </div>
</template>
<script setup>
import {
	mdiBankTransferIn,
	mdiBankTransferOut,
	mdiCardBulletedOutline,
	mdiPound,
	mdiScaleBalance,
} from '@mdi/js';
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import {
	convertAmount, getCoinUnit, getDakarClient, handleError, isAdminIdentity, isPrivilegedIdentity,
} from '@/utilities';
import MixingActivity from '@/components/explorer/address/MixingActivity.vue';
import IconItem from '@/components/common/IconItem.vue';
import SortAndFilter from '@/components/explorer/address/SortAndFilter.vue';
import ClusterLookup from '@/components/explorer/address/ClusterLookup.vue';
import IconTitle from '@/components/common/IconTitle.vue';
import {
	computed, onMounted, onUpdated, ref,
} from 'vue';
import {useMsgStore} from '@/pinia/msg';
import {useRoute} from 'vue-router';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local';
import ExclusionChip from '@/components/explorer/address/ExclusionChip.vue';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';

const props = defineProps({
	addressData: {type: Object, required: true},
	showTitleBar: {type: Boolean, required: false},
});

const route = useRoute();
const context = {addMessage: useMsgStore().addMessage, $route: route};
const {session} = storeToRefs(useLocalStore());
const dakar = getDakarClient(route.params.blockchainMode);

const isLoading = ref(false);
const tab = ref(null);

const addressHash = ref('');
const inputSum = ref(-1);
const outputSum = ref(-1);
const inputCount = ref(-1);
const outputCount = ref(-1);
const queryMaxCount = ref(-1);
const outputItems = ref([]);
const isOutputManipulationSupported = ref(false);

const itemsPerPage = 20;
// EmptyResponse is only used for data loaded after the initial data load
const emptyResponse = ref(false);
const sortAndFilterModel = ref({
	filter: [],
	order: 0,
});
const table = ref({
	page: 1,
	headers: [
		{title: 'Received', key: 'outputTransactionHash', sortable: false},
		{title: '', key: 'outputTimestamp', sortable: false},
		{title: 'Sent', key: 'inputTransactionHash', sortable: false},
		{title: '', key: 'inputTimestamp', sortable: false},
		{title: 'Amount', key: 'amount', sortable: false},
	],
});

// Computed
const offset = computed(() => (table.value.page * itemsPerPage) - itemsPerPage);
const showAdvanced = computed(() => isPrivilegedIdentity(session.value, route.params.blockchainMode)
	|| isAdminIdentity(session.value, route.params.blockchainMode));
const coinUnit = computed(() => getCoinUnit(route.params.blockchainMode));

// Hooks
onMounted(() => {
	init();
});

onUpdated(() => {
	init();
});

// Functions
function dataToRef(data) {
	addressHash.value = data.addresshash;
	outputItems.value = data.outputs;
	inputSum.value = data.inputSum;
	outputSum.value = data.outputSum;
	inputCount.value = data.inputCount;
	outputCount.value = data.outputCount;
	queryMaxCount.value = data.queryMaxCount === 0 ? data.outputCount : data.queryMaxCount;
	isOutputManipulationSupported.value = data.isOutputManipulationSupported;
}

function init() {
	if (props.addressData && addressHash.value !== props.addressData.addresshash) {
		dataToRef(props.addressData);
		resetSorting();
		emptyResponse.value = false;
		table.value.page = 1;
	}
}

function handleFilterOrSortChange() {
	table.value.page = 1;
	getTableData();
}

function resetSorting() {
	if (sortAndFilterModel.value.order === 0 && sortAndFilterModel.value.filter.length === 0) {
		return;
	}

	sortAndFilterModel.value = {
		filter: [],
		order: 0,
	};
}

async function getTableData() {
	if (!props.addressData) {
		return;
	}

	isLoading.value = true;

	try {
		const response = await dakar.data.blockchainOutputsHashPost({
			hash: addressHash.value,
			options: {
				offset: offset.value,
				filter: sortAndFilterModel.value.filter,
				order: sortAndFilterModel.value.order,
			},
		});

		if (response.payload?.outputs?.length > 0) {
			dataToRef(response.payload);
			emptyResponse.value = false;
		} else {
			emptyResponse.value = true;
		}
	} catch (e) {
		handleError(context, e);
	}

	isLoading.value = false;
}

</script>
<style scoped>

</style>
