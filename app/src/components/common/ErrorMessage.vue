<template>
  <v-alert
    v-if="text"
    :text="text"
    type="error"
    variant="tonal"
  />
</template>

<script setup>

defineProps({text: {type: String, required: true}});

</script>
<style scoped>

</style>
