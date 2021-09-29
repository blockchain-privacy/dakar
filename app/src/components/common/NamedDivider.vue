<template>
  <div :class="`my-${verticalMargin}`">
    <div class="d-flex align-center">
      <v-sheet color="grey lighten-2" height="1" width="100%"/>
      <div class="mx-2" v-if="title" style="white-space: nowrap;">{{ title }}</div>
      <v-sheet color="grey lighten-2" height="1" width="100%" v-if="title"/>
    </div>
  </div>
</template>

<script>
export default {
  name: 'NamedDivider',
  props: {
    title: { type: String, required: false },
    verticalMargin: { type: Number, required: false, default: 8 },
  },
};
</script>

<style scoped>

</style>
