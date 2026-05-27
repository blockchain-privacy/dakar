<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="600"
  >
    <v-card
      title="Add Entities"
      :prepend-icon="mdiPlus"
    >
      <v-card-text>
        <alert :text="errorMsg" />
        <v-tabs
          v-model="tab"
          align-tabs="center"
        >
          <v-tab value="query">
            Query
          </v-tab>
          <v-tab value="file">
            File
          </v-tab>
        </v-tabs>
        <v-window
          v-model="tab"
          class="mt-3"
        >
          <v-window-item value="query">
            <v-form
              id="queryForm"
              ref="queryForm"
              validate-on="submit"
            >
              <p class="text-body-large">
                Query for multiple entities by separating them by any special character.
              </p>
              <v-text-field
                v-model="graphQuery"
                class="mt-4"
                autofocus
                variant="outlined"
                density="compact"
                color="primary"
                :rules="inputRules"
                label="Add transactions or address clusters"
                :disabled="!addEntityEnabled"
                :append-inner-icon="mdiMagnify"
                @click:append-inner="onAddEntities(tab)"
                @keydown.enter="onAddEntities(tab)"
              />
            </v-form>
          </v-window-item>
          <v-window-item value="file">
            <v-form
              id="fileForm"
              ref="fileForm"
            >
              <v-file-upload
                v-model="file"
                density="compact"
                :icon="mdiFileUpload"
                :rules="fileRule"
                title="Choose or drag and drop a file here"
                show-size
                accept="text/csv,text/plain"
                @update:model-value="handleFileChange"
              />
            </v-form>
          </v-window-item>
        </v-window>
        <v-expand-transition>
          <div
            v-if="queryItemCount > 1"
            class="d-flex justify-center"
          >
            <v-btn
              variant="text"
              @click="showDetectedEntities = !showDetectedEntities"
            >
              {{ showDetectedEntities?'Hide':'Show' }} detected entities
            </v-btn>
          </div>
        </v-expand-transition>
        <v-expand-transition>
          <div v-if="queryItemCount > 1 && showDetectedEntities">
            <v-list
              v-for="entity in detectedEntities"
              :key="entity"
              density="compact"
            >
              <v-list-item class="ma-0 pa-0">
                {{ entity }}
              </v-list-item>
            </v-list>
          </div>
        </v-expand-transition>
      </v-card-text>
      <v-card-actions>
        <v-btn
          class="ml-auto"
          text="Cancel"
          @click="model = false"
        />
        <v-btn
          :disabled="queryItemCount === 0"
          @click="onAddEntities(tab)"
        >
          Add {{ queryItemCount > 1?queryItemCount:'' }} {{ pluralIrregular('entity','entities', queryItemCount) }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {
	mdiFileUpload,
	mdiMagnify,
	mdiPlus,
} from '@mdi/js';
import {ref, computed} from 'vue';
import {VFileUpload} from 'vuetify/labs/VFileUpload';
import {extractEntities, fileRule, pluralIrregular} from '@/utilities/index.js';
import Alert from '@/components/common/Alert.vue';

const model = defineModel({type: Boolean});
const emit = defineEmits(['addEntities']);

defineProps({
	addEntityEnabled: {type: Boolean, required: true},
});

const graphQuery = ref('');
const queryForm = ref(null);
const fileForm = ref(null);
const tab = ref(null);
const showDetectedEntities = ref(false);
const file = ref(undefined);
const extractedFileContent = ref([]);
const errorMsg = ref('');

const inputRules = [q => extractEntities(q).length > 0 || 'query contains no valid entities'];

// Computed
const detectedEntities = computed(() => {
	if (tab.value === 'query') {
		return extractEntities(graphQuery.value);
	}

	if (tab.value === 'file' && file.value) {
		return extractedFileContent.value;
	}

	return [];
});

const queryItemCount = computed(() => detectedEntities.value.length);

// Functions
async function onAddEntities(t) {
	if (t === 'query') {
		const {valid} = await queryForm.value.validate();
		if (!valid) {
			return;
		}
	} else if (t === 'file') {
		const {valid} = await fileForm.value.validate();
		if (!valid || !file.value) {
			return;
		}
	} else {
		return;
	}

	model.value = false;
	emit('addEntities', detectedEntities.value);
	graphQuery.value = '';
	file.value = undefined;
}

async function handleFileChange() {
	extractedFileContent.value = [];
	const fileBlob = new Blob([file.value]);
	try {
		const text = await fileBlob.text();
		extractedFileContent.value = extractEntities(text);
	} catch (error) {
		errorMsg.value = error.message;
		extractedFileContent.value = [];
		file.value = undefined;
	}
}
</script>

<style scoped>

</style>
