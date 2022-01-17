<template>
  <v-dialog v-model="show" max-width="700px">
    <v-card class="mx-auto elevation-4">
      <v-card-title>
        <span class="text-h5">Import Attributions</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1">
          Import address attributions by uploading a CSV-file.
          The file must have two columns, where the first column contains an
          address hash and the second column the address tag.
        </div>
        <v-expansion-panels flat>
          <v-expansion-panel>
            <v-expansion-panel-header>
              Example CSV-file
            </v-expansion-panel-header>
            <v-expansion-panel-content>
              <p>The following file content would generate five attributions,
                with one address having two tags.</p>
              <pre><code>address;tag
XgfSvDijDxPyWGXUw6CAxe91iYzZDMe3CV;darknet-address
XooBLwqL5wbBjoHJ1D4iZyrHWSRKQeRms9;twitter-@josh
XeNFLcypT3ayuqjVzK5HnzfRMBxwuBVKfB;facebook-some-user-name
XbpGcNSKaLnbfS9hPPa3yoE1boNqd3Ytij;case-123
XbpGcNSKaLnbfS9hPPa3yoE1boNqd3Ytij;exchange-Bitfinex</code></pre>
            </v-expansion-panel-content>
          </v-expansion-panel>
        </v-expansion-panels>
        <v-form ref="csvForm" id="csvForm">
          <v-file-input
              v-model="csv.file"
              :rules="rules.file"
              show-size
              accept="text/csv"
              label="Click here to select a file"
              truncate-length="15"/>
          <v-row>
            <v-col>
              <v-switch v-model="csv.firstRowContainsHeader"
                        label="First row of file contains headers" :disabled="isLoading"/>
            </v-col>
            <v-col>
              <v-select
                  v-model="csv.separator"
                  :items="separatorItems"
                  item-text="text"
                  item-value="value"
                  label="Separator">
              </v-select>
            </v-col>
            <v-col class="d-flex justify-end align-center">
              <v-btn text :disabled="isLoading" class="mr-2" @click="show = false">
                Cancel
              </v-btn>
              <v-btn text :loading="isLoading" @click="handleCSVUpload">
                Upload
              </v-btn>
            </v-col>
          </v-row>
        </v-form>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
import { mdiFileDownloadOutline } from '@mdi/js';
import { ROUTE_ADD_ATTRIBUTION } from '../../constants';
import { doPostUpload } from '../../utilities';

// codeToMsg returns a message for the given message code
function codeToMsg(msgCode) {
  switch (msgCode) {
    case 'empty_header_flag':
      return 'header flag is not set';
    case 'unsupported_separator':
      return 'invalid column separator';
    case 'file_invalid_field_count':
      return 'file must have two columns';
    case 'file_no_data':
      return 'file does not contain data';
    case 'file_invalid_data':
      return 'file contains invalid data';
    case 'file_reading_error':
      return 'could not read file';
    case 'file_too_many_addresses':
      return 'file has more than 1000 addresses';
    case 'file_error_importing':
      return 'error importing file';
    default:
      return msgCode;
  }
}

export default {
  name: 'ImportAttribution',
  props: {
    value: { type: Boolean, required: true },
  },
  data() {
    return {
      icon: { mdiFileDownloadOutline },
      isLoading: false,
      separatorItems: [
        { text: 'Colon (,)', value: ',' },
        { text: 'Semicolon (;)', value: ';' },
      ],
      csv: {
        valid: false,
        file: null,
        separator: ',',
        firstRowContainsHeader: false,
      },
      rules: {
        file: [(v) => !!v || 'File is required'],
        separator: [
          (v) => !!v || 'Separator is required',
          (v) => (v && v.length <= 10) || 'Separator must not greater than 10 characters',
        ],
      },
    };
  },
  computed: {
    show: {
      get() {
        return this.value;
      },
      set(value) {
        this.$emit('input', value);
      },
    },
  },
  methods: {
    setSuccessMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'success', temporary: true });
    },
    setPersistentErrorMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'error', temporary: false });
    },
    handleCSVUpload() {
      if (!this.$refs.csvForm.validate()) return;

      this.isLoading = true;

      // create and fill form data object
      const newForm = new FormData();
      newForm.append('file', this.csv.file);
      newForm.append('separator', this.csv.separator);
      newForm.append('hasHeader', this.csv.firstRowContainsHeader ? '1' : '0');

      // upload to server
      doPostUpload(ROUTE_ADD_ATTRIBUTION, this.$router, this.$store, newForm)
        .then((response) => {
          if (!response.success) {
            let errorMsg;
            if (response.msg) {
              errorMsg = codeToMsg(response.msg);
            } else {
              errorMsg = 'error processing inputs';
            }

            this.setPersistentErrorMessage(errorMsg);
          } else {
            this.setSuccessMessage('import was successful');
            this.$emit('added');
          }
        })
        .finally(() => {
          this.isLoading = false;
          this.csv.file = null;
          this.show = false;
        });
    },
  },
};
</script>

<style scoped>

</style>
