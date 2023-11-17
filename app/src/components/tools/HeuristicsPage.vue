<template>
  <v-card
    class="mx-auto"
    variant="text"
    max-width="1200"
  >
    <icon-title
      title="Heuristics"
      :icon="icon.mdiGraph"
      :one-line="true"
    >
      <v-btn
        v-model="showSearchField"
        variant="text"
        icon
        @click="showSearchField = !showSearchField"
      >
        <v-icon>{{ icon.mdiMagnify }}</v-icon>
      </v-btn>
      <v-menu location="bottom">
        <template #activator="{ props }">
          <v-btn
            variant="text"
            icon
            v-bind="props"
          >
            <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item
            :disabled="isLoading"
            @click="refreshHeuristicList"
          >
            <template #prepend>
              <v-icon>{{ icon.mdiRefresh }}</v-icon>
            </template>
            <v-list-item-title>Refresh</v-list-item-title>
          </v-list-item>
          <v-list-item @click="showDeleteAllHeuristicsDialog">
            <template #prepend>
              <v-icon>{{ icon.mdiDelete }}</v-icon>
            </template>
            <v-list-item-title>Delete All Heuristics</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </icon-title>
    <fade-transition>
      <div
        v-if="showSearchField"
        class="d-flex align-center justify-center mb-4"
      >
        <v-text-field
          v-model="search"
          :append-inner-icon="icon.mdiMagnify"
          label="Filter items"
          single-line
          hide-details
          style="max-width:800px"
        />
      </div>
    </fade-transition>
    <v-data-table
      v-model:sort-by="sortBy"
      :search="search"
      :loading="isLoading || !heuristicList"
      :headers="headers"
      :items="heuristicList?heuristicList:[]"
    >
      <template #item.txhash="{ item }">
        <router-link
          :to="{ name: heuristicRoute,
                 params: { id: item.txhash }}"
        >
          {{ shortenHash(item.txhash) }}
        </router-link>
      </template>
      <template #item.h_count="{ item }">
        {{ item.h_count.toLocaleString() }}
      </template>
      <template #item.modTimeUnix="{ item }">
        <span>{{ new Date(item.modTimeUnix).toLocaleString() }}</span>
      </template>
      <template #[`item.actions`]="{ item }">
        <v-icon @click="showDeleteHeuristicDialog(item)">
          {{ icon.mdiDelete }}
        </v-icon>
      </template>
    </v-data-table>
    <v-dialog
      v-model="showDeleteAllDialog"
      max-width="500px"
    >
      <v-card>
        <v-card-title>
          <span class="text-h5">Delete All Heuristics</span>
        </v-card-title>
        <v-card-text>
          <p class="text-subtitle-1">
            Are you sure you want to delete all of your heuristics of all transactions?
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            variant="text"
            @click="closeDeleteAllHeuristicsDialog"
          >
            Cancel
          </v-btn>
          <v-btn
            color="red"
            variant="text"
            @click="deleteTransactionHeuristic(true)"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog
      v-if="transactionToDelete"
      v-model="showDeleteTransactionHeuristicDialog"
      max-width="500px"
    >
      <v-card>
        <v-card-title>
          <span class="text-h5">Delete heuristics</span>
        </v-card-title>
        <v-card-text>
          <p class="text-subtitle-1">
            All heuristics you defined for transaction
            {{ shortenHash(transactionToDelete.txhash) }}
            will be deleted. Continue?
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            variant="text"
            @click="closeDeleteHeuristicDialog"
          >
            Cancel
          </v-btn>
          <v-btn
            color="red"
            variant="text"
            @click="deleteTransactionHeuristic(false)"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script>

import {
	mdiGraph, mdiRefresh, mdiDelete, mdiMagnify, mdiDotsVertical,
} from '@mdi/js';
import {PAGE_TITLE, ROUTE_NAME_HEURISTIC_PAGE} from '@/constants';
import {handleError, shortenHash} from '@/utilities';
import IconTitle from '@/components/common/IconTitle.vue';
import FadeTransition from '@/components/common/FadeTransition.vue';

export default {
	name: 'HeuristicsPage',
	components: {FadeTransition, IconTitle},
	data() {
		return {
			icon: {
				mdiGraph, mdiRefresh, mdiDelete, mdiMagnify, mdiDotsVertical,
			},
			heuristicList: null,
			heuristicRoute: ROUTE_NAME_HEURISTIC_PAGE,
			showDeleteAllDialog: false,
			showDeleteTransactionHeuristicDialog: false,
			transactionToDelete: null,
			isLoading: false,
			showSearchField: false,
			search: '',
			sortBy: [{key: 'modTimeUnix', order: 'desc'}],
			headers: [
				{
					title: 'Transaction', key: 'txhash', align: 'start', sortable: false,
				},
				{
					title: 'Number of heuristics', key: 'h_count',
				},
				{
					title: 'Last modification', key: 'modTimeUnix',
				},
				{
					title: '', key: 'actions', sortable: false, align: 'end',
				},
			],
		};
	},
	mounted() {
		document.title = `Heuristics - ${PAGE_TITLE}`;
		this.refreshHeuristicList();
	},
	methods: {
		shortenHash,
		setErrorMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'error', temporary: false, category: this.$route.name});
		},
		setInfoMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'info', temporary: true, category: this.$route.name});
		},
		async loadHeuristicList() {
			try {
				const response = await this.dakar.heuristic.heuristicListGet();

				if (!response.items) {
					throw new Error('received malformed response');
				}

				this.heuristicList = response.items;
				this.$store.dispatch('resetMessages');
			} catch (e) {
				handleError(this, e);
			}
		},
		async refreshHeuristicList() {
			this.isLoading = true;
			await this.loadHeuristicList();
			this.isLoading = false;
			this.search = '';

			if (!this.heuristicList) {
				return;
			}

			this.heuristicList = this.heuristicList.map(d => {
				// Convert date to unix time, so it can be sorted in data table
				d.modTimeUnix = new Date(d.mod_time).getTime();
				return d;
			});
		},
		async deleteTransactionHeuristic(all) {
			this.isLoading = true;
			let arg = null;
			if (all) {
				// eslint-disable-next-line camelcase
				arg = {delete_all: true};
			} else {
				// eslint-disable-next-line camelcase
				arg = {tx_hash: this.transactionToDelete.txhash};
			}

			try {
				const response = await this.dakar.heuristic.deleteHeuristicPost({heuristic: arg});
				if (response.msg) {
					this.setInfoMessage(response.msg);
				}

				await this.refreshHeuristicList();
			} catch (e) {
				this.setErrorMessage(e);
			}

			this.isLoading = false;
			this.showDeleteTransactionHeuristicDialog = false;
			this.showDeleteAllDialog = false;
		},
		showDeleteHeuristicDialog(transaction) {
			if (this.isLoading) {
				return;
			}

			this.showDeleteTransactionHeuristicDialog = true;
			this.transactionToDelete = transaction;
		},
		closeDeleteHeuristicDialog() {
			this.showDeleteTransactionHeuristicDialog = false;
		},

		showDeleteAllHeuristicsDialog() {
			this.showDeleteAllDialog = true;
		},
		closeDeleteAllHeuristicsDialog() {
			this.showDeleteAllDialog = false;
		},
	},
};
</script>

<style scoped>

</style>
