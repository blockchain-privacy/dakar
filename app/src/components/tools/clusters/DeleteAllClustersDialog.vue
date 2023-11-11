<template>
  <v-dialog
    v-model="show"
    max-width="400px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>
        <span class="text-h5">Delete All Custom Clusters</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1">
          Are you sure you want to delete all custom clusters?
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
              @click="deleteAllClusters"
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
	name: 'DeleteAllClustersDialog',
	props: {
		modelValue: {type: Boolean, required: true},
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
		setInfoMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'info', temporary: true, category: this.$route.name});
		},
		async deleteAllClusters() {
			this.isLoading = true;

			try {
				const response = await this.dakar.cluster.deleteAllClustersGet();
				if (response.msg) {
					this.setInfoMessage(response.msg);
				}

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
