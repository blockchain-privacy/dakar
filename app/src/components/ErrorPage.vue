<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-container class="fill-height d-flex justify-center align-center flex-wrap">
    <v-card
      variant="text"
      max-width="600px"
      style="width: 100%"
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
        {{ errorDescription?errorDescription:description }}
      </v-card-text>
      <v-card-actions
        v-if="!hideActions"
        class="d-flex justify-end"
      >
        <v-btn :to="{ name: ROUTE_NAME_ENTRY_PAGE}">
          Go to entry page
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script setup>
import {inject, onMounted, ref} from 'vue';
import {useRoute} from 'vue-router';
import {ROUTE_NAME_ENTRY_PAGE, PAGE_TITLE} from '@/constants';

const props = defineProps({
	title: {type: String, required: true},
	description: {type: String, required: true},
	imageSource: {type: String, required: true},
	hideActions: {type: Boolean, required: false},
});

const route = useRoute();
const ory = inject('ory');

const errorDescription = ref('');

onMounted(async () => {
	document.title = `${props.title} - ${PAGE_TITLE}`;

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
