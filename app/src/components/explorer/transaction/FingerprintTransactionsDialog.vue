<template>
  <v-dialog
    v-model="show"
    max-width="700px"
  >
    <v-card class="pb-2">
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
              :icon="icons.mdiTestTube"
              type="info"
              variant="text"
            >
              This feature is under active development. Results may change.
            </v-alert>
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
                        <router-link
                          :to="{ name: routes.transactionRoute, params: { id: item.txhash }}"
                        >
                          {{ item.txhash }}
                        </router-link>
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
        <v-row class="mt-4">
          <v-col class="d-flex justify-end align-center">
            <v-btn
              variant="text"
              class="mr-2"
              @click="show = false"
            >
              Back
            </v-btn>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
import {mdiTestTube} from '@mdi/js';
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import FadeTransition from '@/components/common/FadeTransition.vue';
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

export default {
	name: 'FingerprintTransactionsDialog',
	components: {FadeTransition},
	props: {
		modelValue: {type: Boolean, required: true},
		transactionHash: {type: String, required: true},
	},
	emits: ['update:modelValue'],
	data() {
		return {
			isLoading: false,
			fingerprintScores: [],
			sessionCount: -1,
			// LoadedSuccessful controls if a data load request needs to be sent
			loadedSuccessful: false,
			errorMsg: '',
			icons: {mdiTestTube},
			routes: {
				transactionRoute: ROUTE_NAME_TRANSACTION_PAGE,
			},
		};
	},
	computed: {
		show: {
			get() {
				return this.modelValue;
			},
			set(value) {
				this.$emit('update:modelValue', value);
			},
		},
	},
	watch: {
		modelValue(newVal) {
			if (!newVal) {
				return;
			}

			this.searchForSimilarTransactions();
		},
	},
	methods: {
		scoreToColor,
		async searchForSimilarTransactions() {
			// Check if data was already loaded
			if (this.loadedSuccessful) {
				return;
			}

			this.fingerprintScores = [];
			this.sessionCount = -1;
			this.errorMsg = '';
			this.isLoading = true;

			try {
				const response = await this.dakar.tools.spendingFingerprintHashGet({hash: this.transactionHash});

				this.loadedSuccessful = true;

				if (response.fingerprint_scores) {
					this.fingerprintScores = response.fingerprint_scores
						.sort((item1, item2) => item2.score - item1.score);
				}

				if (response.session_count) {
					this.sessionCount = response.session_count;
				}
			} catch (e) {
				if (e.cause?.status === 500) {
					this.errorMsg = 'Error requesting data from server. Please try again later.';
				} else {
					this.errorMsg = e.message;
				}
			}

			this.isLoading = false;
		},
	},
};
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
