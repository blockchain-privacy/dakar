<template>
  <div v-if="data">
    <v-card variant="text">
      <icon-title
        :title="`Address ${data.addresshash}`"
        :icon="mdiCardBulletedOutline"
      >
        <v-chip
          v-if="showExclusionAlert"
          :rounded="true"
          color="primary"
        >
          <template #append>
            <v-icon
              class="ms-1"
              :icon="mdiCloseCircle"
              @click="deleteExclusionDialog = true"
            />
          </template>
          <span id="address_excluded">
            Excluded
          </span>
          <v-tooltip
            activator="#address_excluded"
            location="bottom"
          >
            This address is part of your address exclusion list.
            Click on the X to remove it from the list.
          </v-tooltip>
        </v-chip>
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
                {{ COIN_UNIT }}
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
                {{ COIN_UNIT }}
              </icon-item>
            </v-col>
            <v-col>
              <icon-item
                :icon="mdiBankTransferOut"
                title="Total amount spent"
              >
                {{ convertAmount(data.input_sum) }}
                {{ COIN_UNIT }}
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
      :fixed-tabs="true"
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
              :data-available="true"
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
                  <router-link
                    v-if="item.input_transaction"
                    :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
                           params: { id: item.input_transaction }}"
                  >
                    {{ shortenHash(item.input_transaction) }}
                  </router-link>
                </template>
                <template #item.output_transaction="{ item }">
                  <router-link
                    v-if="item.output_transaction"
                    :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
                           params: { id: item.output_transaction }}"
                  >
                    {{ shortenHash(item.output_transaction) }}
                  </router-link>
                </template>
                <template #item.input_ts="{ item }">
                  {{ item.input_ts ? new Date(item.input_ts).toLocaleString() : '' }}
                </template>
                <template #item.output_ts="{ item }">
                  {{ item.output_ts ? new Date(item.output_ts).toLocaleString() : '' }}
                </template>
                <template #item.amount="{ item }">
                  {{ convertAmount(item.amount) }}
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
    <delete-address-exclusion-dialog
      v-model="deleteExclusionDialog"
      :address-hash="data.addresshash"
      @deleted="hideExclusionAlert"
    />
  </div>
</template>
<script setup>
import {
	mdiBankTransferIn,
	mdiBankTransferOut,
	mdiCardBulletedOutline,
	mdiCloseCircle,
	mdiPound,
	mdiScaleBalance,
} from '@mdi/js';
import {COIN_UNIT, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import {convertAmount, handleError, isAdminIdentity, isPrivilegedIdentity, shortenHash} from '@/utilities';
import MixingActivity from '@/components/explorer/address/MixingActivity.vue';
import IconItem from '@/components/common/IconItem.vue';
import SortAndFilter from '@/components/explorer/address/SortAndFilter.vue';
import ClusterLookup from '@/components/explorer/address/ClusterLookup.vue';
import IconTitle from '@/components/common/IconTitle.vue';
import DeleteAddressExclusionDialog from '@/components/tools/addressExclusions/DeleteAddressExclusionDialog.vue';
import {computed, inject, onMounted, ref} from 'vue';
import {useMsgStore} from '@/pinia/msg';
import {useRoute} from 'vue-router';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local';

const props = defineProps({
	addressData: {type: Object, required: true},
});

const dakar = inject('dakar');
const route = useRoute();
const context = {addMessage: useMsgStore().addMessage, $route: route};
const {session} = storeToRefs(useLocalStore());

const showExclusionAlert = ref(false);
const isLoading = ref(false);
const tab = ref(null);
const deleteExclusionDialog = ref(false);
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
const offset = computed(() => table.value.page * itemsPerPage - itemsPerPage);
const showAdvanced = computed(() => isPrivilegedIdentity(session.value) || isAdminIdentity(session.value));

// Hooks
onMounted(() => {
	if (props.addressData) {
		data.value = props.addressData;
		getExclusionStatus();
		resetSorting();
		table.value.page = 1;
	}
});

// Functions
function isResponseValid(newData) {
	return !(!newData.type || newData.type !== 'addr' || !newData.payload || !newData.payload.addr_outputs
    || newData.payload.addr_outputs.length === 0);
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

function hideExclusionAlert() {
	showExclusionAlert.value = false;
}

async function getTableData() {
	if (!props.addressData) {
		return;
	}

	isLoading.value = true;

	try {
		const response = await dakar.data.addressOutputRangeAddressHashPost({
			addressHash: data.value.addresshash,
			options: {
				offset: offset.value,
				filter: sortAndFilterModel.value.filter,
				order: sortAndFilterModel.value.order,
			},
		});

		if (isResponseValid(response)) {
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

async function getExclusionStatus() {
	isLoading.value = true;

	try {
		const response = await dakar.addressExclusion.addressExclusionStatusAddressHashGet({addressHash: data.value.addresshash});
		showExclusionAlert.value = response.isExclusion;
	} catch (e) {
		handleError(context, e);
	}

	isLoading.value = false;
}
</script>
<style scoped>

</style>
