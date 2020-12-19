<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col></v-col>
      <v-col>
        <v-img src="../assets/bugs.png" max-height="500px" max-width="600px">
        </v-img>
      </v-col>
      <v-col></v-col>
    </v-row>
    <v-row align="center" justify="center">
      <v-col class="col-12" sm="8" md="6" lg="4">
        <v-card style="top: -100px;">
          <v-card-title class="headline">404 - Page not found!</v-card-title>
          <v-card-subtitle>
            The requested page does not exist.
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
  name: '404',
  methods: {
    toEntryPage() {
      this.$router.push({ name: Constants.ROUTE_NAME_ENTRY_PAGE });
    },
  },
  mounted() {
    document.title = '404 - Page not found';
  },
};
</script>

<style scoped>

</style>
