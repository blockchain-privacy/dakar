<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="700px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>
        Import {{ title }} Address Exclusions
      </v-card-title>
      <v-card-text>
        <div class="text-body-large">
          Import an address exclusion list, which consists of a list of address hashes,
          separated by new line characters. The file must <strong>not</strong> have a header.
          The file may contain at maximum {{ Number(10000).toLocaleString() }} addresses.
          The <wiki-tooltip description-url="addressExclusions.md">
            address exclusions
          </wiki-tooltip> wiki page shows an example.
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
            accept="text/csv,text/plain"
            label="Click here to select a file"
            truncate-length="15"
          />
        </v-form>
      </v-card-text>
      <alert
        :text="errorMsg"
        :type="msgType"
      />
      <v-card-actions>
        <v-btn
          variant="text"
          :disabled="isLoading"
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
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {ref} from 'vue';
import {fileRule, getDakarClient} from '@/utilities';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';
import Alert from '@/components/common/Alert.vue';

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
});
const errorMsg = ref('');
const msgType = ref('');

// Functions
function setInfoMessage(msg) {
	msgType.value = 'info';
	errorMsg.value = msg;
}

function setErrorMessage(msg) {
	msgType.value = 'error';
	errorMsg.value = msg;
}

async function handleCSVUpload() {
	const {valid} = await csvForm.value.validate();
	if (!valid) {
		return;
	}

	isLoading.value = true;
	errorMsg.value = '';

	try {
		const response = await dakar.addressExclusion.exclusionsPost({file: csv.value.file});
		if (response.msg) {
			setInfoMessage(response.msg);
		}

		emit('added');
	} catch (error) {
		setErrorMessage(codeToMsg(error.message));
		return;
	} finally {
		isLoading.value = false;
		csv.value.file = undefined;
	}

	model.value = false;
}

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
			return `file has more than ${Number(10_000).toLocaleString()} addresses`;
		case 'file_error_importing':
			return 'error importing file';
		default:
			return msgCode;
	}
}

</script>

<style scoped>

</style>
