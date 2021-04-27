<template>
  <v-card
      class="mx-auto elevation-4"
      max-width="1000">
    <v-toolbar color="primary" dark flat>
      <v-toolbar-title>
        <v-icon>{{ icon.mdiChartTimelineVariant }}</v-icon>
        Shortest Path
      </v-toolbar-title>
    </v-toolbar>
    <v-card-text>
      <div class="text-subtitle-1">
        Find the shortest path between two transactions.
        <v-hover v-slot:default="{ hover }" open-delay="0">
          <v-icon small id="shortest_path_tooltip">
            {{ hover ? icon.mdiInformation : icon.mdiInformationOutline }}
          </v-icon>
        </v-hover>
        <v-tooltip right activator="#shortest_path_tooltip">
          <span>The result is nondeterministic,
            because multiple paths of the same length can exist.</span>
        </v-tooltip>
      </div>
      <v-row>
        <v-col>
          <v-text-field label="From" v-model="fromTransaction" :disabled="isLoading" autofocus/>
        </v-col>
        <v-col>
          <v-text-field label="To" v-model="toTransaction" :disabled="isLoading"/>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <v-radio-group
              v-model="anyDirection"
              mandatory
              row
              label="Search direction:"
              :disabled="isLoading">
            <v-radio label="Linear" :value="false"/>
            <v-radio label="Any" :value="true"/>
          </v-radio-group>
        </v-col>
        <v-col>
          <v-switch
              label="Traverse private transactions"
              class="mx-5"
              :disabled="isLoading"
              v-model="includePrivacyTransactions"
          />
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
      <v-timeline :dense="this.$vuetify.breakpoint.smAndDown" v-if="this.transactions.length > 0">
        <v-timeline-item
            v-for="(tx) in transactions" :key="tx.txhash"
            :color="tx.privacytype>=0?'purple':'primary'"
            small>
          <template v-slot:opposite>
            <span class="headline" v-text="new Date(tx.bts).toLocaleString()"></span>
          </template>
          <v-card outlined>
            <v-toolbar :color="tx.privacytype>=0?'purple':''" :dark="!!tx.privacytype" flat>
              <v-toolbar-title>
                <router-link class="linkColor" :to="{ name: txRoute, params: { id: tx.txhash }}">
                {{ tx.txhash }}
                </router-link>
              </v-toolbar-title>
            </v-toolbar>
            <v-card-text>
              <p>Privacy type: {{ getPrivacyTypeLabel(tx.privacytype) }}</p>
              <p class="shorten">Blockhash:
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
        </v-timeline-item>
      </v-timeline>
    </v-card-text>
  </v-card>
</template>

<script>
import {
  mdiChartTimelineVariant, mdiInformation, mdiInformationOutline,
} from '@mdi/js';
import {
  doPost, handleError, shortenHash, getPrivacyTypeLabel,
} from '../../utilities';
import {
  PAGE_TITLE, ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_TRANSACTION_PAGE, ROUTE_SHORTEST_TRANSACTION_PATH,
} from '../../constants';

export default {
  name: 'ShortestPath',
  data() {
    return {
      icon: {
        mdiChartTimelineVariant, mdiInformation, mdiInformationOutline,
      },
      blockRoute: ROUTE_NAME_BLOCK_PAGE,
      txRoute: ROUTE_NAME_TRANSACTION_PAGE,
      // v-model
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
  methods: {
    getPrivacyTypeLabel,
    shortenHash,
    setInfoMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'info', temporary: true });
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
      this.isLoading = true;
      doPost(ROUTE_SHORTEST_TRANSACTION_PATH, this.$router, {
        to: this.fromTransaction.trim(),
        from: this.toTransaction.trim(),
        includePrivacyTransactions: this.includePrivacyTransactions,
        anyDirection: this.anyDirection,
      })
        .then((data) => {
          if (data.success === undefined) throw Error('error searching for paths');
          if (data.success === false) throw new Error(data.msg);
          if (data.success === true && data.msg !== undefined) {
            this.setInfoMessage(data.msg);
          }

          if (data.transactions && data.transactions.length > 0) {
            if (this.fromTransaction.trim() !== data.transactions[0].txhash) {
              data.transactions = data.transactions.reverse();
            }

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
    document.title = `Shortest Path - ${PAGE_TITLE}`;
  },
};
</script>

<style scoped>

.shorten {
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}

.linkColor {
  color: inherit;
}

</style>
