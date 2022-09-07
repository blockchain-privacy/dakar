<template>
  <v-hover v-slot:default="{ hover }" open-delay="300" close-delay="300">
        <span class="anchor">
          <router-link :to="{name: routeWiki, query: { file: descriptionUrl }}">
            {{ text }}
          </router-link>
          <!-- call this function when hovered -->
          {{ hover ? requestBlurb() : '' }}
          <v-fade-transition>
            <v-card v-if="hover"
                    class="elevation-4 tooltip"
                    max-width="350px"
                    min-width="300px">
              <v-card-text>
                <div v-if="requestedDescription" v-html="description" class="wikiBlurbDescription"/>
                <v-skeleton-loader v-else type="article"/>
              </v-card-text>
            </v-card>
          </v-fade-transition>
        </span>
  </v-hover>
</template>

<script>
import { doGet } from '../../utilities';
import { WIKIAPI_PATH_PREFIX, ROUTE_NAME_WIKI } from '../../constants';

export default {
  name: 'WikiTooltip',
  props: {
    text: { type: String, required: true },
    descriptionUrl: { type: String, required: true },
  },
  data() {
    return {
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

      return '';
    },
  },
};
</script>

<style scoped>
.anchor {
  position: relative;

}

.tooltip {
  position: absolute;
  top: 30px;
  right: 0;
  z-index: 50;
}

.wikiBlurbDescription >>> h1 {
  margin-bottom: 10px;
}

</style>
