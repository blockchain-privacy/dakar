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
        <span class="text-h5">Delete All Custom {{ title }} Clusters</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1">
          Are you sure you want to delete all custom clusters?
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
              color="red"
              :loading="isLoading"
              @click="deleteAllClusters"
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

const props = defineProps({
	title: {type: String, required: true},
	blockchainMode: {type: String, required: true},
});

const model = defineModel({type: Boolean});
const emit = defineEmits(['deleted']);
const dakar = getDakarClient(props.blockchainMode);
const isLoading = ref(false);

// Functions
function setPersistentErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: false, category: route.name,
	});
}

function setInfoMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'info', temporary: true, category: route.name,
	});
}

async function deleteAllClusters() {
	isLoading.value = true;

	try {
		const response = await dakar.cluster.clustersDelete();
		if (response.msg) {
			setInfoMessage(response.msg);
		}

		emit('deleted');
	} catch (e) {
		setPersistentErrorMessage(e);
	}

	isLoading.value = false;
	model.value = false;
}

</script>

<style scoped>

</style>
