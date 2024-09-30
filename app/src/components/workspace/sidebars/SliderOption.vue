<template>
  <div>
    <div class="text-center text-subtitle-1 mt-2">
      {{ label }}
    </div>
    <div class="d-flex align-center">
      <v-slider
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
