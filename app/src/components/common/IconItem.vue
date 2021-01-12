<template>
  <v-list-item>
    <v-list-item-avatar>
      <v-icon v-if="!isColor" style="max-width: 32px">{{ icon }}</v-icon>
      <v-icon style="max-width: 32px"
              v-if="isColor"
              v-bind:class="{ 'green--text': !isRed, 'red--text': isRed }">
        {{ icon }}
      </v-icon>
    </v-list-item-avatar>
    <v-list-item-content>
      <v-list-item-title>
        {{ title }}
        <v-hover v-slot:default="{ hover }" open-delay="0" v-if="tooltip">
          <v-icon small :id="uuid">
            {{ hover ? icons.mdiHelpCircle : icons.mdiHelpCircleOutline }}
          </v-icon>
        </v-hover>
        <v-tooltip right v-if="tooltip" :activator="`#${uuid}`">
          <span>{{ tooltip }}</span>
        </v-tooltip>
      </v-list-item-title>
      <v-list-item-subtitle>
        <slot/>
      </v-list-item-subtitle>
    </v-list-item-content>
  </v-list-item>
</template>

<script>
import {
  mdiHelpCircle, mdiHelpCircleOutline,
} from '@mdi/js';

// uuidv4 generates a pseudo random unique id
// credits:
// - https://stackoverflow.com/questions/105034/how-to-create-a-guid-uuid
// - https://gist.github.com/jed/982883
function uuidv4() {
  return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, (c) => (
  // eslint-disable-next-line no-bitwise
    c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4)
    .toString(16));
}

export default {
  name: 'IconItem',
  props: {
    title: { type: String, required: true },
    icon: { type: String, required: true },
    tooltip: { type: String, default: '' },
    isColor: { type: Boolean, default: false },
    isRed: { type: Boolean, default: false },
  },
  data() {
    return {
      icons: {
        mdiHelpCircle, mdiHelpCircleOutline,
      },
      uuid: '',
    };
  },
  beforeMount() {
    // calculate uuid if tooltip is set
    if (this.tooltip !== '') this.uuid = `a${uuidv4()}`;
  },
};
</script>

<style scoped>

</style>
