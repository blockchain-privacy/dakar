<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="500px"
  >
    <v-card>
      <v-card-title>Export Workspace '{{ workspace.name }}'</v-card-title>
      <v-card-text>
        <v-radio-group v-model="exportOption">
          <v-card
            variant="outlined"
            class="mb-2"
          >
            <v-radio value="workspace">
              <template #label>
                <v-card variant="flat">
                  <v-card-title>
                    Workspace Export
                  </v-card-title>
                  <v-card-text>
                    Exports the entire workspace. The exported file can be imported; when imported,
                    selectors will be replayed. Use this option to back up, duplicate, or share your
                    workspace with others.
                  </v-card-text>
                </v-card>
              </template>
            </v-radio>
          </v-card>
          <v-card variant="outlined">
            <v-radio value="entities">
              <template #label>
                <v-card variant="flat">
                  <v-card-title>
                    Entity Export
                  </v-card-title>
                  <v-card-text>
                    Exports transaction and address hashes. Selector results are excluded.
                    Use this option to make entities from this workspace available to other tools.
                  </v-card-text>
                </v-card>
              </template>
            </v-radio>
          </v-card>
        </v-radio-group>
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
          Export
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script setup>
import {ref} from 'vue';

const model = defineModel({type: Boolean});
const props = defineProps({
	workspace: {type: Object, required: true},
});
const emit = defineEmits(['submit']);

const exportOption = ref('workspace');

// Functions
function closeDialog() {
	model.value = false;
}

function submit() {
	emit('submit', props.workspace, exportOption.value);
	model.value = false;
}

</script>

<style scoped>

</style>
