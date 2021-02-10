<template>
  <v-card
      class="mx-auto elevation-12"
      max-width="700">
    <v-toolbar color="primary" dark flat>
      <v-toolbar-title>
        <v-icon>{{ icon.mdiChartTimelineVariant }}</v-icon>
        Shortest Path
      </v-toolbar-title>
    </v-toolbar>
    <v-card-text>
      <div class="text-subtitle-1">
        Find the shortest path between two transactions
      </div>
      <v-row>
        <v-col>
          <v-text-field label="From"/>
        </v-col>
        <v-col>
          <v-text-field label="To"/>
        </v-col>
      </v-row>
      <v-row >
        <v-col class="d-flex justify-end">
          <v-btn color="primary">
            Search
          </v-btn>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script>
import {
  mdiChartTimelineVariant,
} from '@mdi/js';
import { PAGE_TITLE } from '../../constants';

export default {
  name: 'ShortestPath',
  data() {
    return {
      icon: {
        mdiChartTimelineVariant,
      },
    };
  },
  mounted() {
    document.title = `Shortest Path - ${PAGE_TITLE}`;
  },
};
</script>

<style scoped>

</style>
