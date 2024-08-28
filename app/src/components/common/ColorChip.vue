<template>
  <v-chip rounded>
    <template #prepend>
      <color-sheet
        :color="color"
        class="me-2"
      />
    </template>
    {{ capitalize(title) }}
  </v-chip>
</template>

<script setup>
import {capitalize} from '@/utilities/index.js';
import ColorSheet from '@/components/common/ColorSheet.vue';

defineProps({
	color: {type: String, required: true},
	title: {type: String, required: true},
});
</script>

<style scoped>

</style>
