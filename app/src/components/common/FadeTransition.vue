<template>
  <Transition
    name="component-fade"
    mode="out-in"
  >
    <slot />
  </Transition>
</template>

<script setup>

</script>

<style>
.component-fade-enter-active, .component-fade-leave-active {
  transition: opacity 0.2s ease;
}

.component-fade-enter-from, .component-fade-leave-to {
  opacity: 0;
}

</style>
