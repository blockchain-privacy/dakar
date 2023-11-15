<template>
  <v-container :fluid="true">
    <v-row
      align="center"
      justify="center"
    >
      <v-col
        cols="12"
        sm="12"
        md="12"
        lg="10"
        xl="8"
      >
        <v-card>
          <v-toolbar :flat="true">
            <v-toolbar-title>{{ pageTitle }}</v-toolbar-title>
          </v-toolbar>
          <v-card-text>
            <div
              v-if="loadedHTML"
              v-html="loadedHTML"
            />
            <v-skeleton-loader
              v-else
              type="article@3"
            />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import {PAGE_TITLE} from '@/constants';

export default {
	name: 'TextLoaderPage',
	props: {
		pageTitle: {type: String, required: true},
		url: {type: String, required: true},
	},
	data() {
		return {
			loadedHTML: '',
		};
	},
	async mounted() {
		document.title = `${this.pageTitle} - ${PAGE_TITLE}`;

		try {
			const response = await fetch(this.url);
			this.loadedHTML = await response.text();
		} catch (_) {
			this.setErrorMessage('Unable to load data, try again later');
		}
	},
	methods: {
		setErrorMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'error', temporary: true, category: this.$route.name});
		},
	},
};
</script>

<style scoped>

</style>
