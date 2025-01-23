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
    :error-messages="errorMessages"
    :messages="infoMessages"
    @click:append-inner="password.show = !password.show"
  />
  <!--    Use a random number for 'key', so the content always gets-->
  <!--    cleared when it is replaced by the result of a call to Kratos.-->
  <!--    Otherwise the finish event gets called on each field update if the input is full -->
  <v-text-field
    v-else-if="attributes?.type === 'text' && attributes?.name === 'totp_code'"
    :key="Math.random()"
    :label="metaLabel"
    :model-value="attributes.value?attributes.value:''"
    :name="attributes.name"
    :prepend-inner-icon="inputIcon"
    :error-messages="errorMessages"
    :messages="infoMessages"
    autocomplete="one-time-code"
  />
  <v-text-field
    v-else-if="attributes.type === 'text' || (attributes.type === 'text'&& metaLabel === 'E-Mail')"
    :key="meta?.label?.id"
    :label="metaLabel"
    :model-value="attributes.value?attributes.value:''"
    :name="attributes.name"
    :error-messages="errorMessages"
    :messages="infoMessages"
    :prepend-inner-icon="inputIcon"
  />
  <v-text-field
    v-else-if="attributes.type === 'email'"
    :label="metaLabel"
    :model-value="attributes.value"
    :name="attributes.name"
    type="email"
    :error-messages="errorMessages"
    :messages="infoMessages"
    :prepend-inner-icon="mdiEmail"
  />
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

	// Eval is bad practice, but this the offical way to call ory kratos script
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
