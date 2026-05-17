<template>
  <v-expand-transition>
    <v-alert
      v-if="text"
      class="my-2"
      :text="text"
      :type="type"
      variant="tonal"
    />
  </v-expand-transition>
</template>

<script setup>

defineProps({
	text: {type: String, required: true},
	type: {type: String, required: false, default: 'error'},
});

</script>
<style scoped>

</style>
