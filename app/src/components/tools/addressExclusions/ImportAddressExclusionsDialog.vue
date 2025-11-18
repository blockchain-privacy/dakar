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
        <span class="text-h5">Import {{ title }} Address Exclusions</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1">
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
import {ref} from 'vue';
import {useRoute} from 'vue-router';
import {fileRule, getDakarClient} from '@/utilities';
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
});

// Functions
function setSuccessMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'success', temporary: true, category: route.name,
	});
}

function setInfoMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'info', temporary: true, category: route.name,
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

	try {
		const response = await dakar.addressExclusion.exclusionsPost({file: csv.value.file});
		if (response.msg) {
			setInfoMessage(response.msg);
		}

		setSuccessMessage('import was successful');
		emit('added');
	} catch (e) {
		setPersistentErrorMessage(codeToMsg(e.message));
	}

	isLoading.value = false;
	csv.value.file = undefined;
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
			return `file has more than ${Number(10000).toLocaleString()} addresses`;
		case 'file_error_importing':
			return 'error importing file';
		default:
			return msgCode;
	}
}

</script>

<style scoped>

</style>
