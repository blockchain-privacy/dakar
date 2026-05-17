<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="500px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>
        Delete {{ title }} Cluster
      </v-card-title>
      <v-card-text>
        <div class="text-body-large">
          Are you sure you want to delete this cluster?
          It is attached to <strong>{{ numAddresses }}</strong> addresses.
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
          @click="deleteCluster"
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
	clusterUid: {type: String, required: true},
	numAddresses: {type: Number, required: true},
	blockchainMode: {type: String, required: true},
	title: {type: String, required: true},
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

async function deleteCluster() {
	if (props.clusterUid === '' || props.numAddresses <= 0) {
		setErrorMessage('could not delete cluster');
		model.value = false;
		return;
	}

	errorMsg.value = '';
	isLoading.value = true;

	try {
		const response = await dakar.cluster.clustersUidDelete({uid: props.clusterUid});
		if (response.msg) {
			setInfoMessage(response.msg);
		}

		emit('deleted', props.clusterUid);
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
