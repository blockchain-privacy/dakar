<template>
  <v-dialog
    v-model="show"
    max-width="400px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>
        <span class="text-h5">Delete Address Exclusion</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1 text-break">
          Are you sure you want to delete the address <code>{{ addressHash }}</code>
          from the address exclusion list?
        </div>
        <v-row class="mt-4">
          <v-col class="d-flex justify-end align-center">
            <v-btn
              variant="text"
              :disabled="isLoading"
              @click="show = false"
            >
              Cancel
            </v-btn>
            <v-btn
              variant="text"
              :loading="isLoading"
              color="red"
              @click="deleteAddressExclusion"
            >
              Delete
            </v-btn>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
	name: 'DeleteAddressExclusionDialog',
	props: {
		modelValue: {type: Boolean, required: true},
		addressHash: {type: String, required: true},
	},
	emits: ['update:modelValue', 'deleted'],
	data() {
		return {
			isLoading: false,
		};
	},
	computed: {
		show: {
			get() {
				return this.modelValue;
			},
			set(value) {
				this.$emit('update:modelValue', value);
			},
		},
	},
	methods: {
		setPersistentErrorMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'error', temporary: false, category: this.$route.name});
		},
		async deleteAddressExclusion() {
			if (this.addressHash === '') {
				this.setPersistentErrorMessage('could not delete address exclusion');
				this.show = false;
				return;
			}

			this.isLoading = true;

			try {
				await this.dakar.addressExclusion.deleteAddressExclusionAddressHashGet({addressHash: this.addressHash});
				this.$emit('deleted', this.addressHash);
			} catch (e) {
				this.setPersistentErrorMessage(e);
			}

			this.isLoading = false;
			this.show = false;
		},
	},
};
</script>

<style scoped>

</style>
