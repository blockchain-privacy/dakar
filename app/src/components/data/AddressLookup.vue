<template>
  <v-container fluid v-if="this.data">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <v-card class="elevation-12">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title v-if="this.data">
              <v-icon>{{ icon.mdiCardBulletedOutline }}</v-icon>
              Address {{ this.data.addresshash }}
            </v-toolbar-title>
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
              <v-divider></v-divider>
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
                        multiple
                    >
                      <template v-slot:selection="{ item }">
                        <v-chip small>
                          <span>{{ item.chip }}</span>
                        </v-chip>
                      </template>
                    </v-select>
                  </v-col>
                  <v-col v-if="this.isSortingByInput">
                    <v-alert
                        type="info"
                        text
                    >Only spent outputs are shown.
                    </v-alert>
                  </v-col>
                </v-row>
                <v-row v-if="this.isLoading">
                  <v-col v-for="i in new Array(3)" :key="i">
                    <v-skeleton-loader type="image"></v-skeleton-loader>
                  </v-col>
                </v-row>
                <v-sheet
                    v-if="!this.isLoading && !this.emptyResponse"
                    min-height="50"
                    class="fill-height"
                    color="transparent">
                  <v-lazy min-height="90" transition="fade-transition" :options="{threshold: 0.7}">
                    <v-row>
                      <v-col
                          v-for="(o,index) in this.data.addr_outputs"
                          v-bind:key="o.input_transaction + o.output_transaction + o.amount">
                        <OutputComponent :address="o" :index="index"/>
                      </v-col>
                    </v-row>
                  </v-lazy>
                </v-sheet>
                <v-row v-if="this.emptyResponse">
                  <v-col align="center" justify="center">
                    <p class="text-h6">No outputs found</p>
                  </v-col>
                </v-row>
                <v-row v-if="this.isLoadingMore">
                  <v-col>
                    <v-progress-linear
                        indeterminate
                        rounded
                        height="6"
                    ></v-progress-linear>
                  </v-col>
                </v-row>
              </v-container>
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
  mdiBankTransferOut, mdiPound,
} from '@mdi/js';
import OutputComponent from './OutputComponent.vue';
import { convertAmount, doPost, handleError } from '../../utilities';
import {
  PAGE_TITLE, ROUTE_NAME_TRANSACTION_PAGE,
  ROUTE_ADDRESS_OUTPUT_RANGE, COIN_UNIT,
} from '../../constants';
import IconItem from '../common/IconItem.vue';

export default {
  name: 'AddressLookup',
  components: { IconItem, OutputComponent },
  data() {
    return {
      icon: {
        mdiCardBulletedOutline,
        mdiScaleBalance,
        mdiBankTransferIn,
        mdiBankTransferOut,
        mdiPound,
      },
      coinUnit: COIN_UNIT,
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
      offset: 0,
      // default sort order: ascending by output timestamp
      sortOrder: 0,
      addressHash: '',
      isLoading: false,
      isLoadingMore: false,
      isSortingByInput: false,
      // emptyResponse is only used for data loaded after the initial data load
      emptyResponse: false,
      transactionRoute: ROUTE_NAME_TRANSACTION_PAGE,
    };
  },
  methods: {
    isResponseValid(data) {
      return !(!data.type || data.type !== 'addr' || !data.payload || !data.payload.addr_outputs
          || data.payload.addr_outputs.length === 0);
    },
    addNewData() {
      if (!this.data) return;

      this.offset += 20;

      // do nothing if all data is already loaded
      if (this.offset >= this.data.query_max_count) return;
      this.isLoadingMore = true;

      doPost(ROUTE_ADDRESS_OUTPUT_RANGE, this.$router,
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
      this.offset = 0;

      this.isSortingByInput = this.sortOrder === 2 || this.sortOrder === 3;

      doPost(ROUTE_ADDRESS_OUTPUT_RANGE, this.$router,
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
    handleScroll() {
      // return if not bottom of page
      if (this.isLoadingMore
          || this.loading
          || document.documentElement.scrollTop + window.innerHeight
          !== document.documentElement.offsetHeight) return;
      this.addNewData();
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
    isFilterEmpty() {
      return this.filter.selected.length === 0;
    },
  },
  mounted() {
    this.setAddressHash();
    this.updateSortState();
    this.updateFilterState();
    window.onscroll = this.handleScroll;
  },
  updated() {
    this.setAddressHash();
    this.updateSortState();
    this.updateFilterState();
  },
};
</script>
