<template>
  <span>
    <v-hover open-delay="0">
      <v-chip :id="uuid" label class="overflow-x-auto" :color="attribution.ispublic?'primary':null">
        {{ attribution.tag }}
      </v-chip>
    </v-hover>
    <v-tooltip bottom :activator="`#${uuid}`" v-if="attribution.ispublic">
      <span>This is a public attribution</span>
    </v-tooltip>
  </span>
</template>

<script>
import { uuidv4 } from '../../../utilities';

export default {
  name: 'AttributionTag',
  props: { attribution: { type: Object, required: true } },
  data() {
    return {
      uuid: '',
    };
  },
  beforeMount() {
    // calculate uuid if tooltip is set
    if (this.attribution.ispublic) this.uuid = `tag_${uuidv4()}`;
  },
};
</script>

<style scoped>

</style>
