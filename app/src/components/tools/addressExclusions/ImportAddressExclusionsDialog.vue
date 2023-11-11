<template>
  <v-dialog
    v-model="show"
    max-width="700px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>
        <span class="text-h5">Import Address Exclusions</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1">
          Import an address exclusion list, which consists of a list of address hashes,
          separated by new line characters. The file must <strong>not</strong> have a header.
          The file may contain at maximum {{ Number(10000).toLocaleString() }} addresses.
        </div>
        <v-expansion-panels>
          <v-expansion-panel elevation="0">
            <v-expansion-panel-title>
              Example file
            </v-expansion-panel-title>
            <v-expansion-panel-text style="overflow: scroll">
              <p>The following file content would add 3 addresses to the address exclusion list.</p>
              <pre style="width: 200px"><code>Xf36MqBkoK8G5wBbjUSwDRy6XTjdNq8hgB
XatWuw7BhTxHvjPLbnvPArWgW9r6hjpt8o
XcsCPgY67TqW9CpsJLCbizDw2Yq2zFoh74</code></pre>
            </v-expansion-panel-text>
          </v-expansion-panel>
        </v-expansion-panels>
        <v-form
          id="csvForm"
          ref="csvForm"
          class="mt-3"
        >
          <v-file-input
            v-model="csv.file"
            :rules="rules.file"
            show-size
            accept="text/csv,text/plain"
            label="Click here to select a file"
            truncate-length="15"
          />
          <div class="d-flex align-center justify-end">
            <v-btn
              variant="text"
              :disabled="isLoading"
              class="mr-2"
              @click="show = false"
            >
              Cancel
            </v-btn>
            <v-btn
              variant="text"
              :loading="isLoading"
              @click="handleCSVUpload"
            >
              Upload
            </v-btn>
          </div>
        </v-form>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
// CodeToMsg returns a message for the given message code
function codeToMsg(msgCode) {
	switch (msgCode) {
		case 'file_invalid_field_count':
			return 'file must have one column';
		case 'file_no_data':
			return 'file does not contain data';
		case 'file_invalid_data':
			return 'file contains invalid data';
		case 'file_reading_error':
			return 'could not read file';
		case 'file_too_many_addresses':
			return `file has more than ${Number(10000).toLocaleString()} addresses`;
		case 'file_error_importing':
			return 'error importing file';
		default:
			return msgCode;
	}
}

export default {
	name: 'ImportAddressExclusionsDialog',
	props: {
		modelValue: {type: Boolean, required: true},
	},
	emits: ['added', 'update:modelValue'],
	data() {
		return {
			isLoading: false,
			csv: {
				valid: false,
				file: null,
			},
			rules: {
				file: [v => Boolean(v) || 'File is required'],
			},
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
		setSuccessMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'success', temporary: true, category: this.$route.name});
		},
		setInfoMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'info', temporary: true, category: this.$route.name});
		},
		setPersistentErrorMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'error', temporary: false, category: this.$route.name});
		},
		async handleCSVUpload() {
			const {valid} = await this.$refs.csvForm.validate();
			if (!valid) {
				return;
			}

			this.isLoading = true;

			try {
				const response = await this.dakar.addressExclusion.addAddressExclusionsPost({file: this.csv.file[0]});
				if (response.msg) {
					this.setInfoMessage(response.msg);
				}

				this.setSuccessMessage('import was successful');
				this.$emit('added');
			} catch (e) {
				this.setPersistentErrorMessage(codeToMsg(e.message));
			}

			this.isLoading = false;
			this.csv.file = null;
			this.show = false;
		},
	},
};
</script>

<style scoped>

</style>
