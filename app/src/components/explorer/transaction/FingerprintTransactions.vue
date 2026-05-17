<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <fade-transition>
    <div
      v-if="isLoading"
      class="text-title-large text-center"
    >
      Searching for similar destination transactions ...
      <v-skeleton-loader type="article,article" />
    </div>
    <div v-else-if="fingerprintScores?.length > 0">
      <p class="text-body-large">
        The following transactions spend outputs of CoinJoin transactions from
        <wiki-tooltip description-url="destinationFingerprinting.md">
          similar
        </wiki-tooltip> time frames as this destination transaction.
      </p>
      <v-alert
        v-if="sessionCount !== -1 && sessionCount < 2"
        type="warning"
        variant="text"
      >
        This transaction uses outputs from only one mixing time frame.
        The results are therefore likely not relevant.
      </v-alert>
      <v-alert
        v-if="errorMsg"
        type="error"
        variant="outlined"
      >
        {{ errorMsg }}
      </v-alert>
      <p
        v-if="sessionCount !== -1"
        class="text-body-small"
      >
        Number of mixing time frames: {{ sessionCount.toLocaleString() }}
      </p>
      <v-table
        class="mt-2"
        density="compact"
      >
        <thead>
          <tr>
            <th>
              Transaction
            </th>
            <th style="max-width: 0px">
              Avg. Min. Distance (h)
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in fingerprintScores"
            :key="item.txhash"
          >
            <td class="transaction-hash">
              <workspace-link
                :to="{ name: ROUTE_NAME_TRANSACTION_PAGE,
                       params: { id: item.txhash, blockchainMode: route.params.blockchainMode }}"
              >
                {{ item.txhash }}
              </workspace-link>
            </td>
            <td>
              {{ (item.score / 3600).toFixed(2) }}
            </td>
          </tr>
        </tbody>
      </v-table>
    </div>
    <div
      v-else
      class="text-title-large text-center"
    >
      No similar transactions found
    </div>
  </fade-transition>
</template>

<script setup>
import {onMounted, onUpdated, ref} from 'vue';
import {useRoute} from 'vue-router';
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import FadeTransition from '@/components/common/FadeTransition.vue';
import {getDakarClient} from '@/utilities/index.js';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';

const props = defineProps({
	transactionHash: {type: String, required: true},
});
const emit = defineEmits(['receivedTransactions']);
const route = useRoute();

const dakar = getDakarClient(route.params.blockchainMode);

const isLoading = ref(false);
const fingerprintScores = ref([]);
const sessionCount = ref(-1);
const errorMsg = ref('');

let oldTransaction = '';

// Hooks
onUpdated(() => {
	searchForSimilarTransactions();
});

onMounted(() => searchForSimilarTransactions());

// Functions
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
				.toSorted((item1, item2) => item1.score - item2.score);
		}

		emit('receivedTransactions', fingerprintScores.value.map(d => d.txhash));

		if (response.session_count) {
			sessionCount.value = response.session_count;
		}
	} catch (error) {
		errorMsg.value = error.cause?.status === 500 ? 'Error requesting data from server. Please try again later.' : error.message;
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
</style>
