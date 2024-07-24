<template>
  <v-menu
    location="bottom"
    :close-on-content-click="false"
    transition="slide-y-transition"
    content-class="mt-7"
  >
    <template #activator="item">
      <a
        v-bind="item.props"
        :class="{'anchor': true,'d-inline-block':true, 'underline': showLink}"
        @click="requestBlurb"
      ><slot /></a>
    </template>
    <v-card
      class="tooltip"
      max-width="350px"
      min-width="300px"
    >
      <v-card-text>
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
          Show full Page
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>

<script setup>
import {mdiOpenInNew} from '@mdi/js';
import {ROUTE_NAME_WIKI} from '@/constants';
import {inject, ref} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const route = useRoute();
const wikiapi = inject('wikiapi');
const msgStore = useMsgStore();

const props = defineProps({
	descriptionUrl: {type: String, required: true},
	showLink: {type: Boolean, required: false, default: true},
});

const description = ref('');
const requestedDescription = ref(false);

function setErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: true, category: route.name,
	});
}

async function requestBlurb() {
	// Check if already tried to request description
	if (requestedDescription.value) {
		return;
	}

	requestedDescription.value = true;

	try {
		const response = await wikiapi.blurbFileNameGet({fileName: props.descriptionUrl});
		if (response.blurb) {
			description.value = response.blurb;
		}
	} catch (e) {
		setErrorMessage(e);
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

.wikiBlurbDescription :deep( h1) {
  margin-bottom: 10px;
  line-height: 1em;
}

.wikiBlurbDescription :deep( img) {
  max-width: 100%
}

</style>
