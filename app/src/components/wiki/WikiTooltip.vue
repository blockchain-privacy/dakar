<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-menu
    location="bottom"
    :close-on-content-click="false"
    transition="slide-y-transition"
    content-class="mt-7"
  >
    <template #activator="item">
      <v-btn
        v-if="icon"
        :icon="icon"
        v-bind="item.props"
        :class="$attrs.class"
        :style="$attrs.style"
        variant="text"
        @click="requestBlurb"
      />
      <a
        v-else
        v-bind="{...$attrs, ...item.props}"
        :class="{'anchor': true,'d-inline-block':true, 'underline': !hideLink}"
        :style="$attrs.style"
        @click="requestBlurb"
      ><slot /></a>
    </template>
    <v-card
      class="tooltip"
      max-width="350px"
      min-width="300px"
    >
      <v-card-text>
        <alert :text="errorMsg" />
        <!-- html is loaded from safe source -->
        <!-- eslint-disable vue/no-v-html -->
        <div
          v-if="requestedDescription"
          class="wikiBlurbDescription"
          v-html="description"
        />
        <v-skeleton-loader
          v-else
          type="article"
        />
      </v-card-text>
      <v-card-actions class="d-flex">
        <v-btn
          :to="{name: ROUTE_NAME_WIKI, params: { file: descriptionUrl }}"
          variant="text"
          class="ml-auto"
        >
          <v-icon>{{ mdiOpenInNew }}</v-icon>
          Show Page
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>

<script setup>
import {mdiOpenInNew} from '@mdi/js';
import {inject, ref} from 'vue';
import {ROUTE_NAME_WIKI} from '@/constants';
import Alert from '@/components/common/Alert.vue';

const wikiapi = inject('wikiapi');

const props = defineProps({
	descriptionUrl: {type: String, required: true},
	hideLink: {type: Boolean, required: false},
	icon: {type: String, required: false, default: ''},
	iconColor: {type: String, required: false, default: undefined},
});

const description = ref('');
const requestedDescription = ref(false);
const errorMsg = ref('');

// Functions
async function requestBlurb() {
	// Check if already tried to request description
	if (requestedDescription.value) {
		return;
	}

	requestedDescription.value = true;
	errorMsg.value = '';
	try {
		const response = await wikiapi.blurbFileNameGet({fileName: props.descriptionUrl});
		if (response.blurb) {
			description.value = response.blurb;
		}
	} catch (error) {
		errorMsg.value = error;
	}
}

</script>

<style scoped>

.anchor {
  cursor: pointer;
}

.underline {
  color: rgb(var(--v-theme-primary));
  text-decoration: underline;
}

.wikiBlurbDescription :deep(h1) {
  margin-bottom: 10px;
  line-height: 1em;
}

.wikiBlurbDescription :deep(img) {
  max-width: 100%
}

.wikiBlurbDescription :deep(ul){
  padding-left: 20px;
}

</style>
