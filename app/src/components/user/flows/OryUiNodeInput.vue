<template>
  <input
      v-if="attributes.type === 'hidden'"
      :name="attributes.name"
      :value="attributes.value"
      type="hidden"
  />
  <v-text-field
      v-else-if="attributes.type === 'password'"
      :label="meta.label?meta.label.text?meta.label.text:'':''"
      :value="attributes.value"
      :name="attributes.name"
      type="password"
      :prepend-inner-icon="icons.mdiLockOutline"
  />
  <v-text-field
      v-else-if="attributes.type === 'text'"
      :label="meta.label?meta.label.text?meta.label.text:'':''"
      :value="attributes.value"
      :name="attributes.name"
  />
  <v-text-field
      v-else-if="attributes.type === 'email'"
      :label="meta.label?meta.label.text?meta.label.text:'':''"
      :value="attributes.value"
      :name="attributes.name"
      type="email"
  />
  <div v-else-if="attributes.type === 'submit'">
    <input
        :name="attributes.name"
        :value="attributes.value"
        type="hidden"
    />
    <v-btn
        @click="emitSubmitEvent"
        block
        class="font-weight-bold" color="primary darken-1"
        :id="id"
        type="submit">
      {{ meta.label ? meta.label.text ? meta.label.text : '' : '' }}
    </v-btn>
  </div>

</template>

<script>
import {
  mdiLockOutline,
} from '@mdi/js';

export default {
  name: 'OryUiNodeInput',
  props: {
    meta: { type: Object, required: true },
    attributes: { type: Object, required: true },
    id: { type: String, required: true },
  },
  data() {
    return {
      icons: { mdiLockOutline },
    };
  },
  methods: {
    emitSubmitEvent(event) {
      event.preventDefault();
      this.$emit('submit');
    },
  },
};
</script>

<style scoped>

</style>
