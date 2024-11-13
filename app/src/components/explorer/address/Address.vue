<template>
  <div v-if="data">
    <v-card variant="text">
      <icon-title
        v-if="showTitleBar"
        :title="`Address ${data.addresshash}`"
        :icon="mdiCardBulletedOutline"
      >
        <exclusion-chip :address-hash="data.addresshash" />
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
                {{ convertAmount(data.output_sum - data.input_sum) }}
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
                {{ convertAmount(data.output_sum) }}
                {{ coinUnit }}
              </icon-item>
            </v-col>
            <v-col>
              <icon-item
                :icon="mdiBankTransferOut"
                title="Total amount spent"
              >
                {{ convertAmount(data.input_sum) }}
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
                {{ data.output_count }}
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
                {{ data.output_count - data.input_count }}
              </icon-item>
            </v-col>
            <v-col>
              <icon-item
                :icon="mdiPound"
                title="Coinbase outputs"
              >
                {{ data.coinbase_count }}
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
            <sort-and-filter
              v-if="data?.output_count > 1"
              v-model="sortAndFilterModel"
              :loading="isLoading"
              :output-count="data.output_count"
              :input-count="data.input_count"
              :coinbase-count="data.coinbase_count"
              data-available
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
                :items="data.addr_outputs"
                :items-length="data.query_max_count"
                :items-per-page="itemsPerPage"
                :footer-props="{itemsPerPageOptions:[itemsPerPage]}"
                :loading="isLoading"
                :items-per-page-options="[{value:20, title:'20'}]"
                @update:page="getTableData"
              >
                <template #item.input_transaction="{ item }">
                  <workspace-link
                    v-if="item.input_transaction"
                    style="max-width:200px"
                    :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
                           params: { id: item.input_transaction, blockchainMode: getSettings.blockchainMode }}"
                  >
                    {{ item.input_transaction }}
                  </workspace-link>
                </template>
                <template #item.output_transaction="{ item }">
                  <workspace-link
                    v-if="item.output_transaction"
                    style="max-width:200px"
                    :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
                           params: { id: item.output_transaction, blockchainMode: getSettings.blockchainMode }}"
                  >
                    {{ item.output_transaction }}
                  </workspace-link>
                </template>
                <template #item.input_ts="{ item }">
                  {{ item.input_ts ? new Date(item.input_ts).toLocaleString() : '' }}
                </template>
                <template #item.output_ts="{ item }">
                  {{ item.output_ts ? new Date(item.output_ts).toLocaleString() : '' }}
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
        <cluster-lookup :address-hash="data.addresshash" />
      </v-window-item>
      <v-window-item>
        <mixing-activity :address-hash="data.addresshash" />
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
const {session, getSettings} = storeToRefs(useLocalStore());
const dakar = getDakarClient(getSettings.value.blockchainMode);

const isLoading = ref(false);
const tab = ref(null);

const data = ref();
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
		{title: 'Received', key: 'output_transaction', sortable: false},
		{title: '', key: 'output_ts', sortable: false},
		{title: 'Sent', key: 'input_transaction', sortable: false},
		{title: '', key: 'input_ts', sortable: false},
		{title: 'Amount', key: 'amount', sortable: false},
	],
});

// Computed
const offset = computed(() => (table.value.page * itemsPerPage) - itemsPerPage);
const showAdvanced = computed(() => isPrivilegedIdentity(session.value) || isAdminIdentity(session.value));
const coinUnit = computed(() => getCoinUnit(getSettings.value.blockchainMode));

// Hooks
onMounted(() => {
	init();
});

onUpdated(() => {
	init();
});

// Functions
function init() {
	if (props.addressData) {
		data.value = props.addressData;
		resetSorting();
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
			hash: data.value.addresshash,
			options: {
				offset: offset.value,
				filter: sortAndFilterModel.value.filter,
				order: sortAndFilterModel.value.order,
			},
		});

		if (response.payload?.addr_outputs?.length > 0) {
			data.value = response.payload;
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
