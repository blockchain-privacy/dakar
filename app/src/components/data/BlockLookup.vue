<template>
  <v-container fluid v-if="data">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="12" lg="10" xl="8">
        <v-row>
          <v-col>
            <v-card class="elevation-4">
              <v-toolbar color="primary" dark flat>
                <v-toolbar-title>
                  <v-icon>{{ icon.mdiCubeOutline }}</v-icon>
                  Block {{ data.blockhash }}
                </v-toolbar-title>
              </v-toolbar>
              <v-card-text>
                <v-container>
                  <v-row>
                    <v-col v-if="data.id">
                      <IconItem :icon="icon.mdiFormatListNumbered"
                                title="Block Height" :subtitle="data.id">
                        {{ data.id }}
                      </IconItem>
                    </v-col>
                    <v-col v-if="data.ts">
                      <IconItem :icon="icon.mdiCalendar" title="Timestamp">
                        {{ data.ts != null ? new Date(data.ts).toLocaleString() : "" }}
                      </IconItem>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col v-if="data.prevblockhash">
                      <IconItem :icon="icon.mdiFormatHeaderPound" title="Previous Block">
                        <router-link :to="{ name: blockRoute,
                    params: { id: data.prevblockhash }}">
                          {{ shortenHash(data.prevblockhash) }}
                        </router-link>
                      </IconItem>
                    </v-col>
                    <v-col v-if="data.nextblockhash">
                      <IconItem :icon="icon.mdiFormatHeaderPound" title="Next Block">
                        <router-link :to="{ name: blockRoute,
                    params: { id: data.nextblockhash }}">
                          {{ shortenHash(data.nextblockhash) }}
                        </router-link>
                      </IconItem>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col>
                      <IconItem :icon="icon.mdiPound" title="Number of Transactions">
                        {{ data.txcount.toLocaleString() }}
                      </IconItem>
                    </v-col>
                  </v-row>
                </v-container>
              </v-card-text>
            </v-card>
          </v-col>
          <template v-if="data.transactions">
            <v-col v-for="tx in data.transactions" :key="tx.txhash+tx.bid">
              <Transaction :tx="tx" show-title-link
                           :show-heuristic-editor-link="showHeuristicEditor"
                           :show-fingerprint-link="showHeuristicEditor"/>
            </v-col>
          </template>
        </v-row>
        <v-row v-if="this.emptyResponse">
          <v-col class="d-flex justify-center">
            <p class="text-h6">No transactions found</p>
          </v-col>
        </v-row>
        <v-row v-if="this.isLoadingMore">
          <v-col>
            <v-progress-linear
                indeterminate
                rounded
                height="6"
            ></v-progress-linear>
          </v-col>
        </v-row>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import {
  mdiCubeOutline, mdiFormatListNumbered, mdiCalendar,
  mdiFormatHeaderPound, mdiTransfer, mdiPound,
} from '@mdi/js';
import {
  doPost, handleError, isAdminIdentity, isPrivilegedIdentity, shortenHash,
} from '../../utilities';
import {
  PAGE_TITLE,
  ROUTE_NAME_BLOCK_PAGE,
  ROUTE_NAME_TRANSACTION_PAGE,
  ROUTE_BLOCK_RANGE,
} from '../../constants';
import IconItem from '../common/IconItem.vue';
import Transaction from './Transaction.vue';

export default {
  name: 'BlockLookup',
  components: { IconItem, Transaction },
  data() {
    return {
      icon: {
        mdiCubeOutline,
        mdiFormatListNumbered,
        mdiCalendar,
        mdiFormatHeaderPound,
        mdiTransfer,
        mdiPound,
      },
      blockRoute: ROUTE_NAME_BLOCK_PAGE,
      transactionRoute: ROUTE_NAME_TRANSACTION_PAGE,
      offset: 0,
      isLoading: false,
      isLoadingMore: false,
      // emptyResponse is only used for data loaded after the initial data load
      emptyResponse: false,
    };
  },
  computed: {
    data() {
      return this.$store.getters.getBlockData;
    },
    session() {
      return this.$store.getters.getSession;
    },
    showHeuristicEditor() {
      return isPrivilegedIdentity(this.session) || isAdminIdentity(this.session);
    },
  },
  methods: {
    shortenHash,
    setPageTitle() {
      let id = ' ';
      if (this.data && this.data.id) {
        id = ` ${this.data.id} `;
      }
      document.title = `Block${id}- ${PAGE_TITLE}`;
    },
    isResponseValid(data) {
      return !(!data.type || data.type !== 'block' || !data.payload || !data.payload.transactions
          || data.payload.transactions.length === 0);
    },
    addNewData() {
      if (!this.data) return;

      this.offset += 10;

      // do nothing if all data is already loaded
      if (this.offset >= this.data.txcount) return;
      this.isLoadingMore = true;

      doPost(
        ROUTE_BLOCK_RANGE,
        this.$router,
        this.$store,
        { offset: this.offset },
        this.data.blockhash,
      )
        .then((data) => {
          if (!this.isResponseValid(data)) {
            this.emptyResponse = true;
            return;
          }

          this.data.transactions = [...this.data.transactions, ...data.payload.transactions];
          this.$store.dispatch('resetMessages');
          this.emptyResponse = false;
        })
        .catch((e) => {
          handleError(this.$store, e);
        })
        .finally(() => {
          this.isLoadingMore = false;
        });
    },
    handleScroll() {
      // return if not bottom of page
      if (this.isLoadingMore
          || this.loading
          || document.documentElement.scrollTop + window.innerHeight
          !== document.documentElement.offsetHeight) return;
      this.addNewData();
    },
  },
  mounted() {
    this.setPageTitle();
    // register scroll handler
    window.onscroll = this.handleScroll;
    this.offset = 0;
  },
  beforeDestroy() {
    // unregister scroll handler
    window.onscroll = null;
  },
  updated() {
    this.setPageTitle();
  },
  watch: {
    $route() {
      // if route gets changed the component could still be loaded but now with different data.
      // Because of this the internal state has to be reset.
      this.offset = 0;
      this.isLoading = false;
      this.isLoadingMore = false;
      this.emptyResponse = false;
    },
  },
};
</script>
