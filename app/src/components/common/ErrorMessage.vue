<template>
  <v-expand-transition>
    <v-alert
      v-if="text"
      class="my-2"
      :text="text"
      type="error"
      variant="tonal"
    />
  </v-expand-transition>
</template>

<script setup>

defineProps({text: {type: String, required: true}});

</script>
<style scoped>

</style>
