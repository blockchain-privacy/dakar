<template>
  <v-card
    variant="text"
    class="mx-auto"
    max-width="1200"
  >
    <IconTitle
      title="Shortest Path"
      :icon="icon.mdiChartTimelineVariant"
    />
    <v-card-text>
      <div class="text-subtitle-1 mb-5">
        Find one of the shortest paths between two transactions. Multiple shortest paths can exist.
      </div>
      <div
        class="d-flex align-center flex-wrap"
        style="gap: 5px 20px"
      >
        <v-text-field
          v-model="fromTransaction"
          style="min-width: 200px"
          label="From"
          :disabled="isLoading"
          :autofocus="true"
        />
        <v-text-field
          v-model="toTransaction"
          style="min-width: 200px"
          label="To"
          :disabled="isLoading"
        />
      </div>
      <div class="d-flex align-center flex-wrap">
        <v-radio-group
          v-model="anyDirection"
          :inline="true"
          label="Search direction:"
          :disabled="isLoading"
        >
          <v-radio
            label="Linear"
            :value="false"
          />
          <v-radio
            label="Any"
            :value="true"
          />
        </v-radio-group>
        <v-switch
          v-model="includePrivacyTransactions"
          label="Traverse private transactions"
          class="mx-5"
          :disabled="isLoading"
        />
        <v-btn
          class="ms-auto"
          color="primary"
          :disabled="!isSearchable"
          :loading="isLoading"
          @click="handleSearch"
        >
          Search
        </v-btn>
      </div>
      <v-divider
        v-if="transactions.length > 0"
        class="my-3"
      />
      <v-timeline
        v-if="transactions.length > 0"
        :density="$vuetify.display.smAndDown?'compact':undefined"
        side="end"
      >
        <v-timeline-item
          v-for="(tx) in transactions"
          :key="tx.txhash"
          :dot-color="tx.privacytype>=0?'purple':'primary'"
          max-width="500px"
        >
          <template #opposite>
            <span
              class="text-h5"
              v-text="new Date(tx.bts).toLocaleString()"
            />
          </template>
          <transaction-item :tx="tx" />
        </v-timeline-item>
      </v-timeline>
    </v-card-text>
  </v-card>
</template>

<script>
import {mdiChartTimelineVariant} from '@mdi/js';
import {handleError} from '@/utilities';
import {PAGE_TITLE} from '@/constants';
import IconTitle from '@/components/common/IconTitle.vue';
import TransactionItem from '@/components/common/TransactionItem.vue';

export default {
	name: 'ShortestPathPage',
	components: {TransactionItem, IconTitle},
	data() {
		return {
			icon: {mdiChartTimelineVariant},
			// V-model
			fromTransaction: '',
			toTransaction: '',
			includePrivacyTransactions: true,
			anyDirection: false,
			isLoading: false,
			transactions: [],
		};
	},
	computed: {
		isSearchable() {
			return this.toTransaction && this.fromTransaction
          && this.toTransaction.trim().length > 0 && this.fromTransaction.trim().length > 0
          && this.toTransaction !== this.fromTransaction;
		},
	},
	mounted() {
		document.title = `Shortest Path - ${PAGE_TITLE}`;
	},
	methods: {
		setInfoMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'info', temporary: true, category: this.$route.name});
		},
		async handleSearch() {
			if (this.isLoading || !this.isSearchable) {
				return;
			}

			this.$store.dispatch('resetMessages');

			this.transactions = [];
			await this.doLookup();
		},
		async doLookup() {
			this.isLoading = true;

			try {
				const response = await this.dakar.tools.shortestTransactionPathPost({transactions: {
					to: this.fromTransaction.trim(),
					from: this.toTransaction.trim(),
					includePrivacyTransactions: this.includePrivacyTransactions,
					anyDirection: this.anyDirection,
				}});

				if (response.msg) {
					this.setInfoMessage(response.msg);
				}

				if (response.transactions && response.transactions.length > 0) {
					if (this.fromTransaction.trim() !== response.transactions[0].txhash) {
						response.transactions = response.transactions.reverse();
					}

					this.transactions = response.transactions;
				}
			} catch (e) {
				handleError(this, e);
			}

			this.isLoading = false;
		},
	},
};
</script>

<style scoped>

</style>
