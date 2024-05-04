<template>
  <v-dialog
    v-model="model"
    max-width="500px"
  >
    <v-card>
      <v-card-title>
        <span class="text-h5">{{ title }}</span>
      </v-card-title>
      <v-card-text>
        <v-textarea
          v-if="textArea"
          v-model="note"
          :label="inputLabel"
          counter
          :maxlength="maxlength"
          :autofocus="true"
        />
        <v-text-field
          v-else
          v-model="note"
          :label="inputLabel"
          counter
          :maxlength="maxlength"
          :autofocus="true"
          @keydown.enter="submit"
        />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn
          variant="text"
          @click="closeDialog"
        >
          {{ cancelLabel }}
        </v-btn>
        <v-btn
          variant="text"
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
	textArea: {type: Boolean, required: false, default: false},
});
const emit = defineEmits(['submit']);
import {VTextarea} from 'vuetify/components';
import {VTextField} from 'vuetify/components';

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
	console.log('submitted');
	emit('submit', note.value);
	model.value = false;
}

</script>

<style scoped>

</style>
