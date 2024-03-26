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

<script setup>
import {ROUTE_NAME_ENTRY_PAGE, PAGE_TITLE} from '@/constants';
import {inject, onMounted, ref} from 'vue';
import {useRoute} from 'vue-router';

const props = defineProps({
	title: {type: String, required: true},
	description: {type: String, required: true},
	imageSource: {type: String, required: true},
});

const errorDescription = ref('');

const route = useRoute();
const ory = inject('ory');

onMounted(async () => {
	document.title = `${route.meta.title} - ${PAGE_TITLE}`;
	// Set description from prop
	errorDescription.value = props.description;

	// If id query parameter is present, then check if error messages can be pulled
	if (route.query.id) {
		const response = await ory.frontend.getFlowError({id: route.query.id});
		if (response?.error?.message) {
			errorDescription.value = `${response.error.message}. ${response.error.reason}`;
		}
	}
});

</script>

<style scoped>

</style>
