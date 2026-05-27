<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="400px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>
        Delete All Custom {{ title }} Clusters
      </v-card-title>
      <v-card-text>
        <div class="text-body-large">
          Are you sure you want to delete all custom clusters?
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
          color="red"
          :loading="isLoading"
          @click="deleteAllClusters"
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
	title: {type: String, required: true},
	blockchainMode: {type: String, required: true},
});

const model = defineModel({type: Boolean});
const emit = defineEmits(['deleted']);
const dakar = getDakarClient(props.blockchainMode);
const isLoading = ref(false);
const errorMsg = ref('');
const msgType = ref('');

// Functions
function setPersistentErrorMessage(msg) {
	msgType.value = 'error';
	errorMsg.value = msg;
}

function setInfoMessage(msg) {
	msgType.value = 'info';
	errorMsg.value = msg;
}

async function deleteAllClusters() {
	isLoading.value = true;
	errorMsg.value = '';
	try {
		const response = await dakar.cluster.clustersDelete();
		if (response.msg) {
			setInfoMessage(response.msg);
		}

		emit('deleted');
	} catch (error) {
		setPersistentErrorMessage(error);
		return;
	} finally {
		isLoading.value = false;
	}

	model.value = false;
}

</script>

<style scoped>

</style>
