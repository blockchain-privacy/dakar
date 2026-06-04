<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="500px"
  >
    <v-card>
      <v-card-title>New Workspace</v-card-title>
      <v-card-text>
        <v-tabs
          v-model="tab"
          align-tabs="center"
        >
          <v-tab value="new">
            New
          </v-tab>
          <v-tab value="file">
            Import
          </v-tab>
        </v-tabs>
        <v-window
          v-model="tab"
          class="mt-3"
        >
          <v-window-item value="new">
            <v-text-field
              v-model="note"
              class="mt-1"
              label="Workspace name"
              counter
              :maxlength="maxlength"
              autofocus
              @keydown.enter="submit"
            />
            <p class="text-title-small mb-2">
              Blockchain
            </p>
            <v-btn-toggle
              v-model="modeSelection"
              mandatory
            >
              <v-btn
                v-for="chain in Object.values(BLOCKCHAIN_ATTRIBUTES)"
                :key="chain.mode"
                variant="text"
                :value="chain.mode"
              >
                <v-icon
                  start
                  :icon="chain.icon"
                  :color="chain.color"
                  size="x-large"
                />
                <span>{{ chain.title }}</span>
              </v-btn>
            </v-btn-toggle>
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
                accept="application/json"
                @update:model-value="handleFileChange"
              />
            </v-form>
          </v-window-item>
        </v-window>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn
          variant="text"
          @click="closeDialog"
        >
          Cancel
        </v-btn>
        <v-btn
          variant="text"
          @click="submit"
        >
          Create
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script setup>
import {onMounted, onUpdated, ref} from 'vue';
import {mdiFileUpload} from '@mdi/js';
import {BLOCKCHAIN_ATTRIBUTES} from '@/constants/index.js';
import {fileRule} from '@/utilities/index.js';

const model = defineModel({type: Boolean});
const props = defineProps({
	maxlength: {type: Number, required: true},
});
const emit = defineEmits(['added', 'imported']);

const note = ref('');
const tab = ref(null);
const file = ref(undefined);
const fileForm = ref(null);

let oldNote = '';
const modeSelection = ref(Object.values(BLOCKCHAIN_ATTRIBUTES)[0].mode);

onMounted(() => {
	if (props.inputValue) {
		note.value = props.inputValue;
		oldNote = props.inputValue;
	}
});

onUpdated(() => {
	if (props.inputValue && props.inputValue !== oldNote) {
		note.value = props.inputValue;
		oldNote = props.inputValue;
	}
});

// Functions
function closeDialog() {
	model.value = false;
}

function submit() {
	emit('added', note.value, modeSelection.value);
	model.value = false;
}

async function handleFileChange() {
	emit('imported', file.value);
}

</script>

<style scoped>

</style>
