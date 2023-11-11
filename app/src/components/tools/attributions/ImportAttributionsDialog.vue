<template>
  <v-dialog
    v-model="show"
    max-width="1000px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>
        <span class="text-h5">Import Attributions</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1">
          Import address attributions by uploading a CSV-file.
          The file must have five columns (<code>address</code>,
          <code>tag</code>,<code>description</code>,<code>source</code> and
          <code>category</code>). The fields <code>address</code>
          and <code>tag</code> are mandatory, the rest are optional.
          The file may contain at maximum {{ Number(1000).toLocaleString() }} attributions.
        </div>
        <v-expansion-panels>
          <v-expansion-panel elevation="0">
            <v-expansion-panel-title>
              Example CSV-file
            </v-expansion-panel-title>
            <v-expansion-panel-text style="overflow: scroll">
              <p>
                The following file content would generate five attributions,
                with one address having two tags.
              </p>
              <pre style="width: 200px"><code>address;tag;description;source;category
XgfSvDijDxPyWGXUw6CAxe91iYzZDMe3CV;darknet-address;;;
XooBLwqL5wbBjoHJ1D4iZyrHWSRKQeRms9;twitter-@josh;Josh Noname;https://twitter.com/josh;social media
XeNFLcypT3ayuqjVzK5HnzfRMBxwuBVKfB;facebook-some-user-name;;;social media
XbpGcNSKaLnbfS9hPPa3yoE1boNqd3Ytij;case-123;;;
XbpGcNSKaLnbfS9hPPa3yoE1boNqd3Ytij;exchange-Bitfinex;;;</code></pre>
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
            accept="text/csv"
            label="Click here to select a file"
            truncate-length="15"
          />
          <div class="d-flex align-center flex-wrap">
            <v-checkbox
              v-model="csv.firstRowContainsHeader"
              label="First row of file contains headers"
              :disabled="isLoading"
            />
            <v-checkbox
              v-if="isAdmin"
              v-model="areAttributionsPublic"
              label="Public attributions"
              :disabled="isLoading"
            />
            <v-select
              v-model="csv.separator"
              :items="separatorItems"
              item-title="text"
              item-value="value"
              label="Separator"
            />
          </div>
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
import {isAdminIdentity} from '@/utilities';

// CodeToMsg returns a message for the given message code
function codeToMsg(msgCode) {
	switch (msgCode) {
		case 'empty_header_flag':
			return 'header flag is not set';
		case 'unsupported_separator':
			return 'invalid column separator';
		case 'file_invalid_field_count':
			return 'file must have five columns';
		case 'file_no_data':
			return 'file does not contain data';
		case 'file_invalid_data':
			return 'file contains invalid data';
		case 'file_reading_error':
			return 'could not read file';
		case 'file_too_many_addresses':
			return `file has more than ${Number(1000).toLocaleString()} attributions`;
		case 'file_error_importing':
			return 'error importing file';
		default:
			return msgCode;
	}
}

export default {
	name: 'ImportAttributionDialog',
	props: {
		modelValue: {type: Boolean, required: true},
	},
	emits: ['added', 'update:modelValue'],
	data() {
		return {
			isLoading: false,
			separatorItems: [
				{text: 'Colon (,)', value: ','},
				{text: 'Semicolon (;)', value: ';'},
			],
			areAttributionsPublic: false,
			csv: {
				valid: false,
				file: null,
				separator: ',',
				firstRowContainsHeader: false,
			},
			rules: {
				file: [v => Boolean(v) || 'File is required'],
				separator: [
					v => Boolean(v) || 'Separator is required',
					v => (v && v.length <= 10) || 'Separator must not greater than 10 characters',
				],
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
		isAdmin() {
			return isAdminIdentity(this.$store.getters.getSession);
		},
	},
	methods: {
		setSuccessMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'success', temporary: true, category: this.$route.name});
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
			const attributionData = {
				separator: this.csv.separator,
				hasHeader: this.csv.firstRowContainsHeader,
				file: this.csv.file[0],
			};

			try {
				if (this.areAttributionsPublic) {
					await this.dakar.attribution.addPublicAttributionPost(attributionData);
				} else {
					await this.dakar.attribution.addPrivateAttributionPost(attributionData);
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
