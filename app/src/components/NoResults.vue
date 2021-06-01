<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col></v-col>
      <v-col>
        <v-img src="../assets/no_results.webp" max-height="500px" max-width="600px">
        </v-img>
      </v-col>
      <v-col></v-col>
    </v-row>
    <v-row align="center" justify="center">
      <v-col class="col-12" sm="8" md="6" lg="4">
        <v-card style="top: -100px;">
          <v-card-title class="text-h5">No results found</v-card-title>
          <v-card-subtitle>
            Your search query did not return any results.
            Either navigate back or click below to get back to the entry page.
          </v-card-subtitle>
          <v-card-actions v-on:click="toEntryPage">
            <v-btn text>Entry page</v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import * as Constants from '../constants';

export default {
  name: 'NoResults',
  methods: {
    toEntryPage() {
      this.$router.push({ name: Constants.ROUTE_NAME_ENTRY_PAGE });
    },
  },
  mounted() {
    document.title = 'No results found';
  },
};
</script>

<style scoped>

</style>
