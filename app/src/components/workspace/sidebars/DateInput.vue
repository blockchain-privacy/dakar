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
        :label="label"
        :model-value="model?.toLocaleDateString()"
        hide-details
        class="me-2"
        :rules="rules"
        :error="error"
      />
    </template>
    <v-card>
      <v-date-picker
        v-model="model"
        hide-header
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

defineProps({
	label: {type: String, required: false, default: 'Date'},
	rules: {type: Array, required: false, default: undefined},
	error: {type: Boolean, required: false, default: false},
});

const model = defineModel({type: Date});
const menuModel = ref(false);

</script>
<style scoped>

</style>
