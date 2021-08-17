<template>
  <v-card
      class="mx-auto elevation-4"
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
        sort-by="mod_time"
        sort-desc
        class="elevation-1">

      <template v-slot:[`item.txhash`]="{ item }">
        <router-link :to="{ name: transactionRoute,
                    params: { id: item.txhash }}">
          {{ shortenHash(item.txhash) }}
        </router-link>
      </template>
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
            @click="showDeleteHeuristicDialog(item)">
          {{ icon.mdiDelete }}
        </v-icon>
      </template>
      <template v-slot:[`item.mod_time`]="{ item }">
        <span>{{ new Date(item.mod_time).toLocaleString() }}</span>
      </template>
    </v-data-table>
    <v-dialog
        v-model="showDeleteAllDialog"
        max-width="500px">
      <v-card>
        <v-card-title>
          <span class="text-h5">Delete all heuristics</span>
        </v-card-title>
        <v-card-text>
          <p class="font-weight-black text-body-1 my-0">
            All your heuristics of all transactions will be deleted. Continue?
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text @click="closeDeleteAllHeuristicsDialog">Cancel</v-btn>
          <v-btn
              color="blue darken-1"
              text
              @click="deleteAllHeuristics">Delete all heuristics
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog
        v-model="showDeleteTransactionHeuristicDialog"
        max-width="500px"
        v-if="this.transactionToDelete">
      <v-card>
        <v-card-title>
          <span class="text-h5">Delete all transaction heuristics</span>
        </v-card-title>
        <v-card-text>
          <p class="font-weight-black text-body-1 my-0">
            All your heuristics of transaction
            {{ shortenHash(this.transactionToDelete.txhash) }}
            will be deleted. Continue?
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text @click="closeDeleteHeuristicDialog">Cancel</v-btn>
          <v-btn
              color="blue darken-1"
              text
              @click="deleteTransactionHeuristic">Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script>

import {
  mdiGraph, mdiRefresh, mdiDelete, mdiMagnify, mdiOpenInNew,
} from '@mdi/js';
import {
  PAGE_TITLE, ROUTE_NAME_HEURISTIC_PAGE, ROUTE_NAME_TRANSACTION_PAGE, ROUTE_DELETE_HEURISTIC,
  ROUTE_HEURISTIC_LIST,
} from '../../constants';
import {
  doGet, doPost, handleError, shortenHash,
} from '../../utilities';

export default {
  name: 'Heuristics',
  data() {
    return {
      icon: {
        mdiGraph, mdiRefresh, mdiDelete, mdiMagnify, mdiOpenInNew,
      },
      heuristicList: null,
      transactionRoute: ROUTE_NAME_TRANSACTION_PAGE,
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
          text: 'Number of heuristics', value: 'h_count',
        },
        {
          text: 'Last modification', value: 'mod_time',
        },
        {
          text: 'Actions', value: 'actions', sortable: false, align: 'end',
        },
      ],
    };
  },
  methods: {
    shortenHash,
    setErrorMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'error', temporary: false });
    },
    setInfoMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'info', temporary: true });
    },
    loadHeuristicList() {
      return doGet(ROUTE_HEURISTIC_LIST, this.$router, this.$store).then((data) => {
        this.heuristicList = data;
        this.$store.dispatch('resetMessages');
      }).catch((e) => {
        handleError(this.$store, e);
        return e;
      });
    },
    async refreshHeuristicList() {
      this.isLoading = true;
      await this.loadHeuristicList();
      this.isLoading = false;
      this.search = '';

      if (!this.heuristicList) return;

      this.heuristicList.items = this.heuristicList.items.map((d) => {
        // convert date to unix time so it can be sorted in data table
        d.mod_time = new Date(d.mod_time).getTime();
        return d;
      });
    },
    deleteHeuristics(body) {
      return doPost(ROUTE_DELETE_HEURISTIC, this.$router, this.$store, body)
        .then((data) => {
          if (data.success === undefined) throw Error('error deleting heuristics');
          if (data.success === false) {
            throw Error(data.msg);
          }

          if (data.msg) {
            this.setInfoMessage(data.msg);
          }

          this.refreshHeuristicList();
        })
        .catch((error) => {
          this.setErrorMessage(error);
        });
    },
    goToHeuristicPage(item) {
      const id = item.txhash;
      this.$router.push({ name: ROUTE_NAME_HEURISTIC_PAGE, params: { id } });
    },
    deleteTransactionHeuristic() {
      this.isLoading = true;
      this.deleteHeuristics({ tx_hash: this.transactionToDelete.txhash })
        .finally(() => {
          this.isLoading = false;
          this.showDeleteTransactionHeuristicDialog = false;
        });
    },
    showDeleteHeuristicDialog(transaction) {
      this.showDeleteTransactionHeuristicDialog = true;
      this.transactionToDelete = transaction;
    },
    closeDeleteHeuristicDialog() {
      this.showDeleteTransactionHeuristicDialog = false;
    },
    deleteAllHeuristics() {
      this.isLoading = true;
      this.deleteHeuristics({ delete_all: true })
        .finally(() => {
          this.isLoading = false;
          this.showDeleteAllDialog = false;
        });
    },
    showDeleteAllHeuristicsDialog() {
      this.showDeleteAllDialog = true;
    },
    closeDeleteAllHeuristicsDialog() {
      this.showDeleteAllDialog = false;
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
