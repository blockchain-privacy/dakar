<template>
  <v-container fluid v-if="this.data">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <v-card class="elevation-4">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title v-if="this.data">
              <v-icon>{{ icon.mdiCardBulletedOutline }}</v-icon>
              Address {{ this.data.addresshash }}
            </v-toolbar-title>
            <v-spacer></v-spacer>
            <v-btn
                class="mr-2"
                id="btn_show_mixing_activity"
                v-if="showClusterLookupEditor"
                outlined icon
                :to="{ name: mixingActivityPage, query: {address: this.addressHash} }">
              <v-icon>{{ icon.mdiChartBar }}</v-icon>
            </v-btn>
            <v-tooltip bottom activator="#btn_show_mixing_activity">
              <span>Show mixing activity</span>
            </v-tooltip>
            <v-btn
                id="btn_open_cluster_lookup"
                v-if="showClusterLookupEditor"
                class="mr-0" outlined icon
                :to="{ name: clusterLookupPage, query: {a1: this.addressHash} }">
              <v-icon>{{ icon.mdiMerge }}</v-icon>
            </v-btn>
            <v-tooltip bottom activator="#btn_open_cluster_lookup">
              <span>Open the cluster lookup for this address</span>
            </v-tooltip>
            <!-- hmi cluster lookup disabled for now -->
            <v-btn v-if="false"
                   id="btn_open_cluster_view"
                   style="margin-right: 0" outlined icon
                   :to="{ name: clusterViewRoute }">
              <v-icon>{{ icon.mdiGraph }}</v-icon>
            </v-btn>
            <v-tooltip bottom activator="#btn_open_cluster_view">
              <span>Open the cluster viewer for this address.</span>
            </v-tooltip>
          </v-toolbar>
          <v-card-text>
            <v-container>
              <v-row>
                <v-col>
                  <IconItem :icon="icon.mdiScaleBalance" title="Balance">
                    {{ convertAmount(this.data.output_sum - this.data.input_sum) }}
                    {{ this.coinUnit }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem :icon="icon.mdiBankTransferIn" title="Total amount received">
                    {{ convertAmount(this.data.output_sum) }}
                    {{ this.coinUnit }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem :icon="icon.mdiBankTransferOut" title="Total amount spent">
                    {{ convertAmount(this.data.input_sum) }}
                    {{ this.coinUnit }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <IconItem :icon="icon.mdiPound" title="Outputs">
                    {{ this.data.output_count }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem :icon="icon.mdiPound" title="Unspent outputs">
                    {{ this.data.output_count - this.data.input_count }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem :icon="icon.mdiPound" title="Coinbase outputs">
                    {{ this.data.coinbase_count }}
                  </IconItem>
                </v-col>
              </v-row>
            </v-container>
          </v-card-text>
        </v-card>

        <v-card class="mt-4">
          <v-card-title>
            Outputs
          </v-card-title>
          <v-card-text>
            <v-container>
              <v-row v-if="this.data.output_count > 1">
                <v-col>
                  <v-select
                      :disabled="this.isLoading"
                      :loading="this.isLoading?'primary':false"
                      style="max-width: 300px; min-width: 200px;"
                      v-model="combobox.selected.id"
                      :items="combobox.items"
                      item-value="id"
                      item-text="text"
                      label="Sort by"
                      v-on:change="handleSortAndFilter"
                  ></v-select>
                </v-col>
                <v-col>
                  <v-select
                      :disabled="this.isLoading"
                      :loading="this.isLoading?'primary':false"
                      style="max-width: 300px; min-width: 200px;"
                      v-model="filter.selected"
                      :items="filter.items"
                      item-value="id"
                      item-text="text"
                      label="Filter"
                      v-on:change="handleSortAndFilter"
                      multiple>
                    <template v-slot:selection="{ item }">
                      <v-chip small>
                        <span>{{ item.chip }}</span>
                      </v-chip>
                    </template>
                  </v-select>
                </v-col>
                <v-col v-if="this.isSortingByInput">
                  <v-alert type="info" text>Only spent outputs are shown.</v-alert>
                </v-col>
              </v-row>
              <v-sheet
                  v-if="!this.isLoading && !this.emptyResponse"
                  min-height="50"
                  class="fill-height"
                  color="transparent">
                <v-data-table
                    :headers="table.headers"
                    :items="this.data.addr_outputs"
                    :server-items-length="data.query_max_count"
                    :options.sync="table.options"
                    disable-sort
                    :items-per-page="itemsPerPage"
                    :footer-props="{itemsPerPageOptions:[itemsPerPage]}"
                    :loading="isLoading">
                  <template v-slot:[`item.input_transaction`]="{ item }">
                    <router-link :to="{ name: transactionRoute,
                    params: { id: item.input_transaction }}">
                      {{ shortenHash(item.input_transaction) }}
                    </router-link>
                  </template>
                  <template v-slot:[`item.output_transaction`]="{ item }">
                    <router-link :to="{ name: transactionRoute,
                    params: { id: item.output_transaction }}">
                      {{ shortenHash(item.output_transaction) }}
                    </router-link>
                  </template>
                  <template v-slot:[`item.input_ts`]="{ item }">
                    {{ item.input_ts ? new Date(item.input_ts).toLocaleString() : '' }}
                  </template>
                  <template v-slot:[`item.output_ts`]="{ item }">
                    {{ item.output_ts ? new Date(item.output_ts).toLocaleString() : '' }}
                  </template>
                  <template v-slot:[`item.amount`]="{ item }">
                    {{ convertAmount(item.amount) }}
                  </template>
                </v-data-table>
              </v-sheet>
              <v-row v-if="this.emptyResponse">
                <v-col class="d-flex justify-center">
                  <p class="text-h6">No outputs found</p>
                </v-col>
              </v-row>
            </v-container>
          </v-card-text>
        </v-card>

      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import {
  mdiCardBulletedOutline, mdiScaleBalance, mdiBankTransferIn,
  mdiBankTransferOut, mdiPound, mdiMerge, mdiChartBar,
} from '@mdi/js';
import {
  convertAmount, doPost, handleError, isAdminUser, isPrivilegedUser, shortenHash,
} from '../../utilities';
import {
  PAGE_TITLE, ROUTE_NAME_TRANSACTION_PAGE, ROUTE_NAME_CLUSTER_LOOKUP_PAGE,
  ROUTE_ADDRESS_OUTPUT_RANGE, COIN_UNIT, ROUTE_NAME_CLUSTER_VIEW_PAGE,
  ROUTE_NAME_MIXING_ACTIVITY,
} from '../../constants';
import IconItem from '../common/IconItem.vue';

export default {
  name: 'AddressLookup',
  components: {
    IconItem,
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
        mdiChartBar,
      },
      coinUnit: COIN_UNIT,
      transactionRoute: ROUTE_NAME_TRANSACTION_PAGE,
      clusterViewRoute: ROUTE_NAME_CLUSTER_VIEW_PAGE,
      clusterLookupPage: ROUTE_NAME_CLUSTER_LOOKUP_PAGE,
      mixingActivityPage: ROUTE_NAME_MIXING_ACTIVITY,
      combobox: {
        selected: {
          id: 0,
        },
        items: [
          { id: 0, text: 'Ascending by output date', disabled: false },
          { id: 1, text: 'Descending by output date', disabled: false },
          { divider: true },
          { id: 2, text: 'Ascending by input date', disabled: false },
          { id: 3, text: 'Descending by input date', disabled: false },
          { divider: true },
          { id: 4, text: 'Ascending by amount', disabled: false },
          { id: 5, text: 'Descending by amount', disabled: false },
        ],
      },
      filter: {
        selected: [],
        items: [
          {
            id: 0, text: 'Only show coinbase outputs', chip: 'Coinbase outputs', disabled: false,
          },
          {
            id: 1, text: 'Only show unspent outputs', chip: 'Unspent outputs', disabled: false,
          },
        ],
      },
      // default sort order: ascending by output timestamp
      sortOrder: 0,
      itemsPerPage: 20,
      addressHash: '',
      isLoading: false,
      isLoadingMore: false,
      isSortingByInput: false,
      // emptyResponse is only used for data loaded after the initial data load
      emptyResponse: false,
      table: {
        options: {},
        headers: [
          { text: 'Received', value: 'output_transaction', sortable: false },
          { text: '', value: 'output_ts' },
          { text: 'Sent', value: 'input_transaction', sortable: false },
          { text: '', value: 'input_ts' },
          { text: 'Amount', value: 'amount' },
        ],
      },
    };
  },
  methods: {
    shortenHash,
    isResponseValid(data) {
      return !(!data.type || data.type !== 'addr' || !data.payload || !data.payload.addr_outputs
          || data.payload.addr_outputs.length === 0);
    },
    getTableData() {
      if (!this.data || this.addressHash === '') return;

      this.isLoadingMore = true;

      doPost(ROUTE_ADDRESS_OUTPUT_RANGE, this.$router, this.$store,
        {
          offset: this.offset,
          order: this.sortOrder,
          filter: this.filter.selected,
        },
        this.addressHash)
        .then((data) => {
          if (!this.isResponseValid(data)) {
            this.emptyResponse = true;
            return;
          }

          this.data.addr_outputs = data.payload.addr_outputs;
          this.$store.dispatch('resetMessages');
          this.emptyResponse = false;
        })
        .catch((e) => {
          handleError(this.$store, e);
        })
        .finally(() => {
          this.isLoadingMore = false;
        });
    },
    updateSortState() {
      let isUnspentFilterSelected = false;

      if (this.data && this.data.input_count === 0) {
        isUnspentFilterSelected = true;
      } else {
        this.filter.selected.some((d) => {
          if (d === 1) {
            isUnspentFilterSelected = true;
            // break
            return true;
          }
          return false;
        });
      }
      this.combobox.items.forEach((d) => {
        if (d.id === 2 || d.id === 3) {
          d.disabled = isUnspentFilterSelected;
        }
      });
    },
    updateFilterState() {
      let disableUnspentFilter = false;
      if (this.data && this.data.output_count - this.data.input_count === 0) {
        disableUnspentFilter = true;
      } else {
        const selection = this.combobox.selected.id;
        if (selection === 2 || selection === 3) {
          disableUnspentFilter = true;
        }
      }

      let disableCoinbaseFilter = false;
      if (this.data && (this.data.coinbase_count === 0
          || this.data.coinbase_count === this.data.output_count)) {
        disableCoinbaseFilter = true;
      }

      this.filter.items.forEach((d) => {
        if (d.id === 0) {
          d.disabled = disableCoinbaseFilter;
        } else if (d.id === 1) {
          d.disabled = disableUnspentFilter;
        }
      });
    },
    handleSortAndFilter() {
      this.updateSortState();
      this.updateFilterState();
      this.isLoading = true;
      this.sortOrder = this.combobox.selected.id;
      this.table.options.page = 1;

      this.isSortingByInput = this.sortOrder === 2 || this.sortOrder === 3;

      doPost(ROUTE_ADDRESS_OUTPUT_RANGE, this.$router, this.$store,
        { offset: this.offset, order: this.sortOrder, filter: this.filter.selected },
        this.addressHash)
        .then((data) => {
          if (!this.isResponseValid(data)) {
            this.emptyResponse = true;
            return;
          }

          this.data = data.payload;
          this.$store.dispatch('resetMessages');
          this.emptyResponse = false;
        })
        .catch((e) => {
          handleError(this.$store, e);
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
    convertAmount,
    setAddressHash() {
      let h = ' ';
      if (this.data && this.data.addresshash && this.data.addresshash !== this.addressHash) {
        this.addressHash = this.data.addresshash;
        h = ` ${this.addressHash} `;
      } else if (this.addressHash) {
        h = ` ${this.addressHash} `;
      }
      document.title = `Address${h}- ${PAGE_TITLE}`;
    },
  },
  computed: {
    data: {
      get() {
        return this.$store.getters.getAddressData;
      },
      set(value) {
        this.$store.dispatch('setAddressData', value);
      },
    },
    userData() {
      return this.$store.getters.getActiveUser;
    },
    showClusterLookupEditor() {
      return isPrivilegedUser(this.userData) || isAdminUser(this.userData);
    },
    offset() {
      return this.table.options.page * this.itemsPerPage - this.itemsPerPage;
    },
  },
  mounted() {
    this.setAddressHash();
    this.updateSortState();
    this.updateFilterState();
  },
  updated() {
    this.setAddressHash();
    this.updateSortState();
    this.updateFilterState();
  },
  watch: {
    'table.options.page': {
      handler() {
        this.getTableData();
      },
    },
  },
};
</script>

<style scoped>

</style>
