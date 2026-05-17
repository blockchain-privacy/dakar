<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="500px"
  >
    <v-card>
      <v-card-title>{{ title }}</v-card-title>
      <v-card-text>
        <slot />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn
          variant="text"
          :color="cancelColor"
          @click="closeDialog"
        >
          {{ cancelLabel }}
        </v-btn>
        <v-btn
          variant="text"
          :color="confirmColor"
          @click="confirm"
        >
          {{ confirmLabel }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script setup>
const model = defineModel({type: Boolean});
defineProps({
	title: {type: String, required: true},
	confirmLabel: {type: String, required: false, default: 'Confirm'},
	cancelLabel: {type: String, required: false, default: 'Cancel'},
	cancelColor: {type: String, required: false, default: undefined},
	confirmColor: {type: String, required: false, default: undefined},
});
const emit = defineEmits(['confirm']);

// Functions

function closeDialog() {
	model.value = false;
}

function confirm() {
	emit('confirm');
	model.value = false;
}

</script>

<style scoped>

</style>
