
<template>
  <v-dialog
    v-model="model"
    max-width="500px"
  >
    <v-card>
      <v-card-title>{{ title }}</v-card-title>
      <v-card-text>
        <p class="text-body-large">
          {{ confirmationText }}
        </p>
        <p
          v-for="p in properties"
          :key="p[0]"
          class="text-body-large"
        >
          {{ p[0] }}: {{ p[1] }}
        </p>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn @click="model = false; emit('canceled')">
          Cancel
        </v-btn>
        <v-btn
          color="red"
          @click="model = false; emit('accepted',id)"
        >
          Delete
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script setup>

const model = defineModel({type: Boolean});

const emit = defineEmits(['accepted', 'canceled']);

defineProps({
	// The id of the resource which should be deleted
	id: {type: String, required: true},
	title: {type: String, required: true},
	confirmationText: {type: String, required: true},
	// A 2-dim array, which maps label to item: [[label,item], [label,item], ...]
	properties: {type: Array, required: false, default: () => []},
});

</script>

<style scoped>

</style>
