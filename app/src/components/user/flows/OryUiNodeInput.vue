<template>
  <input
      v-if="attributes.type === 'hidden'"
      :name="attributes.name"
      :value="attributes.value"
      type="hidden"
  />
  <v-text-field
      v-else-if="attributes.type === 'password'"
      :label="metaLabel"
      :value="attributes.value"
      :name="attributes.name"
      :prepend-inner-icon="icons.mdiLockOutline"
      :type="password.show ? 'text' : 'password'"
      :append-icon="password.show ?  icons.mdiEye : icons.mdiEyeOff"
      @click:append="password.show = !password.show"
  />
  <v-text-field
      v-else-if="attributes.type === 'text'"
      :label="metaLabel"
      :value="attributes.value"
      :name="attributes.name"
      :prepend-inner-icon="metaLabel === 'ID'?icons.mdiAccount:null"
  />
  <v-text-field
      v-else-if="attributes.type === 'email'"
      :label="metaLabel"
      :value="attributes.value"
      :name="attributes.name"
      type="email"
      :prepend-inner-icon="icons.mdiEmail"
  />
  <div v-else-if="attributes.type === 'submit'">
    <input
        :name="attributes.name"
        :value="attributes.value"
        type="hidden"
    />
    <v-btn
        @click="(event) => emitSubmitEvent(event, id)"
        :loading="!submitEnabled"
        depressed
        block
        class="font-weight-bold mt-2" color="primary darken-1"
        :id="id"
        type="submit">{{ metaLabel }}</v-btn>
  </div>
</template>

<script>
import {
  mdiLockOutline, mdiEmail, mdiAccount, mdiEye, mdiEyeOff,
} from '@mdi/js';

export default {
  name: 'OryUiNodeInput',
  props: {
    meta: { type: Object, required: true },
    attributes: { type: Object, required: true },
    id: { type: String, required: true },
    submitEnabled: { type: Boolean, require: false, default: true },
  },
  computed: {
    metaLabel() {
      if (this.meta.label && this.meta.label.text) {
        return this.meta.label.text;
      }
      return '';
    },
  },
  data() {
    return {
      icons: {
        mdiLockOutline, mdiEmail, mdiAccount, mdiEye, mdiEyeOff,
      },
      password: {
        show: false,
      },
      isLoading: false,
    };
  },
  methods: {
    emitSubmitEvent(event, btnID) {
      event.preventDefault();
      this.isLoading = true;
      this.$emit('submit', btnID);
    },
  },
};
</script>

<style scoped>

</style>
