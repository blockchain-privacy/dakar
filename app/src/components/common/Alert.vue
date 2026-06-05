<template>
  <v-expand-transition>
    <v-alert
      v-if="text"
      class="my-2"
      :text="text"
      :type="type"
      variant="tonal"
      :closable="closable"
    />
  </v-expand-transition>
</template>

<script setup>

defineProps({
	text: {type: String, required: true},
	type: {type: String, required: false, default: 'error'},
	closable: {type: Boolean, required: false},
});

</script>
<style scoped>

</style>
