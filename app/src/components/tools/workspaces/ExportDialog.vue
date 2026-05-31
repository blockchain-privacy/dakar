<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="500px"
  >
    <v-card>
      <v-card-title>Export Workspace</v-card-title>
      <v-card-text>
        Export workspace '{{ workspace.name }}'?
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
import {onMounted, onUpdated} from 'vue';

const model = defineModel({type: Boolean});
const props = defineProps({
	workspace: {type: Object, required: true},
});
const emit = defineEmits(['submit']);

onMounted(() => {});

onUpdated(() => {});

// Functions

function closeDialog() {
	model.value = false;
}

function submit() {
	emit('submit', props.workspace);
	model.value = false;
}

</script>

<style scoped>

</style>
