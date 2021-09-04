<template>
  <div>
    <v-card
        class="mx-auto elevation-4"
        max-width="1000">
      <v-toolbar color="primary" dark flat>
        <v-toolbar-title>
          <v-icon>{{ icon.mdiTextBoxSearch }}</v-icon>
          Connection Lookup
        </v-toolbar-title>
      </v-toolbar>
      <v-card-text>
        <div class="text-subtitle-1">
          Find transactions connected to a privacy transaction.
          <v-hover v-slot:default="{ hover }" open-delay="0">
            <v-icon small id="connection_lookup_tooltip">
              {{ hover ? icon.mdiInformation : icon.mdiInformationOutline }}
            </v-icon>
          </v-hover>
          <v-tooltip right activator="#connection_lookup_tooltip">
          <span>The returned transactions are connected to the
            given transaction via mixing transactions.</span>
          </v-tooltip>
        </div>
        <v-row>
          <v-col>
            <v-text-field label="Start transaction"
                          v-model="fromTransaction"
                          :disabled="isLoading"
                          @keydown.enter="handleSearch"
                          autofocus/>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-slider
                class="mt-5 customSlider"
                label="Time limit:"
                v-model="maxLookBackTime"
                max="90"
                min="1"
                thumb-label="always"
                thumb-size="35"
                hide-details>
              <template v-slot:append>
                {{ maxLookBackTime === 1 ? 'day' : 'days' }}
                <div class="ml-2">
                  <v-hover v-slot:default="{ hover }" open-delay="0">
                    <v-icon small id="max_time_tooltip">
                      {{ hover ? icon.mdiInformation : icon.mdiInformationOutline }}
                    </v-icon>
                  </v-hover>
                  <v-tooltip right activator="#max_time_tooltip">
                    <span>Maximum time to look forward or backward.</span>
                  </v-tooltip>
                </div>
              </template>
            </v-slider>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-radio-group
                v-model="isDirectionForward"
                mandatory
                row
                label="Search direction:"
                :disabled="isLoading">
              <v-radio class="mr-2" label="Backward" :value="false"/>
              <div class="mr-3">
                <v-hover v-slot:default="{ hover }" open-delay="0">
                  <v-icon small id="reverse_lookup_tooltip">
                    {{ hover ? icon.mdiInformation : icon.mdiInformationOutline }}
                  </v-icon>
                </v-hover>
                <v-tooltip right activator="#reverse_lookup_tooltip">
                <span>Starting with the given transaction, all mixing transactions
                  connected via inputs will be traversed.</span>
                </v-tooltip>
              </div>
              <v-radio class="mr-2" label="Forward" :value="true"/>
              <div class="mr-3">
                <v-hover v-slot:default="{ hover }" open-delay="0">
                  <v-icon small id="forward_lookup_tooltip">
                    {{ hover ? icon.mdiInformation : icon.mdiInformationOutline }}
                  </v-icon>
                </v-hover>
                <v-tooltip right activator="#forward_lookup_tooltip">
                <span>Starting with the given transaction, all mixing transactions
                  connected via outputs will be traversed.</span>
                </v-tooltip>
              </div>
            </v-radio-group>
          </v-col>
          <v-col class="d-flex justify-end align-center">
            <v-btn
                color="primary"
                :disabled="!isSearchable"
                :loading="isLoading"
                @click="handleSearch">
              Search
            </v-btn>
          </v-col>
        </v-row>
        <v-divider class="my-3" v-if="this.transactions.length > 0"/>
        <p v-if="this.transactionCount >=0">
          Found {{ this.transactionCount.toLocaleString() }}
          {{ this.transactionCount === 1 ? 'transaction' : 'transactions' }}
        </p>
        <p v-if="this.transactions.length > 30">Result list is limited to 30 transactions.</p>
      </v-card-text>
    </v-card>
    <div v-if="this.transactions.length > 0">
      <v-card outlined v-for="(tx) in transactions" :key="tx.txhash"
              class="mx-auto mt-3 elevation-4" max-width="1000">
        <v-toolbar :color="tx.privacytype>=0?'purple':''" :dark="!!tx.privacytype" flat>
          <v-toolbar-title>
            <router-link class="linkColor" :to="{ name: txRoute, params: { id: tx.txhash }}">
              {{ tx.txhash }}
            </router-link>
          </v-toolbar-title>
        </v-toolbar>
        <v-card-text>
          <p>Privacy type: {{ getPrivacyTypeLabel(tx.privacytype) }}</p>
          <p class="shorten">Block Hash:
            <router-link :to="{ name: blockRoute, params: { id: tx.bhash }}">
              {{ tx.bhash }}
            </router-link>
          </p>
          <p>Block Id:
            <router-link :to="{ name: blockRoute, params: { id: tx.bid }}">
              {{ tx.bid }}
            </router-link>
          </p>
        </v-card-text>
      </v-card>
    </div>
  </div>
</template>

<script>
import {
  mdiTextBoxSearch, mdiInformation, mdiInformationOutline,
} from '@mdi/js';
import {
  doGet, handleError, getPrivacyTypeLabel,
} from '../../utilities';
import {
  PAGE_TITLE, ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_TRANSACTION_PAGE, ROUTE_CONNECTION_LOOKUP,
} from '../../constants';

export default {
  name: 'ConnectionLookup',
  data() {
    return {
      icon: {
        mdiTextBoxSearch, mdiInformation, mdiInformationOutline,
      },
      blockRoute: ROUTE_NAME_BLOCK_PAGE,
      txRoute: ROUTE_NAME_TRANSACTION_PAGE,
      // v-model
      fromTransaction: '',
      isDirectionForward: false,
      isLoading: false,
      transactions: [],
      transactionCount: -1,
      maxLookBackTime: 5,
    };
  },
  computed: {
    isSearchable() {
      return this.fromTransaction && this.fromTransaction.trim().length > 0;
    },
  },
  methods: {
    getPrivacyTypeLabel,
    setInfoMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'info', temporary: true });
    },
    setWarningMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'warning', temporary: true });
    },
    handleSearch() {
      if (this.isLoading || !this.isSearchable) {
        return;
      }

      this.$store.dispatch('resetMessages');

      this.transactions = [];
      this.doLookup();
    },
    doLookup() {
      this.transactionCount = -1;
      this.isLoading = true;
      const txString = `${this.fromTransaction.trim()}?forward=${this.isDirectionForward ? '1' : '0'}&t=${this.maxLookBackTime}`;
      doGet(ROUTE_CONNECTION_LOOKUP, this.$router, this.$store, txString)
        .then((data) => {
          if (data.success === undefined) throw Error('error searching for paths');
          if (data.success === false) {
            if (data.warning) {
              this.setWarningMessage(data.msg);
            } else throw new Error(data.msg);
          }
          if (data.success === true && data.msg !== undefined) {
            this.setInfoMessage(data.msg);
          }

          if (data.transactions && data.transactions.length > 0) {
            if (data.count) this.transactionCount = data.count; else this.transactionCount = -1;
            this.transactions = data.transactions;
          }
        })
        .catch((e) => {
          handleError(this.$store, e);
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
  },
  mounted() {
    document.title = `Connection Lookup - ${PAGE_TITLE}`;
  },
};
</script>

<style scoped>

.customSlider {
  min-width: 250px;
  max-width: 400px;
}

>>> .customSlider .v-label {
  font-size: 14px;
}

.shorten {
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}

.linkColor {
  color: inherit;
}

</style>
