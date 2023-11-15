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
              :icon="icon.mdiCardBulletedOutline"
            >
              <v-chip
                v-if="showExclusionAlert"
                :rounded="true"
                color="primary"
              >
                <template #append>
                  <v-icon
                    class="ms-1"
                    :icon="icon.mdiCloseCircle"
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
                    <IconItem
                      :icon="icon.mdiScaleBalance"
                      title="Balance"
                    >
                      {{ convertAmount(addressData.output_sum - addressData.input_sum) }}
                      {{ coinUnit }}
                    </IconItem>
                  </v-col>
                  <v-col
                    cols="12"
                    sm="4"
                  >
                    <IconItem
                      :icon="icon.mdiBankTransferIn"
                      title="Total amount received"
                    >
                      {{ convertAmount(addressData.output_sum) }}
                      {{ coinUnit }}
                    </IconItem>
                  </v-col>
                  <v-col>
                    <IconItem
                      :icon="icon.mdiBankTransferOut"
                      title="Total amount spent"
                    >
                      {{ convertAmount(addressData.input_sum) }}
                      {{ coinUnit }}
                    </IconItem>
                  </v-col>
                </v-row>
                <v-row>
                  <v-col
                    cols="12"
                    sm="4"
                  >
                    <IconItem
                      :icon="icon.mdiPound"
                      title="Outputs"
                    >
                      {{ addressData.output_count }}
                    </IconItem>
                  </v-col>
                  <v-col
                    cols="12"
                    sm="4"
                  >
                    <IconItem
                      :icon="icon.mdiPound"
                      title="Unspent outputs"
                    >
                      {{ addressData.output_count - addressData.input_count }}
                    </IconItem>
                  </v-col>
                  <v-col>
                    <IconItem
                      :icon="icon.mdiPound"
                      title="Coinbase outputs"
                    >
                      {{ addressData.coinbase_count }}
                    </IconItem>
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
                    v-model="sortAndFilter"
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
                          :to="{ name: transactionRoute,
                                 params: { id: item.input_transaction }}"
                        >
                          {{ shortenHash(item.input_transaction) }}
                        </router-link>
                      </template>
                      <template #item.output_transaction="{ item }">
                        <router-link
                          v-if="item.output_transaction"
                          :to="{ name: transactionRoute,
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

<script>
import {
	mdiCardBulletedOutline, mdiScaleBalance, mdiBankTransferIn,
	mdiBankTransferOut, mdiPound, mdiMerge, mdiDotsVertical, mdiDelete, mdiCloseCircle,
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

export default {
	name: 'AddressPage',
	components: {
		IconTitle,
		ClusterLookup, MixingActivity, SortAndFilter, IconItem,
		DeleteAddressExclusionDialog,
	},
	data() {
		return {
			icon: {
				mdiCardBulletedOutline,
				mdiScaleBalance,
				mdiBankTransferIn,
				mdiBankTransferOut,
				mdiPound,
				mdiMerge,
				mdiDotsVertical,
				mdiDelete,
				mdiCloseCircle,
			},
			coinUnit: COIN_UNIT,
			transactionRoute: ROUTE_NAME_TRANSACTION_PAGE,
			itemsPerPage: 20,
			addressHash: '',
			tab: null,
			deleteExclusionDialog: false,
			showExclusionAlert: false,
			isLoading: false,
			// EmptyResponse is only used for data loaded after the initial data load
			emptyResponse: false,
			sortAndFilter: {
				filter: [],
				order: 0,
			},
			table: {
				page: 1,
				headers: [
					{title: 'Received', key: 'output_transaction', sortable: false},
					{title: '', key: 'output_ts', sortable: false},
					{title: 'Sent', key: 'input_transaction', sortable: false},
					{title: '', key: 'input_ts', sortable: false},
					{title: 'Amount', key: 'amount', sortable: false},
				],
			},
		};
	},
	computed: {
		addressData: {
			get() {
				return this.$store.getters.getAddressData;
			},
			set(value) {
				this.$store.dispatch('setAddressData', value);
			},
		},
		session() {
			return this.$store.getters.getSession;
		},
		showAdvanced() {
			return isPrivilegedIdentity(this.session) || isAdminIdentity(this.session);
		},
		offset() {
			return this.table.page * this.itemsPerPage - this.itemsPerPage;
		},
	},
	watch: {
		addressHash() {
			// Only get exclusion status if this is an at least privileged user
			if (this.showAdvanced) {
				this.getExclusionStatus();
			}
		},
		addressData() {
			this.setInitialState();
		},
	},
	mounted() {
		this.setInitialState();
	},
	updated() {
		this.setInitialState();
	},
	methods: {
		shortenHash,
		convertAmount,
		isResponseValid(data) {
			return !(!data.type || data.type !== 'addr' || !data.payload || !data.payload.addr_outputs
          || data.payload.addr_outputs.length === 0);
		},
		async getTableData() {
			if (!this.addressData || this.addressHash === '') {
				return;
			}

			this.isLoading = true;

			try {
				const response = await this.dakar.data.addressOutputRangeAddressHashPost({
					addressHash: this.addressHash,
					options: {
						offset: this.offset,
						filter: this.sortAndFilter.filter,
						order: this.sortAndFilter.order,
					},
				});

				if (this.isResponseValid(response)) {
					this.addressData = response.payload;
					this.emptyResponse = false;
				} else {
					this.emptyResponse = true;
				}
			} catch (e) {
				handleError(this, e);
			}

			this.isLoading = false;
		},
		async getExclusionStatus() {
			if (this.addressHash === '') {
				return;
			}

			this.isLoading = true;

			try {
				const response = await this.dakar.addressExclusion.addressExclusionStatusAddressHashGet({addressHash: this.addressHash});
				this.showExclusionAlert = response.isExclusion;
			} catch (e) {
				handleError(this, e);
			}

			this.isLoading = false;
		},
		setInitialState() {
			let h = ' ';

			// Detect if address hash has changed
			if (this.addressData && this.addressData.addresshash
          && this.addressData.addresshash !== this.addressHash) {
				this.addressHash = this.addressData.addresshash;

				h = ` ${this.addressHash} `;

				this.resetSorting();
				this.table.page = 1;
			} else if (this.addressHash) {
				h = ` ${this.addressHash} `;
			}

			document.title = `Address${h}- ${PAGE_TITLE}`;
		},
		handleFilterOrSortChange() {
			this.table.page = 1;
			this.getTableData();
		},
		resetSorting() {
			if (this.sortAndFilter.order === 0 && this.sortAndFilter.filter.length === 0) {
				return;
			}

			this.sortAndFilter = {
				filter: [],
				order: 0,
			};
		},
		hideExclusionAlert() {
			this.showExclusionAlert = false;
		},
	},
};
</script>

<style scoped>

</style>
