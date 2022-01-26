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
            <v-tooltip bottom activator="#btn_show_mixing_activity">
              <span>Show mixing activity</span>
            </v-tooltip>
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
        <v-card class="mt-4 elevation-4">
          <v-tabs v-model="tab" fixed-tabs>
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
          <v-tabs-items v-model="tab">
            <v-tab-item>
              <v-card flat>
                <v-card-text>
                  <sort-and-filter
                      v-model="sortAndFilter"
                      v-if="this.data.output_count > 1"
                      :loading="isLoading"
                      :output-count="this.data.output_count"
                      :input-count="this.data.input_count"
                      :coinbase-count="this.data.coinbase_count"
                      :data-available="this.data !== undefined"
                      @change="handleFilterOrSortChange"
                  />
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
                        <router-link v-if="item.input_transaction"
                                     :to="{ name: transactionRoute,
                    params: { id: item.input_transaction }}">
                          {{ shortenHash(item.input_transaction) }}
                        </router-link>
                      </template>
                      <template v-slot:[`item.output_transaction`]="{ item }">
                        <router-link v-if="item.output_transaction"
                                     :to="{ name: transactionRoute,
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
                </v-card-text>
              </v-card>
            </v-tab-item>
            <v-tab-item>
              <cluster-lookup :addressHash="addressHash" />
            </v-tab-item>
            <v-tab-item>
              <mixing-activity :address-hash="addressHash" />
            </v-tab-item>
          </v-tabs-items>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import {
  mdiCardBulletedOutline, mdiScaleBalance, mdiBankTransferIn,
  mdiBankTransferOut, mdiPound, mdiMerge,
} from '@mdi/js';
import {
  convertAmount, doPost, handleError, isAdminUser, isPrivilegedUser, shortenHash,
} from '../../../utilities';
import {
  PAGE_TITLE, ROUTE_NAME_TRANSACTION_PAGE,
  ROUTE_ADDRESS_OUTPUT_RANGE, COIN_UNIT, ROUTE_NAME_CLUSTER_VIEW_PAGE,
} from '../../../constants';
import IconItem from '../../common/IconItem.vue';
import SortAndFilter from './SortAndFilter.vue';
import MixingActivity from './MixingActivity.vue';
import ClusterLookup from './ClusterLookup.vue';

export default {
  name: 'AddressLookup',
  components: {
    ClusterLookup, MixingActivity, SortAndFilter, IconItem,
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
      },
      coinUnit: COIN_UNIT,
      transactionRoute: ROUTE_NAME_TRANSACTION_PAGE,
      clusterViewRoute: ROUTE_NAME_CLUSTER_VIEW_PAGE,
      itemsPerPage: 20,
      addressHash: '',
      tab: null,
      isLoading: false,
      // emptyResponse is only used for data loaded after the initial data load
      emptyResponse: false,
      sortAndFilter: {
        filter: [],
        order: 0,
      },
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
    showAdvanced() {
      return isPrivilegedUser(this.userData) || isAdminUser(this.userData);
    },
    offset() {
      return this.table.options.page * this.itemsPerPage - this.itemsPerPage;
    },
  },
  methods: {
    shortenHash,
    convertAmount,
    isResponseValid(data) {
      return !(!data.type || data.type !== 'addr' || !data.payload || !data.payload.addr_outputs
          || data.payload.addr_outputs.length === 0);
    },
    getTableData() {
      if (!this.data || this.addressHash === '') return;
      this.isLoading = true;
      doPost(ROUTE_ADDRESS_OUTPUT_RANGE, this.$router, this.$store,
        {
          offset: this.offset,
          order: this.sortAndFilter.order,
          filter: this.sortAndFilter.filter,
        }, this.addressHash)
        .then((data) => {
          if (!this.isResponseValid(data)) {
            this.emptyResponse = true;
            return;
          }

          this.data = data.payload;

          // this.data.addr_outputs = data.payload.addr_outputs;
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
    handleFilterOrSortChange() {
      this.table.options.page = 1;
      this.getTableData();
    },
    resetSorting() {
      if (this.sortAndFilter.order === 0 && this.sortAndFilter.filter.length === 0) return;

      this.sortAndFilter = {
        filter: [],
        order: 0,
      };
    },
  },
  mounted() {
    this.setAddressHash();
  },
  updated() {
    this.setAddressHash();
    this.resetSorting();
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
