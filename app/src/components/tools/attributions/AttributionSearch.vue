<template>
  <v-card>
    <v-card-text>
      <v-text-field label="Search public attributions" :append-icon="icon.mdiMagnify"/>
      <v-divider v-if="results.length > 0"/>

    </v-card-text>
  </v-card>
</template>

<script>
import { mdiMagnify } from '@mdi/js';

export default {
  name: 'AttributionSearch',
  data() {
    return {
      icon: {
        mdiMagnify,
      },
      results: [],
    };
  },
};
</script>

<style scoped>

</style>
