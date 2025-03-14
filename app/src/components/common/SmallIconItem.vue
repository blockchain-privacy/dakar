<template>
  <div
    v-tooltip="{'text':tooltip, 'location':'top'}"
    class="d-flex align-center"
  >
    <v-icon
      :icon="icon"
      class="me-1"
      size="large"
      style="opacity: var(--v-medium-emphasis-opacity)"
    />
    <p
      v-if="title"
      class="text-subtitle-1"
    >
      {{ title }}
    </p>
    <slot />
  </div>
</template>

<script setup>
defineProps({
	icon: {type: String, required: true},
	title: {type: String, required: false, default: ''},
	tooltip: {type: String, default: ''},
});
</script>

<style scoped>

</style>
