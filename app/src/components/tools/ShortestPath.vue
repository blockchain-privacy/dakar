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
      test
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
