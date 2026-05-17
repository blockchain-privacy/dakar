<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="d-flex justify-center my-2 text-body-large">
    <label :for="`${componentID}_range`">{{ label }}</label>
  </div>
  <div class="d-flex align-center mb-5">
    <v-text-field
      :id="`${componentID}_range`"
      v-model="model.min"
      min-width="100px"
      :rules="rules"
      label="From"
      :placeholder="placeholder"
      hide-details
      class="me-2"
    />
    <v-text-field
      v-model="model.max"
      min-width="100px"
      :rules="rules"
      label="To"
      :placeholder="placeholder"
      hide-details
    />
  </div>
</template>

<script setup>
import {useId} from 'vue';

const componentID = useId();
const model = defineModel({type: Object});

defineProps({
	label: {type: String, required: true},
	rules: {type: Array, required: false, default: undefined},
	placeholder: {type: String, required: true},
});
</script>

<style scoped>

</style>
