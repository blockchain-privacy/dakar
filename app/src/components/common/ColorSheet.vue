<template>
  <v-sheet
    style="width:15px; height:15px"
    rounded
    :color="color?color:'black'"
  />
</template>

<script setup>

defineProps({color: {type: String, required: true}});

</script>
<style scoped>

</style>
