<template>
  <v-alert
    :type="message.type"
    class="mt-2"
    density="compact"
    variant="text"
  >
    {{ message.text }}
  </v-alert>
</template>

<script setup>

defineProps({message: {type: Object, required: true}});

</script>

<style scoped>

</style>
