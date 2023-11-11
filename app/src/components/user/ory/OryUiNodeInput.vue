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
    :prepend-inner-icon="icons.mdiLockOutline"
    :type="password.show ? 'text' : 'password'"
    :append-inner-icon="password.show ? icons.mdiEye : icons.mdiEyeOff"
    autocomplete="current-password"
    @click:append-inner="password.show = !password.show"
  />
  <!--  todo: enable if v-otp-input is usable on mobile -->
  <!--  <div-->
  <!--    v-else-if="attributes?.type === 'text' && attributes?.name === 'totp_code'"-->
  <!--    class="d-flex flex-column align-center"-->
  <!--  >-->
  <!--    <v-label :text="metaLabel" />-->
  <!--    &lt;!&ndash; Use a random number for key so the content always gets cleared on field update.-->
  <!--     Otherwise the finish event gets called on each field update if the input is full &ndash;&gt;-->
  <!--    <v-otp-input-->
  <!--      :key="Math.random()"-->
  <!--      autofocus-->
  <!--      :model-value="attributes.value?attributes.value:''"-->
  <!--      :name="attributes.name"-->
  <!--      @finish="emitSubmitEvent(null, id)"-->
  <!--    />-->
  <!--  </div>-->
  <v-text-field
    v-else-if="attributes.type === 'text'"
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
    :prepend-inner-icon="icons.mdiEmail"
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

<script>
import {
	mdiLockOutline, mdiEmail, mdiAccount, mdiEye, mdiEyeOff, mdiFormTextboxPassword,
} from '@mdi/js';

export default {
	name: 'OryUiNodeInput',
	props: {
		meta: {type: Object, required: true},
		attributes: {type: Object, required: true},
		name: {type: String, required: true},
		submitEnabled: {type: Boolean, require: false, default: true},
	},
	emits: ['submit'],
	data() {
		return {
			icons: {
				mdiLockOutline, mdiEmail, mdiAccount, mdiEye, mdiEyeOff, mdiFormTextboxPassword,
			},
			password: {
				show: false,
			},
		};
	},
	computed: {
		metaLabel() {
			if (this.meta.label && this.meta.label.text) {
				return this.meta.label.text;
			}

			return '';
		},
		inputIcon() {
			if (this.attributes?.name === 'totp_code') {
				return this.icons.mdiFormTextboxPassword;
			}

			if (this.attributes?.name === 'code') {
				return this.icons.mdiFormTextboxPassword;
			}

			if (this.metaLabel === 'ID') {
				return this.icons.mdiAccount;
			}

			return null;
		},
	},
	methods: {
		emitSubmitEvent(event, btnName) {
			if (event) {
				event.preventDefault();
			}

			// NextTick is needed so the complete otp-input is getting submitted
			this.$nextTick(() => {
				this.$emit('submit', btnName);
			});
		},
	},
};
</script>

<style scoped>

</style>
