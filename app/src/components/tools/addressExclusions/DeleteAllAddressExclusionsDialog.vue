<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="400px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>Delete All {{ title }} Address Exclusions</v-card-title>
      <v-card-text>
        <div class="text-body-large">
          Are you sure you want to delete all {{ count }} address exclusions?
        </div>
      </v-card-text>
      <alert :text="errorMsg" />
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
          @click="deleteAllAddressExclusions"
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

const model = defineModel({type: Boolean});
const props = defineProps({
	count: {type: Number, required: true},
	title: {type: String, required: true},
	blockchainMode: {type: String, required: true},
});
const emit = defineEmits(['deleted']);
const dakar = getDakarClient(props.blockchainMode);

const isLoading = ref(false);
const errorMsg = ref('');

// Functions
async function deleteAllAddressExclusions() {
	isLoading.value = true;
	errorMsg.value = '';

	try {
		await dakar.addressExclusion.exclusionsDelete();
		emit('deleted');
	} catch (error) {
		errorMsg.value = error.message;
		return;
	} finally {
		isLoading.value = false;
	}

	model.value = false;
}

</script>

<style scoped>

</style>
