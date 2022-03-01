<template>
  <v-dialog v-model="show" max-width="700px">
    <v-card class="mx-auto elevation-4">
      <v-card-title>
        <span class="text-h5">Import Address Exclusions</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1">
          Import an address exclusion list, which consists of a list of address hashes,
          separated by new line characters. The file must <strong>not</strong> have a header.
          The file may contain at maximum {{ Number(10000).toLocaleString() }} addresses.
        </div>
        <v-expansion-panels flat>
          <v-expansion-panel>
            <v-expansion-panel-header>
              Example file
            </v-expansion-panel-header>
            <v-expansion-panel-content>
              <p>The following file content would add 3 addresses to the address exclusion list.</p>
              <pre><code>Xf36MqBkoK8G5wBbjUSwDRy6XTjdNq8hgB
XatWuw7BhTxHvjPLbnvPArWgW9r6hjpt8o
XcsCPgY67TqW9CpsJLCbizDw2Yq2zFoh74</code></pre>
            </v-expansion-panel-content>
          </v-expansion-panel>
        </v-expansion-panels>
        <v-form ref="csvForm" id="csvForm">
          <v-file-input
              v-model="csv.file"
              :rules="rules.file"
              show-size
              accept="text/csv,text/plain"
              label="Click here to select a file"
              truncate-length="15"/>
          <v-row>
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
import { ROUTE_ADD_ADDRESS_EXCLUSION } from '../../constants';
import { doPostUpload } from '../../utilities';

// codeToMsg returns a message for the given message code
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
  name: 'ImportAddressExclusions',
  props: {
    value: { type: Boolean, required: true },
  },
  data() {
    return {
      isLoading: false,
      csv: {
        valid: false,
        file: null,
      },
      rules: {
        file: [(v) => !!v || 'File is required'],
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

      // upload to server
      doPostUpload(ROUTE_ADD_ADDRESS_EXCLUSION, this.$router, this.$store, newForm)
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
