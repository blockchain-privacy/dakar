<template>
  <v-menu
    location="bottom"
    :close-on-content-click="false"
    transition="slide-y-transition"
    content-class="mt-7"
  >
    <template #activator="{ props }">
      <a
        v-bind="props"
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
          :to="{name: routeWiki, params: { file: descriptionUrl }}"
          variant="text"
          class="ml-auto"
        >
          <v-icon>{{ icons.mdiOpenInNew }}</v-icon>
          Show full Page
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>

<script>
import {mdiOpenInNew} from '@mdi/js';
import {ROUTE_NAME_WIKI} from '@/constants';

export default {
	name: 'WikiTooltip',
	props: {
		descriptionUrl: {type: String, required: true},
		showLink: {type: Boolean, required: false, default: true},
	},
	data() {
		return {
			icons: {mdiOpenInNew},
			routeWiki: ROUTE_NAME_WIKI,
			showTooltip: false,
			description: '',
			requestedDescription: false,
		};
	},
	methods: {
		async requestBlurb() {
			// Check if already tried to request description
			if (this.requestedDescription) {
				return;
			}

			this.requestedDescription = true;

			try {
				const response = await this.wikiapi.blurbFileNameGet({fileName: this.descriptionUrl});
				if (response.blurb) {
					this.description = response.blurb;
				}
			} catch (e) {
				this.setErrorMessage(e);
			}
		},
	},
};
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
