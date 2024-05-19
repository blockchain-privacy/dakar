<template>
  <v-card
    variant="text"
    class="mx-auto"
    max-width="1200"
  >
    <icon-title
      title="Shortest Path"
      :icon="mdiChartTimelineVariant"
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
          autofocus
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
          inline
          label="Search direction:"
          :disabled="isLoading"
        >
          <v-radio
            label="Linear"
            :value="false"
          />
          <v-radio
            label="Any"
            value
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

<script setup>
import {mdiChartTimelineVariant} from '@mdi/js';
import {handleError} from '@/utilities';
import {PAGE_TITLE} from '@/constants';
import IconTitle from '@/components/common/IconTitle.vue';
import TransactionItem from '@/components/common/TransactionItem.vue';
import {
	computed, inject, onMounted, ref,
} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();
const context = {addMessage: msgStore.addMessage, $route: route};

// V-model
const fromTransaction = ref('');
const toTransaction = ref('');
const includePrivacyTransactions = ref(true);
const anyDirection = ref(false);
const isLoading = ref(false);
const transactions = ref([]);

// Computed
const isSearchable = computed(() => toTransaction.value && fromTransaction
    && toTransaction.value.trim().length > 0 && fromTransaction.value.trim().length > 0
    && toTransaction.value !== fromTransaction.value);

// Hooks
onMounted(() => {
	document.title = `Shortest Path - ${PAGE_TITLE}`;
});

// Functions
function setInfoMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'info', temporary: true, category: route.name,
	});
}

async function handleSearch() {
	if (isLoading.value || !isSearchable.value) {
		return;
	}

	msgStore.resetMessages();

	transactions.value = [];
	await doLookup();
}

async function doLookup() {
	isLoading.value = true;

	try {
		const response = await dakar.tools.shortestTransactionPathPost({
			transactions: {
				to: fromTransaction.value.trim(),
				from: toTransaction.value.trim(),
				includePrivacyTransactions: includePrivacyTransactions.value,
				anyDirection: anyDirection.value,
			},
		});

		if (response.msg) {
			setInfoMessage(response.msg);
		}

		if (response.transactions && response.transactions.length > 0) {
			if (fromTransaction.value.trim() !== response.transactions[0].txhash) {
				response.transactions = response.transactions.reverse();
			}

			transactions.value = response.transactions;
		}
	} catch (e) {
		handleError(context, e);
	}

	isLoading.value = false;
}

</script>

<style scoped>

</style>
