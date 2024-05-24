<template>
  <div class="d-flex align-center flex-wrap justify-center">
    <span
      v-if="label"
      class="ms-2 text-subtitle-2"
    >
      {{ label }}
    </span>
    <!-- selected-class="" is intentionally left blank to avoid a shadow over the chip elements -->
    <v-chip-group
      v-model="model"
      column
      multiple
      filter
      :mandatory="mandatory"
      :disabled="disabled"
      selected-class=""
      class="ms-2"
      @update:model-value="handleModelChange"
    >
      <v-chip
        v-for="item in items"
        :key="item.text"
        rounded
      >
        <template #prepend>
          <v-sheet
            style="width:25px; height:15px"
            rounded
            :color="item.color?item.color:'black'"
            class="me-2"
          />
        </template>
        {{ capitalize( item.text) }}
      </v-chip>
    </v-chip-group>
  </div>
</template>

<script setup>
import {capitalize} from '@/utilities';

const model = defineModel({type: Array});

defineProps({
	disabled: {type: Boolean, required: false, default: false},
	// Example: [{color: red: text: 'some text'}, ...]
	items: {type: Array, required: true},
	label: {type: String, required: false, default: ''},
	mandatory: {type: Boolean, required: false, default: false},
});

const emit = defineEmits(['changed']);

// Functions
function handleModelChange() {
	emit('changed');
}

</script>

<style scoped>

</style>
