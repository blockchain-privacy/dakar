<template>
  <v-hover v-slot:default="{ hover }" open-delay="300" close-delay="300">
        <span class="anchor">
          {{ text }}
          <!-- call this function when hovered -->
          {{ hover ? requestBlurb() : '' }}
          <v-fade-transition>
            <v-card v-if="hover"
                    class="elevation-4 tooltip"
                    max-width="350px"
                    min-width="300px">
              <v-card-text class="text-center">
                <div v-if="requestedDescription" v-html="description" class="description"/>
                <v-skeleton-loader v-else type="article"/>
              </v-card-text>
            </v-card>
          </v-fade-transition>
        </span>
  </v-hover>
</template>

<script>
import { doGet } from '../../utilities';

export default {
  name: 'WikiTooltip',
  props: {
    text: { type: String, required: true },
    descriptionUrl: { type: String, required: true },
  },
  data() {
    return {
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
        doGet(`/wikiapi/blurb/${this.descriptionUrl}`, this.$router, this.$store)
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
  color: #008ee5;
}

.tooltip {
  position: absolute;
  top: 30px;
  right: 0;
  z-index: 50;
}

.description >>> h1 {
  margin-bottom: 10px;
}

</style>
