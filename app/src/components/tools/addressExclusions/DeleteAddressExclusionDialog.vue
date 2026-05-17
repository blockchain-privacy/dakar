<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="500px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>Delete Address Exclusion</v-card-title>
      <v-card-text>
        <div class="text-body-large text-break">
          Are you sure you want to delete the address <code>{{ addressHash }}</code>
          from the address exclusion list?
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
          :loading="isLoading"
          color="red"
          @click="deleteAddressExclusion"
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
	addressHash: {type: String, required: true},
	blockchainMode: {type: String, required: true},
});
const emit = defineEmits(['deleted']);
const dakar = getDakarClient(props.blockchainMode);
const isLoading = ref(false);
const errorMsg = ref('');

// Functions
async function deleteAddressExclusion() {
	if (props.addressHash === '') {
		errorMsg.value = 'could not delete address exclusion';
		model.value = false;
		return;
	}

	isLoading.value = true;
	errorMsg.value = '';

	try {
		await dakar.addressExclusion.exclusionsHashDelete({hash: props.addressHash});
		emit('deleted', props.addressHash);
	} catch (error) {
		errorMsg.value = error.message;
	}

	isLoading.value = false;
	model.value = false;
}

</script>

<style scoped>

</style>
