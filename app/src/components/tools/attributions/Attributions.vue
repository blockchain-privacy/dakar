<template>
  <v-card
      class="mx-auto elevation-4" max-width="1200">
    <v-toolbar dark flat color="primary" class="mb-1">
      <v-toolbar-title>
        <v-icon>{{ icon.mdiTag }}</v-icon>
        Attributions
      </v-toolbar-title>
      <template v-slot:extension>
        <!-- 'color="white"' should not be necessary, but due to a bug it is.
        If it is not set, the text color of the tabs is always set to 'primary'.
        Should be fixed in vuetify v3 -->
        <v-tabs v-model="tab" align-with-title dark color="white">
          <v-tab>
            Overview
          </v-tab>
          <v-tab>
            Search
          </v-tab>
        </v-tabs>
      </template>
    </v-toolbar>
    <v-tabs-items v-model="tab">
      <v-tab-item>
        <attribution-overview/>
      </v-tab-item>
      <v-tab-item>
        <attribution-search />
      </v-tab-item>
    </v-tabs-items>
  </v-card>
</template>

<script>
import { mdiTag } from '@mdi/js';
import AttributionOverview from './AttributionOverview.vue';
import AttributionSearch from './AttributionSearch.vue';

export default {
  name: 'Attributions',
  components: { AttributionSearch, AttributionOverview },
  data() {
    return {
      icon: {
        mdiTag,
      },
      tab: null,
    };
  },
};
</script>

<style scoped>

</style>
