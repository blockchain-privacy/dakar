<template>
  <v-card
      class="mx-auto elevation-12"
      max-width="1000">
    <v-toolbar class="hidden-sm-and-up rounded-toolbar" color="primary" dark flat>
      <v-toolbar-title>
        <v-icon>{{ icon.mdiGraph }}</v-icon>
        Heuristics
      </v-toolbar-title>
    </v-toolbar>
    <v-toolbar dark flat class="hidden-sm-and-up" color="primary">
      <v-text-field
          v-model="search"
          :append-icon="icon.mdiMagnify"
          label="Filter items"
          single-line
          hide-details
          style="max-width: 500px"
      ></v-text-field>
      <v-spacer></v-spacer>
      <v-btn outlined @click="refreshHeuristicList" :disabled="isLoading">
        <v-icon>{{ icon.mdiRefresh }}</v-icon>
      </v-btn>
      <v-btn
          outlined
          class="ml-1"
          @click="showDeleteAllHeuristicsDialog">
        <v-icon>{{ icon.mdiDelete }}</v-icon>
        <div class="ml-2 hidden-sm-and-down">Delete All</div>
      </v-btn>
    </v-toolbar>
    <v-toolbar dark flat class="hidden-xs-only rounded-toolbar" color="primary">
      <v-toolbar-title>
        <v-icon>{{ icon.mdiGraph }}</v-icon>
        Heuristics
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-text-field
          v-model="search"
          :append-icon="icon.mdiMagnify"
          label="Filter items"
          single-line
          hide-details
          style="max-width: 500px"
      ></v-text-field>
      <v-spacer></v-spacer>
      <v-btn outlined @click="refreshHeuristicList" :disabled="isLoading">
        <v-icon>{{ icon.mdiRefresh }}</v-icon>
        <div class="ml-2 hidden-sm-and-down">Refresh</div>
      </v-btn>
      <v-btn
          outlined
          class="ml-1"
          @click="showDeleteAllHeuristicsDialog">
        <v-icon>{{ icon.mdiDelete }}</v-icon>
        <div class="ml-2 hidden-sm-and-down">Delete All</div>
      </v-btn>
    </v-toolbar>
    <v-data-table
        :headers="headers"
        :items="this.heuristicList?this.heuristicList.items:[]"
        :search="search"
        :loading="this.isLoading || !this.heuristicList"
        item-key="tx"
        sort-by="num_heuristics"
        sort-desc
        class="elevation-1">

      <template v-slot:[`item.actions`]="{ item }">
        <v-icon
            small
            :disabled="isLoading"
            class="mr-2"
            @click="goToHeuristicPage(item)">
          {{ icon.mdiOpenInNew }}
        </v-icon>
        <v-icon
            small
            :disabled="isLoading"
            @click="showDeleteDialog(item)">
          {{ icon.mdiDelete }}
        </v-icon>
      </template>
    </v-data-table>
  </v-card>
</template>

<script>

import {
  mdiGraph, mdiRefresh, mdiDelete, mdiMagnify, mdiOpenInNew,
} from '@mdi/js';
import { PAGE_TITLE, ROUTE_NAME_HEURISTIC_PAGE } from '../../constants';

export default {
  name: 'Heuristics',
  data() {
    return {
      icon: {
        mdiGraph, mdiRefresh, mdiDelete, mdiMagnify, mdiOpenInNew,
      },
      showDeleteAllDialog: false,
      showDeleteTransactionHeuristicDialog: false,
      transactionToDelete: null,
      isLoading: false,
      search: '',
      headers: [
        {
          text: 'Transaction', value: 'txhash', align: 'start', sortable: false,
        },
        {
          text: 'Number of heuristics', value: 'heuristic_count',
        },
        {
          text: 'Actions', value: 'actions', sortable: false, align: 'end',
        },
      ],
    };
  },
  computed: {
    heuristicList: {
      get() {
        return this.$store.getters.getHeuristicList;
      },
      set(value) {
        this.$store.dispatch('setHeuristicList', value);
      },
    },
  },
  methods: {
    async refreshHeuristicList() {
      this.isLoading = true;
      await this.$store.dispatch('updateHeuristicList');
      this.isLoading = false;
      this.search = '';
    },
    goToHeuristicPage(item) {
      const id = item.txhash;
      this.$router.push({ name: ROUTE_NAME_HEURISTIC_PAGE, params: { id } });
    },
    showDeleteDialog(transaction) {
      this.showDeleteTransactionHeuristicDialog = true;
      this.transactionToDelete = transaction;
    },
    showDeleteAllHeuristicsDialog() {
      this.showDeleteAllDialog = true;
    },
  },
  mounted() {
    document.title = `Heuristics - ${PAGE_TITLE}`;
    this.refreshHeuristicList();
  },
};
</script>

<style scoped>

.rounded-toolbar {
  border-top-left-radius: inherit !important;
  border-top-right-radius: inherit !important;
}

</style>
