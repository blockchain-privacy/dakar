<template>
  <v-container fluid>
    <v-data-iterator
        :items="items"
        :items-per-page.sync="itemsPerPage"
        :page.sync="page"
        :search="search"
        :sort-by="sortBy.toLowerCase()"
        :sort-desc="sortDesc"
        hide-default-footer>
      <template v-slot:header>
        <v-toolbar class="mb-1" flat>
          <v-text-field v-model="search" clearable flat solo-inverted hide-details
                        :prepend-inner-icon="icons.mdiMagnify" label="Search"/>
          <template v-if="$vuetify.breakpoint.mdAndUp">
            <v-spacer/>
            <v-select
                v-model="sortBy"
                flat
                solo-inverted
                hide-details
                :items="keys"
                :prepend-inner-icon="icons.mdiMagnify"
                label="Sort by"/>
            <v-spacer/>
            <v-btn-toggle v-model="sortDesc" mandatory>
              <v-btn large depressed color="blue" :value="false">
                <v-icon>{{ icons.mdiArrowUp }}</v-icon>
              </v-btn>
              <v-btn large depressed color="blue" :value="true">
                <v-icon>{{ icons.mdiArrowDown }}</v-icon>
              </v-btn>
            </v-btn-toggle>
          </template>
        </v-toolbar>
      </template>
      <template v-slot:default="props">
        <v-row>
          <v-col v-for="item in props.items" :key="item.cluster" cols="12" sm="6" md="4" lg="3">
            <v-card outlined>
              <v-card-title class="subheading font-weight-bold">
                Cluster ID {{ item.id }}, txcount:  {{item.txCount}}
                Address count {{ item.address_count}}
              </v-card-title>
            </v-card>
          </v-col>
        </v-row>
      </template>
      <template v-slot:footer>
        <span class="mr-4 grey--text">
          Page {{ page }} of {{ numberOfPages }}
        </span>
        <v-btn icon class="mr-1" @click="formerPage">
          <v-icon>{{ icons.mdiChevronLeft }}</v-icon>
        </v-btn>
        <v-btn icon class="ml-1" @click="nextPage">
          <v-icon>{{ icons.mdiChevronRight }}</v-icon>
        </v-btn>
      </template>
    </v-data-iterator>
  </v-container>
</template>

<script>
import {
  mdiChevronLeft, mdiChevronRight, mdiMagnify, mdiArrowUp, mdiArrowDown,
  mdiChevronDown,
} from '@mdi/js';

export default {
  name: 'Results',
  props: {
    items: { type: Array, required: true },
  },
  computed: {
    numberOfPages() {
      return Math.ceil(this.items.length / this.itemsPerPage);
    },
  },
  data() {
    return {
      sortBy: 'address_count',
      sortDesc: false,
      itemsPerPage: 15,
      itemsPerPageArray: [4, 8, 12],
      search: '',
      page: 1,
      keys: [
        'txCount',
        'address_count',
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
  methods: {
    nextPage() {
      if (this.page + 1 <= this.numberOfPages) this.page += 1;
    },
    formerPage() {
      if (this.page - 1 >= 1) this.page -= 1;
    },
    updateItemsPerPage(number) {
      this.itemsPerPage = number;
    },
  },
};
</script>

<style scoped>

</style>
