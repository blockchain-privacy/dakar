<template>
  <v-card class="mx-auto px-2" tile flat :to="to" max-width="90">
    <div class="d-flex justify-center">
      <v-icon size="40">{{ icon }}</v-icon>
    </div>
    <div class="d-flex justify-center pt-1">
      <p class="text-subtitle-1 title-text">{{ title }}</p>
    </div>
  </v-card>
</template>

<script>
export default {
  name: 'LinkCard',
  props: {
    title: { type: String, required: true },
    to: { type: String, required: true },
    icon: { type: String, required: true },
  },
};
</script>

<style scoped>
.title-text {
  text-align: center;
  line-height: 1.3;
}
</style>
