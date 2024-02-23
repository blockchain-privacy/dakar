<template>
  <input
    v-if="attributes.type === 'hidden'"
    :name="attributes.name"
    :value="attributes.value"
    type="hidden"
  >
  <v-text-field
    v-else-if="attributes.type === 'password'"
    :label="metaLabel"
    :model-value="attributes.value"
    :name="attributes.name"
    :prepend-inner-icon="mdiLockOutline"
    :type="password.show ? 'text' : 'password'"
    :append-inner-icon="password.show ? mdiEye : mdiEyeOff"
    autocomplete="current-password"
    @click:append-inner="password.show = !password.show"
  />
  <div
    v-else-if="attributes?.type === 'text' && attributes?.name === 'totp_code'"
    class="d-flex flex-column align-center"
  >
    <v-label :text="metaLabel" />
    <!-- Use a random number for 'key', so the content always gets
    cleared when it is replaced by the result of a call to Kratos.
    Otherwise the finish event gets called on each field update if the input is full -->
    <v-otp-input
      :key="Math.random()"
      :model-value="attributes.value?attributes.value:''"
      :autofocus="true"
      :name="attributes.name"
      @finish="emitSubmitEvent(null, 'otp-input')"
    />
  </div>
  <v-text-field
    v-else-if="attributes.type === 'text' || (attributes.type === 'text'&& metaLabel === 'E-Mail')"
    :key="meta?.label?.id"
    :label="metaLabel"
    :model-value="attributes.value?attributes.value:''"
    :name="attributes.name"
    :prepend-inner-icon="inputIcon"
  />
  <v-text-field
    v-else-if="attributes.type === 'email'"
    :label="metaLabel"
    :model-value="attributes.value"
    :name="attributes.name"
    type="email"
    :prepend-inner-icon="mdiEmail"
  />
  <div v-else-if="attributes.type === 'submit'">
    <input
      :name="attributes.name"
      :value="attributes.value"
      type="hidden"
    >
    <v-btn
      :name="name"
      :loading="!submitEnabled"
      variant="flat"
      :block="true"
      class="font-weight-bold mt-2"
      color="primary-darken-1"
      type="submit"
      @click="(event) => emitSubmitEvent(event, name)"
    >
      {{ metaLabel }}
    </v-btn>
  </div>
</template>

<script setup>
import {
	mdiLockOutline, mdiEmail, mdiAccount, mdiEye, mdiEyeOff, mdiFormTextboxPassword,
} from '@mdi/js';
import {computed, nextTick, ref} from 'vue';

const props = defineProps({
	meta: {type: Object, required: true},
	attributes: {type: Object, required: true},
	name: {type: String, required: true},
	submitEnabled: {type: Boolean, require: false, default: true},
});

const emit = defineEmits(['submit']);

const password = ref({show: false});

// Computed
const metaLabel = computed(() => {
	if (props.meta.label?.text) {
		return props.meta.label.text;
	}

	return '';
});

const inputIcon = computed(() => {
	if (props.attributes?.name === 'totp_code') {
		return mdiFormTextboxPassword;
	}

	if (props.attributes?.name === 'code') {
		return mdiFormTextboxPassword;
	}

	if (metaLabel.value === 'E-Mail') {
		return mdiEmail;
	}

	if (props.attributes?.name === 'identifier') {
		return mdiAccount;
	}

	return null;
});

// Functions
function emitSubmitEvent(event, btnName) {
	if (event) {
		event.preventDefault();
	}

	// NextTick is needed so the complete otp-input is getting submitted
	nextTick(() => {
		emit('submit', btnName);
	});
}

</script>

<style scoped>

</style>
