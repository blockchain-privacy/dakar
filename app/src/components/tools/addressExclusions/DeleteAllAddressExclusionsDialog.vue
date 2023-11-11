<template>
  <v-dialog
    v-model="show"
    max-width="400px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>
        <span class="text-h5">Delete All Address Exclusions</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1">
          Are you sure you want to delete all {{ count }} address exclusions?
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
              color="red"
              :loading="isLoading"
              @click="deleteAllAddressExclusions"
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
	name: 'DeleteAllAddressExclusionsDialog',
	props: {
		modelValue: {type: Boolean, required: true},
		count: {type: Number, required: true},
	},
	emits: ['deleted', 'update:modelValue'],
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
		async deleteAllAddressExclusions() {
			this.isLoading = true;

			try {
				await this.dakar.addressExclusion.deleteAllAddressExclusionsGet();
				this.$emit('deleted');
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
