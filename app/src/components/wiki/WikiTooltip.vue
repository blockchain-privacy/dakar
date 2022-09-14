<template>
  <v-menu bottom :close-on-content-click="false"
          transition="slide-y-transition" content-class="mt-7">
    <template v-slot:activator="{ on, attrs }">
      <span v-bind="attrs" v-on="on" class="anchor" @click="requestBlurb"><slot/></span>
    </template>
    <v-card class="elevation-4 tooltip" max-width="350px" min-width="300px">
      <v-card-text>
        <div v-if="requestedDescription" v-html="description" class="wikiBlurbDescription"/>
        <v-skeleton-loader v-else type="article"/>
      </v-card-text>
      <v-card-actions class="d-flex">
        <v-btn :to="{name: routeWiki, params: { file: descriptionUrl }}" text class="ml-auto">
          <v-icon>{{ icons.mdiOpenInNew }}</v-icon>
          Show full Page
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>

<script>
import { mdiOpenInNew } from '@mdi/js';
import { doGet } from '../../utilities';
import { WIKIAPI_PATH_PREFIX, ROUTE_NAME_WIKI } from '../../constants';

export default {
  name: 'WikiTooltip',
  props: {
    descriptionUrl: { type: String, required: true },
  },
  data() {
    return {
      icons: { mdiOpenInNew },
      routeWiki: ROUTE_NAME_WIKI,
      showTooltip: false,
      description: '',
      requestedDescription: false,
    };
  },
  methods: {
    requestBlurb() {
      // check if already tried to request description
      if (!this.requestedDescription) {
        this.requestedDescription = true;
        doGet(`${WIKIAPI_PATH_PREFIX}/blurb/${this.descriptionUrl}`, this.$router, this.$store)
          .then((d) => {
            if (!d.success) return;
            if (d.data && d.data.blurb) this.description = d.data.blurb;
          })
          .catch((e) => this.setErrorMessage(e));
      }
    },
  },
};
</script>

<style scoped>

.anchor {
  color: var(--v-primary-base);
  text-decoration: underline;
}

.wikiBlurbDescription >>> h1 {
  margin-bottom: 10px;
}

.wikiBlurbDescription >>> img {
  max-width: 100%
}

</style>
