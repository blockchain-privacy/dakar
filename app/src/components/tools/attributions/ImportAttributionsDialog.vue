<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="1000px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>
        <span class="text-h5">Import {{ title }} Attributions</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1">
          Import address attributions by uploading a CSV-file.
          The file must have five columns (<code>address</code>,
          <code>tag</code>,<code>description</code>,<code>source</code> and
          <code>category</code>). The fields <code>address</code>
          and <code>tag</code> are mandatory, the rest are optional.
          The file may contain at maximum {{ Number(1000).toLocaleString() }} attributions.

          The <wiki-tooltip description-url="attributions.md">
            attribution
          </wiki-tooltip> wiki page shows a CSV-file example.
        </div>
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
import {fileRule, getDakarClient} from '@/utilities';
import {ref} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';

const route = useRoute();
const msgStore = useMsgStore();

const props = defineProps({
	title: {type: String, required: true},
	blockchainMode: {type: String, required: true},
});

const dakar = getDakarClient(props.blockchainMode);

const model = defineModel({type: Boolean});
const emit = defineEmits(['added']);

// Template ref
const csvForm = ref(null);
const isLoading = ref(false);
const csv = ref({
	valid: false,
	file: undefined,
	separator: ',',
	firstRowContainsHeader: false,
});

const separatorItems = [
	{text: 'Colon (,)', value: ','},
	{text: 'Semicolon (;)', value: ';'},
];

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
		file: csv.value.file,
	};

	try {
		await dakar.attribution.attributionsPost(attributionData);

		setSuccessMessage('import was successful');
		emit('added');
	} catch (e) {
		setPersistentErrorMessage(codeToMsg(e.message));
	}

	isLoading.value = false;
	csv.value.file = undefined;
	model.value = false;
}
</script>

<style scoped>

</style>
