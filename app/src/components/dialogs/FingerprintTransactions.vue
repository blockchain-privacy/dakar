<template>
  <v-dialog v-model="show" max-width="700px">
    <v-card class="mx-auto elevation-4">
      <v-card-title>
        <span class="text-h5">Similar Destination Transactions</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1" v-if="isLoading">
          Search for similar destination transactions ... {{ transactionHash }}
          <v-skeleton-loader type="table"></v-skeleton-loader>
        </div>
        <div v-if="!isLoading">
          <p class="text-subtitle-1">
            The following transactions spend outputs from the same mixing timeframe(s)
            as this transaction. Therefore, it is likely that they were created by the same user.
            The larger the score the closer the timeframe(s) are.
          </p>
          <v-alert v-if="errorMsg" type="error" outlined>{{ errorMsg }}</v-alert>
          <v-simple-table v-else-if="fingerprintScores && fingerprintScores.length > 0">
            <template v-slot:default>
              <thead>
              <tr>
                <th class="text-left">Transaction</th>
                <th class="text-left">Score</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="item in fingerprintScores" :key="item.txhash">
                <td class="transaction-hash">
                  <router-link :to="{ name: routes.transactionRoute, params: { id: item.txhash }}">
                    {{ item.txhash }}
                  </router-link>
                </td>
                <td>{{ item.score }}</td>
              </tr>
              </tbody>
            </template>
          </v-simple-table>
          <div v-else class="text-body-1">
            No similar transactions found
          </div>
        </div>
        <v-row class="mt-4">
          <v-col class="d-flex justify-end align-center">
            <v-btn text class="mr-2" @click="show = false">Back</v-btn>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
import { doGet } from '../../utilities';
import { ROUTE_SPENDING_FINGERPRINT, ROUTE_NAME_TRANSACTION_PAGE } from '../../constants';

export default {
  name: 'FingerprintTransactions',
  props: {
    value: { type: Boolean, required: true },
    transactionHash: { type: String, required: true },
  },
  data() {
    return {
      isLoading: false,
      fingerprintScores: [],
      // loadedSuccessful controls if a data load request needs to be sent
      loadedSuccessful: false,
      errorMsg: '',
      routes: {
        transactionRoute: ROUTE_NAME_TRANSACTION_PAGE,
      },
    };
  },
  computed: {
    show: {
      get() {
        return this.value;
      },
      set(value) {
        this.$emit('input', value);
      },
    },
  },
  methods: {
    searchForSimilarTransactions() {
      // check if data was already loaded
      if (this.loadedSuccessful) return;

      this.fingerprintScores = [];
      this.errorMsg = '';
      this.isLoading = true;

      doGet(ROUTE_SPENDING_FINGERPRINT, this.$router, this.$store, this.transactionHash)
        .then((d) => {
          if (d.success === undefined || (!d.success && d.msg === undefined)) throw new Error('Error searching for similar transactions.');
          if (!d.success && d.msg !== undefined) throw new Error(d.msg);
          if (d.success) this.loadedSuccessful = true;
          if (d.fingerprint_scores) {
            this.fingerprintScores = d.fingerprint_scores
              .sort((item1, item2) => item1.score < item2.score)
              .map((item) => {
                item.score = item.score.toFixed(3);
                return item;
              });
          }
        })
        .catch((e) => {
          this.errorMsg = e.message;
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
  },
  watch: {
    value(newVal) {
      if (!newVal) return;
      this.searchForSimilarTransactions();
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
</style>
