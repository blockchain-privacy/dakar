<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div>
    <div class="text-center text-body-large mt-2">
      <label :for="`${componentID}_slider`">{{ label }}</label>
    </div>
    <div class="d-flex align-center">
      <v-slider
        :id="`${componentID}_slider`"
        v-model="model"
        :max="max"
        :min="min"
        :step="step"
        hide-details
      />
      <div class="ms-5">
        {{ model }}
      </div>
    </div>
  </div>
</template>
<script setup>
import {useId} from 'vue';

const componentID = useId();

defineProps({
	max: {type: Number, required: true},
	min: {type: Number, required: false, default: 1},
	step: {type: Number, required: false, default: 1},
	label: {type: String, required: true},
});

const model = defineModel({type: Number});
</script>

<style scoped>

</style>
