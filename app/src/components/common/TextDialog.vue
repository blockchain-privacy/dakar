<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="500px"
  >
    <v-card>
      <v-card-title>{{ title }}</v-card-title>
      <v-card-text>
        <v-textarea
          v-if="textArea"
          v-model="note"
          :label="inputLabel"
          counter
          :maxlength="maxlength"
          autofocus
        />
        <v-text-field
          v-else
          v-model="note"
          :label="inputLabel"
          counter
          :maxlength="maxlength"
          autofocus
          @keydown.enter="submit"
        />
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
          :color="submitColor"
          @click="submit"
        >
          {{ submitLabel }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script setup>
import {onMounted, onUpdated, ref} from 'vue';

const model = defineModel({type: Boolean});
const props = defineProps({
	title: {type: String, required: true},
	maxlength: {type: Number, required: true},
	submitLabel: {type: String, required: false, default: 'Submit'},
	cancelLabel: {type: String, required: false, default: 'Cancel'},
	inputLabel: {type: String, required: false, default: ''},
	inputValue: {type: String, required: false, default: ''},
	cancelColor: {type: String, required: false, default: undefined},
	submitColor: {type: String, required: false, default: undefined},
	textArea: {type: Boolean, required: false},
});
const emit = defineEmits(['submit']);

const note = ref('');
let oldNote = '';

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
	emit('submit', note.value);
	model.value = false;
}

</script>

<style scoped>

</style>
