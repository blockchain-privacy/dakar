<template>
  <fade-transition>
    <div
      v-if="isLoading"
      class="text-h6 text-center"
    >
      Searching for similar destination transactions ...
      <v-skeleton-loader type="article,article" />
    </div>
    <div v-else-if="fingerprintScores?.length > 0">
      <p class="text-subtitle-1">
        The following transactions spend outputs of CoinJoin transactions from
        <wiki-tooltip description-url="destinationFingerprinting.md">
          similar
        </wiki-tooltip> timeframes as this destination transaction.
      </p>
      <v-alert
        v-if="sessionCount !== -1 && sessionCount < 2"
        type="warning"
        variant="text"
      >
        This transaction uses outputs from only one mixing timeframe.
        The results are therefore likely not relevant.
      </v-alert>
      <v-alert
        v-if="errorMsg"
        type="error"
        variant="outlined"
      >
        {{ errorMsg }}
      </v-alert>
      <v-row>
        <v-col>
          <p
            v-if="sessionCount !== -1"
            class="text-caption"
          >
            Number of mixing timeframes: {{ sessionCount.toLocaleString() }}
          </p>
        </v-col>
        <v-col>
          <div class="d-flex align-center justify-space-between">
            <div class="text-caption text-center">
              More similar
            </div>
            <div class="gradient" />
            <div class="text-caption text-center">
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
      class="text-h6 text-center"
    >
      No similar transactions found
    </div>
  </fade-transition>
</template>

<script setup>
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import FadeTransition from '@/components/common/FadeTransition.vue';
import {onMounted, onUpdated, ref} from 'vue';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';
import {getDakarClient} from '@/utilities/index.js';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';

const {getSettings} = storeToRefs(useLocalStore());
const props = defineProps({
	transactionHash: {type: String, required: true},
});
const emit = defineEmits(['receivedTransactions']);

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

		emit('receivedTransactions', fingerprintScores.value.map(d => d.txhash));

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
