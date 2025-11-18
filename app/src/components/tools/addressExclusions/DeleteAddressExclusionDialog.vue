<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="400px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>
        <span class="text-h5">Delete {{ title }} Address Exclusion</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1 text-break">
          Are you sure you want to delete the address <code>{{ addressHash }}</code>
          from the address exclusion list?
        </div>
        <v-row class="mt-4">
          <v-col class="d-flex justify-end align-center">
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
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {ref} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import {getDakarClient} from '@/utilities/index.js';

const route = useRoute();
const msgStore = useMsgStore();

const model = defineModel({type: Boolean});
const props = defineProps({
	addressHash: {type: String, required: true},
	title: {type: String, required: true},
	blockchainMode: {type: String, required: true},
});
const emit = defineEmits(['deleted']);
const dakar = getDakarClient(props.blockchainMode);
const isLoading = ref(false);

// Functions
function setPersistentErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: false, category: route.name,
	});
}

async function deleteAddressExclusion() {
	if (props.addressHash === '') {
		setPersistentErrorMessage('could not delete address exclusion');
		model.value = false;
		return;
	}

	isLoading.value = true;

	try {
		await dakar.addressExclusion.exclusionsHashDelete({hash: props.addressHash});
		emit('deleted', props.addressHash);
	} catch (e) {
		setPersistentErrorMessage(e);
	}

	isLoading.value = false;
	model.value = false;
}

</script>

<style scoped>

</style>
