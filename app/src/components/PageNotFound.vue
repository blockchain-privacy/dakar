<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <v-card>
          <v-card-title class="headline">404 Page not found!</v-card-title>

          <v-card-subtitle>The requested page does not exist. Either navigate back or click below to get back to the
            entry page.
          </v-card-subtitle>
          <v-card-actions v-on:click="toEntryPage">
            <v-btn text>Take me to the entry page</v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import * as Constants from "@/constants";

export default {
  name: "404",
  methods: {
    toEntryPage: function () {
      this.$router.push({name: Constants.ROUTE_NAME_ENTRY_PAGE});
    }
  },
  mounted() {
    document.title = 'Page not found';
  }
}
</script>

<style scoped>

</style>