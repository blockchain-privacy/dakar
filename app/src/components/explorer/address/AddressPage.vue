<template>
  <v-container :fluid="true">
    <v-row
      align="center"
      justify="center"
    >
      <v-col
        cols="12"
        sm="12"
        md="12"
        lg="9"
        xl="8"
      >
        <template v-if="addressData">
          <v-card variant="text">
            <icon-title
              :title="`Address
              ${addressData.addresshash}`"
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
                      {{ convertAmount(addressData.output_sum - addressData.input_sum) }}
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
                      {{ convertAmount(addressData.output_sum) }}
                      {{ COIN_UNIT }}
                    </icon-item>
                  </v-col>
                  <v-col>
                    <icon-item
                      :icon="mdiBankTransferOut"
                      title="Total amount spent"
                    >
                      {{ convertAmount(addressData.input_sum) }}
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
                      {{ addressData.output_count }}
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
                      {{ addressData.output_count - addressData.input_count }}
                    </icon-item>
                  </v-col>
                  <v-col>
                    <icon-item
                      :icon="mdiPound"
                      title="Coinbase outputs"
                    >
                      {{ addressData.coinbase_count }}
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
                    v-if="addressData?.output_count > 1"
                    v-model="sortAndFilterModel"
                    :loading="isLoading"
                    :output-count="addressData.output_count"
                    :input-count="addressData.input_count"
                    :coinbase-count="addressData.coinbase_count"
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
                      :items="addressData.addr_outputs"
                      :items-length="addressData.query_max_count"
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
              <cluster-lookup :address-hash="addressHash" />
            </v-window-item>
            <v-window-item>
              <mixing-activity :address-hash="addressHash" />
            </v-window-item>
          </v-window>
        </template>
        <v-skeleton-loader
          v-else
          class="mx-auto"
          type="list-item-three-line, list-item-three-line, list-item-three-line"
        />
      </v-col>
    </v-row>
    <delete-address-exclusion-dialog
      v-model="deleteExclusionDialog"
      :address-hash="addressHash"
      @deleted="hideExclusionAlert"
    />
  </v-container>
</template>

<script setup>
import {
	mdiCardBulletedOutline, mdiScaleBalance, mdiBankTransferIn,
	mdiBankTransferOut, mdiPound, mdiCloseCircle,
} from '@mdi/js';
import {
	convertAmount,
	handleError,
	shortenHash,
	isPrivilegedIdentity,
	isAdminIdentity,
} from '@/utilities';
import {PAGE_TITLE, ROUTE_NAME_TRANSACTION_PAGE, COIN_UNIT} from '@/constants';
import IconItem from '../../common/IconItem.vue';
import SortAndFilter from './SortAndFilter.vue';
import MixingActivity from './MixingActivity.vue';
import ClusterLookup from './ClusterLookup.vue';
import DeleteAddressExclusionDialog from '../../tools/addressExclusions/DeleteAddressExclusionDialog.vue';
import IconTitle from '@/components/common/IconTitle.vue';
import {computed, inject, onMounted, onUpdated, ref, watch} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import {useExplorerStore} from '@/pinia/explorer';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local';

const dakar = inject('dakar');
const route = useRoute();
const {address: addressData} = storeToRefs(useExplorerStore());
const {session} = storeToRefs(useLocalStore());
const context = {addMessage: useMsgStore().addMessage, $route: route};

const itemsPerPage = 20;

const addressHash = ref('');
const tab = ref(null);
const deleteExclusionDialog = ref(false);
const showExclusionAlert = ref(false);
const isLoading = ref(false);
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
const showAdvanced = computed(() => isPrivilegedIdentity(session.value) || isAdminIdentity(session.value));
const offset = computed(() => table.value.page * itemsPerPage - itemsPerPage);

// Watchers
watch(addressHash, () => {
	// Only get exclusion status if this is an at least privileged user
	if (showAdvanced.value) {
		getExclusionStatus();
	}
});

watch(addressData, () => {
	setInitialState();
});

// Hooks
onMounted(() => {
	setInitialState();
});

onUpdated(() => {
	setInitialState();
});

// Functions
function isResponseValid(data) {
	return !(!data.type || data.type !== 'addr' || !data.payload || !data.payload.addr_outputs
      || data.payload.addr_outputs.length === 0);
}

async function getTableData() {
	if (!addressData.value || addressHash.value === '') {
		return;
	}

	isLoading.value = true;

	try {
		const response = await dakar.data.addressOutputRangeAddressHashPost({
			addressHash: addressHash.value,
			options: {
				offset: offset.value,
				filter: sortAndFilterModel.value.filter,
				order: sortAndFilterModel.value.order,
			},
		});

		if (isResponseValid(response)) {
			addressData.value = response.payload;
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
	if (addressHash.value === '') {
		return;
	}

	isLoading.value = true;

	try {
		const response = await dakar.addressExclusion.addressExclusionStatusAddressHashGet({addressHash: addressHash.value});
		showExclusionAlert.value = response.isExclusion;
	} catch (e) {
		handleError(context, e);
	}

	isLoading.value = false;
}

function setInitialState() {
	let h = ' ';

	// Detect if address hash has changed
	if (addressData.value && addressData.value.addresshash
      && addressData.value.addresshash !== addressHash.value) {
		addressHash.value = addressData.value.addresshash;

		h = ` ${addressHash.value} `;

		resetSorting();
		table.value.page = 1;
	} else if (addressHash.value) {
		h = ` ${addressHash.value} `;
	}

	document.title = `Address${h}- ${PAGE_TITLE}`;
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

</script>

<style scoped>

</style>
