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
        <p class="text-title-small mb-2">
          Blockchain
        </p>
        <v-btn-toggle
          v-if="showModeSwitch"
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
import {BLOCKCHAIN_ATTRIBUTES} from '@/constants/index.js';

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
	showModeSwitch: {type: Boolean, required: false},
});
const emit = defineEmits(['submit']);

const note = ref('');
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
	emit('submit', note.value, modeSelection.value);
	model.value = false;
}

</script>

<style scoped>

</style>
