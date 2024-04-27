<template>
  <side-bar
    v-model="model"
    title="Connections"
    :icon="mdiShapeCirclePlus"
    max-width="648px"
  >
    <template #body>
      <p>Connections appear here ...</p>
      <p>{{ connection }}</p>
    </template>
  </side-bar>
</template>

<script setup>
import {mdiShapeCirclePlus} from '@mdi/js';
import SideBar from '@/components/common/SideBar.vue';

defineProps({
	connection: {type: Object, required: true},
});

const model = defineModel({type: Boolean});

</script>

<style scoped>

</style>
