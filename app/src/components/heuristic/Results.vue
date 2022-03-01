<template>
  <v-container fluid v-if="items && items.length > 0">
    <v-row>
      <v-col>
        <v-text-field v-model="search" clearable flat hide-details style="min-width: 230px"
                      :prepend-inner-icon="icons.mdiMagnify" label="Search"/>
      </v-col>
      <v-col>
        <v-select v-model="sortBy" flat hide-details :items="keys"
                  label="Sort by" style="min-width: 230px"/>
      </v-col>
      <v-col cols="12" sm="3" md="3" lg="3">
        <v-btn-toggle v-model="sortDesc" mandatory>
          <v-btn large depressed :value="false">
            <v-icon>{{ icons.mdiArrowUp }}</v-icon>
          </v-btn>
          <v-btn large depressed :value="true">
            <v-icon>{{ icons.mdiArrowDown }}</v-icon>
          </v-btn>
        </v-btn-toggle>
      </v-col>
    </v-row>
    <v-data-iterator
        :items="clusters"
        item-key="id"
        :items-per-page.sync="itemsPerPage"
        :page.sync="page"
        :sort-by="sortBy"
        :sort-desc="sortDesc"
        hide-default-footer>
      <template v-slot:default="props">
        <v-row>
          <v-col v-for="cluster in props.items" :key="cluster.id" cols="12" sm="12" md="8" lg="6">
            <v-card outlined>
              <v-card-title>
                <v-expansion-panels flat>
                  <v-expansion-panel>
                    <v-expansion-panel-header>
                      {{ cluster.transactionCount }}
                      {{ plural('transaction', cluster.transactionCount) }}
                    </v-expansion-panel-header>
                    <v-expansion-panel-content>
                      <v-list>
                        <v-list-item v-for="tx in cluster.txs" :key="tx.txhash"
                                     :to="{ name: routes.ROUTE_NAME_TRANSACTION_PAGE,
                                  params: { id: tx.txhash }}">
                          <v-list-item-title>
                            {{ tx.txhash }}
                            <div v-if="tx.destinationCount">
                              Destinations: {{ tx.destinationCount }}
                            </div>
                          </v-list-item-title>
                        </v-list-item>
                      </v-list>
                    </v-expansion-panel-content>
                  </v-expansion-panel>
                </v-expansion-panels>
              </v-card-title>
              <v-card-subtitle>
                <AttributionTag v-for="(attribution, i) in cluster.attributions"
                                :key="i" :attribution="attribution"/>
              </v-card-subtitle>
            </v-card>
          </v-col>
        </v-row>
      </template>
    </v-data-iterator>
    <span class="mr-4 grey--text">
          Page {{ page }} of {{ numberOfPages }}
        </span>
    <v-btn icon class="mr-1" @click="formerPage">
      <v-icon>{{ icons.mdiChevronLeft }}</v-icon>
    </v-btn>
    <v-btn icon class="ml-1" @click="nextPage">
      <v-icon>{{ icons.mdiChevronRight }}</v-icon>
    </v-btn>
  </v-container>
</template>

<script>
import {
  mdiChevronLeft, mdiChevronRight, mdiMagnify, mdiArrowUp, mdiArrowDown,
  mdiChevronDown,
} from '@mdi/js';
import AttributionTag from '../tools/attributions/AttributionTag.vue';
import { ROUTE_NAME_TRANSACTION_PAGE } from '../../constants';

// plural appends an 's' at the end of subject if count is higher than one
function plural(subject, count) {
  return count > 1 ? `${subject}s` : subject;
}

export default {
  name: 'Results',
  components: { AttributionTag },
  props: {
    items: { type: Array, required: true },
  },
  data() {
    return {
      routes: {
        ROUTE_NAME_TRANSACTION_PAGE,
      },
      sortBy: 'transactionCount',
      sortDesc: false,
      itemsPerPage: 15,
      itemsPerPageArray: [4, 8, 12],
      search: '',
      page: 1,
      keys: [
        { text: 'Number of Transactions', value: 'transactionCount' },
        { text: 'Number of Attributions', value: 'attributionCount' },
      ],
      icons: {
        mdiChevronLeft,
        mdiChevronRight,
        mdiMagnify,
        mdiArrowUp,
        mdiArrowDown,
        mdiChevronDown,
      },
    };
  },
  computed: {
    numberOfPages() {
      return Math.ceil(this.items.length / this.itemsPerPage);
    },
    clusters() {
      if (!this.items) return [];

      let clusterCounter = 0;

      // filter items based on search query and set counts
      return this.items.filter((cluster) => {
        const query = this.search.trim();
        if (!query) return true;

        let found = false;

        // check if any transaction hash contains the search query
        cluster.txs.some((tx) => {
          if (tx.txhash.includes(query)) {
            found = true;
            return true;
          }
          return true;
        });

        if (!found && cluster.attributions) {
          // check if any flag contains the search query
          cluster.attributions.some((attribution) => {
            if (attribution.tag.includes(query)) {
              found = true;
              return true;
            }
            return true;
          });
        }

        return found;
      }).map((cluster) => {
        // check if additional attributes are already set
        if (cluster.id !== undefined) {
          return cluster;
        }

        cluster.id = clusterCounter;
        clusterCounter += 1;

        // set transaction count
        cluster.transactionCount = 0;
        if (cluster.txs) {
          cluster.transactionCount = cluster.txs.length;
        }

        // set attribution count
        cluster.attributionCount = 0;
        if (cluster.attributions) {
          cluster.attributionCount = cluster.attributions.length;
        }

        return cluster;
      });
    },
  },
  methods: {
    plural,
    nextPage() {
      if (this.page + 1 <= this.numberOfPages) this.page += 1;
    },
    formerPage() {
      if (this.page - 1 >= 1) this.page -= 1;
    },
  },
};
</script>

<style scoped>

</style>
