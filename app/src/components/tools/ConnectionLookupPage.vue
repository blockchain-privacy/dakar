<template>
  <div>
    <v-card
      variant="text"
      class="mx-auto"
      max-width="1200"
    >
      <icon-title
        title="Connection Lookup"
        :icon="mdiTextBoxSearch"
      />
      <v-card-text>
        <p class="text-subtitle-1 mb-5">
          Find transactions connected to a privacy transaction. The returned transactions are connected to the
          given transaction via mixing transactions.
        </p>
        <v-row>
          <v-col>
            <v-text-field
              v-model="fromTransaction"
              label="Start transaction"
              :disabled="isLoading"
              :autofocus="true"
              @keydown.enter="handleSearch"
            />
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-slider
              v-model="maxLookBackTime"
              class="mt-5 customSlider"
              label="Time limit:"
              max="90"
              min="1"
              :step="1"
              thumb-label="always"
              thumb-size="20"
              hide-details
            >
              <template #append>
                {{ maxLookBackTime === 1 ? 'day' : 'days' }}
                <div class="ml-2">
                  <v-hover
                    v-slot="{ hover }"
                    open-delay="0"
                  >
                    <v-icon
                      id="max_time_tooltip"
                      size="small"
                    >
                      {{ hover ? mdiInformation : mdiInformationOutline }}
                    </v-icon>
                  </v-hover>
                  <v-tooltip
                    location="right"
                    activator="#max_time_tooltip"
                  >
                    <span>Maximum time to look forward or backward.</span>
                  </v-tooltip>
                </div>
              </template>
            </v-slider>
          </v-col>
        </v-row>
        <v-row>
          <v-col
            cols="12"
            sm="8"
          >
            <v-radio-group
              v-model="isDirectionForward"
              mandatory
              label="Search direction:"
              :disabled="isLoading"
            >
              <div class="d-flex align-center">
                <div class="mr-2">
                  <v-radio
                    label="Backward"
                    :value="false"
                  />
                </div>
                <v-hover
                  v-slot="{ hover }"
                  open-delay="0"
                >
                  <v-icon
                    id="reverse_lookup_tooltip"
                    size="small"
                  >
                    {{ hover ? mdiInformation : mdiInformationOutline }}
                  </v-icon>
                </v-hover>
                <v-tooltip
                  location="right"
                  activator="#reverse_lookup_tooltip"
                >
                  <span>Starting with the given transaction, all mixing transactions
                    connected via inputs will be traversed.</span>
                </v-tooltip>
                <div class="mr-2">
                  <v-radio
                    label="Forward"
                    :value="true"
                  />
                </div>
                <div>
                  <v-hover
                    v-slot="{ hover }"
                    open-delay="0"
                  >
                    <v-icon
                      id="forward_lookup_tooltip"
                      size="small"
                    >
                      {{ hover ? mdiInformation : mdiInformationOutline }}
                    </v-icon>
                  </v-hover>
                  <v-tooltip
                    location="right"
                    activator="#forward_lookup_tooltip"
                  >
                    <span>Starting with the given transaction, all mixing transactions
                      connected via outputs will be traversed.</span>
                  </v-tooltip>
                </div>
              </div>
            </v-radio-group>
          </v-col>
          <v-col
            class="d-flex justify-end align-center"
            cols="12"
            sm="4"
          >
            <v-btn
              color="primary"
              :disabled="!isSearchable"
              :loading="isLoading"
              @click="handleSearch"
            >
              Search
            </v-btn>
          </v-col>
        </v-row>
        <v-divider
          v-if="transactions.length > 0"
          class="my-3"
        />
        <p v-if="transactionCount >=0">
          Found {{ transactionCount.toLocaleString() }}
          {{ plural('transaction', transactionCount) }}
        </p>
        <p v-if="transactions.length > 30">
          Result list is limited to 30 transactions.
        </p>
      </v-card-text>
    </v-card>
    <div v-if="transactions.length > 0">
      <transaction-item
        v-for="(tx) in transactions"
        :key="tx.txhash"
        :tx="tx"
        class="mx-auto mt-3"
        max-width="1000"
      />
    </div>
  </div>
</template>

<script setup>
import {
	mdiTextBoxSearch, mdiInformation, mdiInformationOutline,
} from '@mdi/js';
import {handleError, plural} from '@/utilities';
import {PAGE_TITLE} from '@/constants';
import IconTitle from '@/components/common/IconTitle.vue';
import TransactionItem from '@/components/common/TransactionItem.vue';
import {computed, inject, onMounted, ref} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();
const context = {addMessage: msgStore.addMessage, $route: route};

const fromTransaction = ref('');
const isDirectionForward = ref(false);
const isLoading = ref(false);
const transactions = ref([]);
const transactionCount = ref(-1);
const maxLookBackTime = ref(5);

// Computed
const isSearchable = computed(() => fromTransaction.value && fromTransaction.value.trim().length > 0);

// Hooks
onMounted(() => {
	document.title = `Connection Lookup - ${PAGE_TITLE}`;
});

// Functions
function setInfoMessage(msg) {
	msgStore.addMessage({text: msg, type: 'info', temporary: true, category: route.name});
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
	transactionCount.value = -1;
	isLoading.value = true;

	try {
		const response = await dakar.tools.connectionLookupTxHashGet({
			txHash: fromTransaction.value.trim(),
			forward: isDirectionForward.value,
			t: maxLookBackTime.value,
		});

		if (response.msg) {
			setInfoMessage(response.msg);
		}

		if (response.transactions && response.transactions.length > 0) {
			if (response.count) {
				transactionCount.value = response.count;
			} else {
				transactionCount.value = -1;
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

.customSlider {
  min-width: 250px;
  max-width: 400px;
}

:deep( .customSlider .v-label ) {
  font-size: 14px;
}

</style>
