<template>
  <div
    v-tooltip="{'text':tooltip, 'location':'top'}"
    class="d-flex align-center"
  >
    <v-icon
      :icon="icon"
      class="me-1"
      size="large"
    />
    <p class="text-subtitle-1">
      {{ title }}
    </p>
  </div>
</template>

<script setup>

defineProps({
	title: {type: String, required: true},
	icon: {type: String, required: true},
	tooltip: {type: String, default: ''},
});

</script>

<style scoped>

</style>
