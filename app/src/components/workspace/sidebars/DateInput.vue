<template>
  <v-menu
    v-model="menuModel"
    eager
    :close-on-content-click="false"
  >
    <template #activator="a">
      <v-text-field
        v-bind="a.props"
        readonly
        label="From"
        :model-value="model?.toLocaleDateString()"
        hide-details
        class="me-2"
        :rules="rules"
      />
    </template>
    <v-card>
      <v-date-picker
        v-model="model"
        landscape
      />
      <v-card-actions>
        <v-btn
          class="ms-auto"
          variant="text"
          @click="menuModel = false"
        >
          OK
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>
<script setup>
import {ref} from 'vue';

defineProps({rules: {type: Array, required: false, default: undefined}});

const model = defineModel({type: Date});
const menuModel = ref(false);

</script>
<style scoped>

</style>
