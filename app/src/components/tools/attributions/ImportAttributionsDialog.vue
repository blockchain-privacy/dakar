<template>
  <v-dialog
    v-model="model"
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
            <v-expansion-panel-text style="overflow: auto">
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
            :rules="fileRule"
            show-size
            accept="text/csv"
            label="Click here to select a file"
            truncate-length="15"
          />
          <div class="d-inline-flex align-center flex-wrap">
            <v-checkbox
              v-model="csv.firstRowContainsHeader"
              label="First row of file contains headers"
              :disabled="isLoading"
              class="me-2"
            />
            <v-checkbox
              v-if="isAdmin"
              v-model="areAttributionsPublic"
              label="Public attributions"
              :disabled="isLoading"
              class="me-2"
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
import {fileRule, getDakarClient, isAdminIdentity} from '@/utilities';
import {computed, ref} from 'vue';
import {useRoute} from 'vue-router';
import {useLocalStore} from '@/pinia/local';
import {useMsgStore} from '@/pinia/msg';
import {storeToRefs} from 'pinia';

const {getSettings} = storeToRefs(useLocalStore());
const route = useRoute();
const localStore = useLocalStore();
const msgStore = useMsgStore();
const dakar = getDakarClient(getSettings.value.blockchainMode);

const model = defineModel({type: Boolean});
const emit = defineEmits(['added']);

// Template ref
const csvForm = ref(null);
const isLoading = ref(false);
const areAttributionsPublic = ref(false);
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

// Computed
const isAdmin = computed(() => isAdminIdentity(localStore.getSession, localStore.getSettings.blockchainMode));

// Functions
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

function setSuccessMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'success', temporary: true, category: route.name,
	});
}

function setPersistentErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: false, category: route.name,
	});
}

async function handleCSVUpload() {
	const {valid} = await csvForm.value.validate();
	if (!valid) {
		return;
	}

	isLoading.value = true;
	const attributionData = {
		separator: csv.value.separator,
		hasHeader: csv.value.firstRowContainsHeader,
		file: csv.value.file[0],
	};

	try {
		if (areAttributionsPublic.value) {
			await dakar.attribution.attributionsPublicPost(attributionData);
		} else {
			await dakar.attribution.attributionsPost(attributionData);
		}

		setSuccessMessage('import was successful');
		emit('added');
	} catch (e) {
		setPersistentErrorMessage(codeToMsg(e.message));
	}

	isLoading.value = false;
	csv.value.file = null;
	model.value = false;
}
</script>

<style scoped>

</style>
