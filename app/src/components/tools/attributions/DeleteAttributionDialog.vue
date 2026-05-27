<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="500px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>Delete {{ title }} Attribution</v-card-title>
      <v-card-text>
        <div class="text-body-large">
          Are you sure you want to delete the attribution <code>{{ tag }}</code>?
        </div>
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
          color="red"
          @click="deleteAttribution"
        >
          Delete
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {ref} from 'vue';
import {getDakarClient} from '@/utilities/index.js';
import Alert from '@/components/common/Alert.vue';

const props = defineProps({
	attributionUid: {type: String, required: true},
	tag: {type: String, required: true},
	public: {type: Boolean, required: true},
	title: {type: String, required: true},
	blockchainMode: {type: String, required: true},
});

const dakar = getDakarClient(props.blockchainMode);

const model = defineModel({type: Boolean});
const emit = defineEmits(['deleted']);

const isLoading = ref(false);
const errorMsg = ref('');
const msgType = ref('');

// Functions
function setErrorMessage(msg) {
	msgType.value = 'error';
	errorMsg.value = msg;
}

function setInfoMessage(msg) {
	msgType.value = 'info';
	errorMsg.value = msg;
}

async function deleteAttribution() {
	if (props.attributionUid === '') {
		setErrorMessage('could not delete attribution');
		model.value = false;
		return;
	}

	isLoading.value = true;
	errorMsg.value = '';
	try {
		const response = props.public
			? await dakar.attribution.attributionsPublicUidDelete({uid: props.attributionUid})
			: await dakar.attribution.attributionsUidDelete({uid: props.attributionUid});

		if (response.msg) {
			setInfoMessage(response.msg);
		}

		emit('deleted', props.attributionUid);
	} catch (error) {
		setErrorMessage(error);
		return;
	} finally {
		isLoading.value = false;
	}

	model.value = false;
}
</script>

<style scoped>

</style>
