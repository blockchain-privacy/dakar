<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

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
      selected-class=""
      class="ms-2"
      @update:model-value="handleModelChange"
    >
      <div class="d-flex align-center justify-center flex-wrap">
        <color-chip
          v-for="item in items"
          :key="item.text"
          :title="item.text"
          :color="item.color"
        />
      </div>
    </v-chip-group>
  </div>
</template>

<script setup>
import ColorChip from '@/components/common/ColorChip.vue';

const model = defineModel({type: Array});

defineProps({
	// Example: [{color: red: text: 'some text'}, ...]
	items: {type: Array, required: true},
	label: {type: String, required: false, default: ''},
	mandatory: {type: Boolean, required: false},
});

const emit = defineEmits(['changed']);

// Functions
function handleModelChange() {
	emit('changed');
}

</script>

<style scoped>

</style>
