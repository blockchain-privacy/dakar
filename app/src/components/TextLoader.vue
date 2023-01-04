<template>
  <v-container fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="12" lg="10" xl="8">
        <v-card class="elevation-4">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title>{{ pageTitle }}</v-toolbar-title>
          </v-toolbar>
          <v-card-text>
            <div v-if="loadedHTML" v-html="loadedHTML" />
            <v-skeleton-loader v-else type="article@3" />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { PAGE_TITLE } from '../constants';

export default {
  name: 'TextLoader',
  props: {
    pageTitle: { type: String, required: true },
    url: { type: String, required: true },
  },
  data() {
    return {
      loadedHTML: '',
    };
  },
  methods: {
    setErrorMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'error', temporary: true });
    },
  },
  mounted() {
    document.title = `${this.pageTitle} - ${PAGE_TITLE}`;

    // get HTML data
    fetch(this.url).then((data) => data.text()).then((data) => {
      this.loadedHTML = data;
    }).catch(() => {
      this.setErrorMessage('Unable to load data, try again later');
    });
  },
};
</script>

<style scoped>

</style>
