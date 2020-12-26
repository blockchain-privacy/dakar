<template>
  <v-container class="fill-height" fluid v-if="this.data">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <v-card class="elevation-12">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title v-if="this.data">
              <v-icon>mdi-card-bulleted-outline</v-icon>
              Address {{ this.data.addresshash }}
            </v-toolbar-title>
          </v-toolbar>
          <v-card-text>
            <v-container>
              <v-row>
                <v-col>
                  <IconItem icon="mdi-scale-balance" title="Balance">
                    {{ convertAmount(this.data.output_sum - this.data.input_sum) }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-row>
                <v-col>
                  <IconItem icon="mdi-bank-transfer-in" title="Total amount received">
                    {{ convertAmount(this.data.output_sum) }}
                  </IconItem>
                </v-col>
                <v-col>
                  <IconItem icon="mdi-bank-transfer-out" title="Total amount spent">
                    {{ convertAmount(this.data.input_sum) }}
                  </IconItem>
                </v-col>
              </v-row>
              <v-divider></v-divider>
              <v-row v-if="this.data.addr_outputs.length > 1">
                <v-col>
                  <v-select
                      :disabled="this.isLoading"
                      :loading="this.isLoading?'primary':false"
                      style="max-width: 300px;"
                      v-model="combobox.selected.id"
                      :items="combobox.items"
                      item-value="id"
                      item-text="text"
                      label="Sort by"
                      v-on:change="handleSort"
                  ></v-select>
                </v-col>
              </v-row>
              <v-row v-if="this.isLoading">
                <v-col v-for="i in [1,2,3]" :key="i">
                  <v-skeleton-loader type="image"></v-skeleton-loader>
                </v-col>
              </v-row>
              <v-sheet
                  v-if="!this.isLoading"
                  min-height="50"
                  class="fill-height"
                  color="transparent">
                <v-lazy min-height="90" transition="fade-transition" :options="{threshold: 0.7}">
                  <v-row>
                    <v-col v-for="o in this.data.addr_outputs"
                           v-bind:key="o.input_transaction + o.output_transaction + o.amount">
                      <OutputComponent :address="o"/>
                    </v-col>
                  </v-row>
                </v-lazy>
              </v-sheet>
              <v-progress-linear
                  v-if="this.isLoadingMore"
                  indeterminate
                  rounded
                  height="6"
              ></v-progress-linear>
            </v-container>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import OutputComponent from './OutputComponent.vue';
import { convertAmount, doPost, handleError } from '../utilities';
import { PAGE_TITLE, ROUTE_NAME_TRANSACTION_PAGE, ROUTE_ADDRESS_OUTPUT_RANGE } from '../constants';
import IconItem from './common/IconItem.vue';

export default {
  name: 'AddressLookup',
  components: { IconItem, OutputComponent },
  methods: {
    addNewData() {
      if (!this.data) return;

      this.offset += 20;

      // do nothing if all data is already loaded
      if (this.offset >= this.data.num_outputs) return;
      this.isLoadingMore = true;

      doPost(ROUTE_ADDRESS_OUTPUT_RANGE, this.addressHash,
        { offset: this.offset, order: this.sortOrder })
        .then((data) => {
          this.data.addr_outputs = [...this.data.addr_outputs, ...data.payload.addr_outputs];
          this.$store.dispatch('resetMsg');
        })
        .catch((e) => {
          handleError(this.$store, e);
        })
        .finally(() => {
          this.isLoadingMore = false;
        });
    },
    handleSort() {
      this.sortOrder = this.combobox.selected.id;
      this.offset = 0;
      this.isLoading = true;
      doPost(ROUTE_ADDRESS_OUTPUT_RANGE, this.addressHash,
        { offset: this.offset, order: this.sortOrder })
        .then((data) => {
          this.data = data.payload;
          this.$store.dispatch('resetMsg');
        })
        .catch((e) => {
          handleError(this.$store, e);
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
    convertAmount,
    calculateAmountReceived(outputs) {
      if (outputs === undefined) return 0;
      return outputs
        .map((e) => parseInt(e.amount, 10))
        .reduce((sum, e) => sum + e, 0);
    },
    calculateAmountSpent(outputs) {
      if (outputs === undefined) return 0;
      return outputs
        .filter((e) => e.input_transaction !== '')
        .map((e) => parseInt(e.amount, 10))
        .reduce((sum, e) => sum + e, 0);
    },
    handleScroll() {
      // return if not bottom of page
      if (this.isLoadingMore || document.documentElement.scrollTop + window.innerHeight
          !== document.documentElement.offsetHeight) return;
      this.addNewData();
    },
    setAddressHash() {
      let h = ' ';
      if (this.data && this.data.addresshash && this.data.addresshash !== this.addressHash) {
        this.addressHash = this.data.addresshash;
        h = ` ${this.addressHash} `;
      }
      document.title = `Address${h}- ${PAGE_TITLE}`;
    },
  },
  data() {
    return {
      combobox: {
        selected: {
          id: 2,
        },
        items: [
          { id: 0, text: 'ascending by input date' },
          { id: 1, text: 'descending by input date' },
          { id: 2, text: 'ascending by output date' },
          { id: 3, text: 'descending by output date' },
        ],
      },
      offset: 0,
      // default sort order: ascending by output timestamp
      sortOrder: 2,
      addressHash: '',
      isLoading: false,
      isLoadingMore: false,
      transactionRoute: ROUTE_NAME_TRANSACTION_PAGE,
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
  },
  mounted() {
    this.setAddressHash();
    window.onscroll = this.handleScroll;
  },
  updated() {
    this.setAddressHash();
  },
};
</script>
