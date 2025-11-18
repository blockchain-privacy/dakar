<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <input
    v-if="attributes.type === 'hidden'"
    :name="attributes.name"
    :value="attributes.value"
    type="hidden"
  >
  <div v-else-if="attributes.type === 'submit'">
    <input
      :name="attributes.name"
      :value="attributes.value"
      type="hidden"
    >
    <!-- workaround for: https://github.com/ory/kratos/issues/2504 -->
    <template v-if="name === 'totp_unlink'">
      <input
        name="method"
        value="totp"
        type="hidden"
      >
    </template>
    <template v-else-if="name === 'webauthn_remove'">
      <input
        name="method"
        value="webauthn"
        type="hidden"
      >
    </template>
    <v-btn
      :name="name"
      :loading="!submitEnabled"
      variant="flat"
      block
      class="font-weight-bold mt-2"
      color="primary-darken-1"
      type="submit"
      @click="(event) => emitSubmitEvent(event, name)"
    >
      {{ metaLabel }}
    </v-btn>
  </div>
  <v-btn
    v-else-if="attributes.type === 'button' && attributes.onclick"
    @click="(event) => evalScript(event, attributes.onclick)"
  >
    {{ metaLabel }}
  </v-btn>
  <v-text-field
    v-else
    :key="meta?.label?.id"
    :label="metaLabel"
    :model-value="attributes.value"
    :name="attributes.name"
    :prepend-inner-icon="prependIcon"
    :type="inputType"
    :append-inner-icon="appendIcon"
    :autocomplete="autocomplete"
    :error-messages="errorMessages"
    :messages="infoMessages"
    @click:append-inner="appendClick"
  />
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
	submitEnabled: {type: Boolean, required: false},
	messages: {type: Array, required: false, default: () => []},
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

const prependIcon = computed(() => {
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

	if (props.attributes?.type === 'password') {
		return mdiLockOutline;
	}

	return null;
});

const appendIcon = computed(() => {
	if (props.attributes?.type === 'password') {
		return password.value.show ? mdiEye : mdiEyeOff;
	}

	return null;
});

const autocomplete = computed(() => {
	if (props.attributes?.name === 'totp_code') {
		return 'one-time-code';
	}

	if (props.attributes?.type === 'password') {
		return 'current-password';
	}

	return null;
});

const inputType = computed(() => {
	if (props.attributes?.type === 'password') {
		return password.value.show ? 'text' : 'password';
	}

	return null;
});

const errorMessages = computed(() => {
	if (!props.messages || props.messages.length === 0) {
		return props.messages;
	}

	return props.messages.filter(msg => msg.type === 'error').map(msg => msg.text);
});

const infoMessages = computed(() => {
	if (!props.messages || props.messages.length === 0) {
		return props.messages;
	}

	return props.messages.filter(msg => msg.type !== 'error').map(msg => msg.text);
});

const appendClick = computed(() => {
	if (props.attributes?.type === 'password') {
		return () => {
			password.value.show = !password.value.show;
		};
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

function evalScript(event, script) {
	event.preventDefault();
	event.stopPropagation();

	// Eval is bad practice, but this the official way to call an ory kratos script
	// eslint-disable-next-line no-new-func
	const evalScript = new Function(script);
	evalScript();
}

</script>

<style scoped>

/* set opacity because some password managers (keepassxc) can not find inputs */
/* it seems vuetify sets the opacity of inputs to zero when they are not focused */
:deep(.v-field__input) {
  opacity: 1;
}

</style>
