<template>
  <v-container class="fill-height">
    <v-card
      class="mx-auto"
      variant="text"
      max-width="600px"
    >
      <v-img
        v-if="imageSource"
        class="mb-2"
        :src="imageSource"
      />
      <v-card-title>
        {{ title }}
      </v-card-title>
      <v-card-text>
        {{ errorDescription }}
      </v-card-text>
      <v-card-actions class="d-flex justify-end">
        <v-btn :to="{ name: ROUTE_NAME_ENTRY_PAGE }">
          Go to entry page
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script>
import {ROUTE_NAME_ENTRY_PAGE, PAGE_TITLE} from '@/constants';

export default {
	name: 'ErrorPage',
	props: {
		title: {type: String, required: true},
		description: {type: String, required: true},
		imageSource: {type: String, required: true},
	},
	data() {
		return {
			ROUTE_NAME_ENTRY_PAGE,
			errorDescription: '',
		};
	},
	async mounted() {
		document.title = `${this.$route.meta.title} - ${PAGE_TITLE}`;
		// Set description from prop
		this.errorDescription = this.description;

		// If id query parameter is present, then check if error messages  can be pulled
		if (this.$route.query.id) {
			const response = await this.ory.frontend.getFlowError({id: this.$route.query.id});
			if (response.data?.error?.message) {
				this.errorDescription = `${response.data.error.message}. ${response.data.error.reason}`;
			}
		}
	},
};
</script>

<style scoped>

</style>
