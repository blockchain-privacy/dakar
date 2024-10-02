<template>
  <div
    class="mb-1 mt-2 d-flex align-center text-h6"
    style="white-space: nowrap"
  >
    <v-icon class="me-2">
      {{ icon }}
    </v-icon>
    <router-link
      v-if="to"
      class="shorten"
      style="color: inherit;"
      :to="to"
    >
      {{ title }}
    </router-link>
    <span
      v-else
      class="shorten"
    > {{ title }}</span>
    <div
      v-if="oneLine || !$vuetify.display.xs"
      class="ms-auto"
    >
      <slot />
    </div>
  </div>
  <div
    v-if="!oneLine && $vuetify.display.xs"
    class="d-flex align-center justify-end"
  >
    <slot />
  </div>
</template>

<script setup>

defineProps({
	title: {type: String, required: true},
	icon: {type: String, required: true},
	to: {type: Object, required: false, default: null},
	oneLine: {type: Boolean, required: false, default: false},
});

</script>

<style scoped>
.shorten {
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}
</style>
