<template>
  <v-card>
    <v-card-title class="text-h5 text-wrap">
      Similar Destination Transactions
    </v-card-title>
    <v-card-text>
      <fade-transition>
        <div
          v-if="isLoading"
          class="text-subtitle-1"
        >
          Searching for similar destination transactions ... {{ transactionHash }}
          <v-skeleton-loader type="article,article" />
        </div>
        <div v-else>
          <p class="text-subtitle-1">
            The following transactions spend outputs from similar mixing timeframe(s)
            as this transaction. Therefore, it is likely that they were created by the same user.
          </p>
          <v-alert
            v-if="sessionCount !== -1 && sessionCount < 2"
            type="warning"
            variant="text"
          >
            This transaction uses outputs from only one mixing session.
            The results are therefore likely not relevant.
          </v-alert>
          <v-alert
            v-if="errorMsg"
            type="error"
            variant="outlined"
          >
            {{ errorMsg }}
          </v-alert>
          <div v-else-if="fingerprintScores && fingerprintScores.length > 0">
            <v-row>
              <v-col>
                <p
                  v-if="sessionCount !== -1"
                  class="text-caption"
                >
                  Number of mixing sessions: {{ sessionCount.toLocaleString() }}
                </p>
              </v-col>
              <v-col>
                <div class="d-flex align-center">
                  <div class="ml-auto text-caption">
                    More similar
                  </div>
                  <div class="gradient" />
                  <div class="text-caption">
                    Less similar
                  </div>
                </div>
              </v-col>
            </v-row>
            <v-table>
              <template #default>
                <thead>
                  <tr>
                    <th />
                    <th class="text-left">
                      Transaction
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="item in fingerprintScores"
                    :key="item.txhash"
                  >
                    <td
                      :style="{background: scoreToColor(item.score),
                               width: '20px', padding: '0px 0px 0px 0px'}"
                    />
                    <td class="transaction-hash">
                      <workspace-link
                        disable-select
                        :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
                               params: { id: item.txhash, blockchainMode: getSettings.blockchainMode }}"
                      >
                        {{ item.txhash }}
                      </workspace-link>
                    </td>
                  </tr>
                </tbody>
              </template>
            </v-table>
          </div>
          <div
            v-else
            class="text-subtitle-1 text-center"
          >
            No similar transactions found
          </div>
        </div>
      </fade-transition>
    </v-card-text>
  </v-card>
</template>

<script setup>
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import FadeTransition from '@/components/common/FadeTransition.vue';
import {onMounted, onUpdated, ref} from 'vue';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';
import {getDakarClient} from '@/utilities/index.js';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';

const {getSettings} = storeToRefs(useLocalStore());
const props = defineProps({
	transactionHash: {type: String, required: true},
});

const dakar = getDakarClient(getSettings.value.blockchainMode);

const isLoading = ref(false);
const fingerprintScores = ref([]);
const sessionCount = ref(-1);
const errorMsg = ref('');

let oldTransaction = '';

// Functions
function scoreToColor(scaleNum) {
	if (scaleNum <= 0.6) {
		return '#E53935';
	}

	if (scaleNum <= 0.8) {
		return '#EF5350';
	}

	if (scaleNum <= 1.1) {
		return '#66BB6A';
	}

	return '#388E3C';
}

onUpdated(() => {
	searchForSimilarTransactions();
});

onMounted(() => searchForSimilarTransactions());

async function searchForSimilarTransactions() {
	if (props.transactionHash === '' || props.transactionHash === oldTransaction) {
		return;
	}

	oldTransaction = props.transactionHash;

	fingerprintScores.value = [];
	sessionCount.value = -1;
	errorMsg.value = '';
	isLoading.value = true;

	try {
		const response = await dakar.tools.spendingFingerprintHashGet({hash: props.transactionHash});

		if (response.fingerprint_scores) {
			fingerprintScores.value = response.fingerprint_scores
				.sort((item1, item2) => item2.score - item1.score);
		}

		if (response.session_count) {
			sessionCount.value = response.session_count;
		}
	} catch (e) {
		if (e.cause?.status === 500) {
			errorMsg.value = 'Error requesting data from server. Please try again later.';
		} else {
			errorMsg.value = e.message;
		}
	}

	isLoading.value = false;
}
</script>

<style scoped>
.transaction-hash {
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 200px;
  white-space: nowrap;
}

.gradient {
  width: 160px;
  height: 10px;
  margin: 0 5px 0 5px;
  background: linear-gradient(to left, #E53935 0%, #EF5350 33%, #66BB6A 66%, #388E3C 100%);
}
</style>
