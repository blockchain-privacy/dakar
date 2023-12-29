<template>
  <v-dialog
    v-model="model"
    max-width="700px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>
        <span class="text-h5">Import Clusters</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1">
          Import address clusters by uploading a CSV-file.
          The file must have two columns, where the first column contains an
          identifier for each cluster and the second column the addresses.
          The file may contain at maximum {{ Number(1000).toLocaleString() }} clusters.
        </div>
        <v-expansion-panels>
          <v-expansion-panel elevation="0">
            <v-expansion-panel-title>
              Example CSV-file
            </v-expansion-panel-title>
            <v-expansion-panel-text style="overflow: auto">
              <p>The following file content would generate two clusters with two addresses each.</p>
              <pre style="width: 200px"><code>cluster-id,address
1,XgG6Nosmei5woQ2VTDzwmLX7SzdNYKHdiz
1,Xf36MqBkoK8G5wBbjUSwDRy6XTjdNq8hgB
2,XatWuw7BhTxHvjPLbnvPArWgW9r6hjpt8o
2,XcsCPgY67TqW9CpsJLCbizDw2Yq2zFoh74</code></pre>
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
            :rules="fileRule"
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
              @click="model = false"
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

<script setup>
import {inject, ref} from 'vue';
import {useRoute} from 'vue-router';
import {fileRule} from '@/utilities';
import {useMsgStore} from '@/pinia/msg';

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();

const model = defineModel({type: Boolean});
const emit = defineEmits(['added']);

// CsvForm is a template ref
const csvForm = ref(null);
const isLoading = ref(false);
const csv = ref({
	valid: false,
	file: null,
	separator: ',',
	firstRowContainsHeader: false,
});

const separatorItems = [
	{text: 'Colon (,)', value: ','},
	{text: 'Semicolon (;)', value: ';'},
];

// Functions
function setSuccessMessage(msg) {
	msgStore.addMessage({text: msg, type: 'success', temporary: true, category: route.name});
}

function setPersistentErrorMessage(msg) {
	msgStore.addMessage({text: msg, type: 'error', temporary: false, category: route.name});
}

async function handleCSVUpload() {
	const {valid} = await csvForm.value.validate();
	if (!valid) {
		return;
	}

	isLoading.value = true;

	try {
		await dakar.cluster.addClusterPost({
			separator: csv.value.separator,
			hasHeader: csv.value.firstRowContainsHeader,
			file: csv.value.file[0],
		});

		setSuccessMessage('import was successful');
		emit('added');
	} catch (e) {
		setPersistentErrorMessage(codeToMsg(e.message));
	}

	isLoading.value = false;
	csv.value.file = null;
	model.value = false;
}

// CodeToMsg returns a message for the given message code
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
			return `file has more than ${Number(1000).toLocaleString()} clusters`;
		case 'file_shallow_cluster':
			return 'file contains clusters with only one address';
		case 'file_error_importing':
			return 'error importing file';
		default:
			return msgCode;
	}
}

</script>

<style scoped>

</style>
