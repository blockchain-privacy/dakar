<template>
  <div class="my-2 mx-1">
    <v-card elevation="4">
      <v-card-text>
        <v-text-field
            v-model="query"
            label="Search for attributions"
            :append-icon="icon.mdiMagnify"
            @click:append="handleQuery"
            @keydown.enter="handleQuery"/>
        <v-progress-linear v-if="loading"/>
      </v-card-text>
    </v-card>
    <v-row v-if="!loading && attributions.length > 0" class="mt-2">
      <v-col v-for="(attribution, i) in attributions" :key="i" cols="12" sm="6" md="4" lg="4">
        <attribution-details :attribution="attribution" @deleted="handleAttributionDeletion"/>
      </v-col>
    </v-row>
  </div>

</template>

<script>
import { mdiMagnify } from '@mdi/js';
import AttributionDetails from './AttributionDetails.vue';
import { doPost, handleError } from '../../../utilities';
import { ROUTE_SEARCH_ATTRIBUTIONS } from '../../../constants';

function isValidQuery(query) {
  return query.trim().length > 0;
}

export default {
  name: 'AttributionSearch',
  components: { AttributionDetails },
  data() {
    return {
      icon: {
        mdiMagnify,
      },
      loading: false,
      query: '',
      attributions: [],
    };
  },
  methods: {
    setWarningMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'warning', temporary: true });
    },
    handleQuery() {
      const q = this.query;
      if (!isValidQuery(q)) {
        this.setWarningMessage('search query is not valid');
        return;
      }
      this.loadData(q);
    },
    loadData(query) {
      this.loading = true;
      this.attributions = [];
      doPost(ROUTE_SEARCH_ATTRIBUTIONS, this.$router, this.$store, { q: query })
        .then((data) => {
          if (!data.success || data.attributions === undefined) throw new Error('could not get attribution data');

          if (data.attributions === null) {
            this.attributions = [];
            return;
          }

          // parse date
          data.attributions = data.attributions.map((d) => {
            d.ts = new Date(d.ts);
            return d;
          });

          // sort attributions by time stamp
          this.attributions = data.attributions.sort((a, b) => b.ts - a.ts);
        })
        .catch((e) => {
          handleError(this.$store, e);
        })
        .finally(() => {
          this.loading = false;
        });
    },
    handleAttributionDeletion(attributionUid) {
      this.attributions = this.attributions.filter((d) => d.uid !== attributionUid);
    },
  },
};
</script>

<style scoped>

</style>
