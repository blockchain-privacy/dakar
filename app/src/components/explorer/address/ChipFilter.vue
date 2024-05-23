<template>
  <!-- selected-class="" is intentionally left blank to avoid a shadow over the chip elements -->
  <v-chip-group
    v-model="selectedChips"
    column
    multiple
    filter
    mandatory
    :disabled="disabled"
    selected-class=""
    color="primary"
    @update:model-value="handleModelChange"
  >
    <v-chip
      v-for="item in items"
      :key="item.id"
    >
      <template #prepend>
        <v-sheet
          style="width:25px; height:15px"
          rounded
          :color="item.color?item.color:'black'"
          class="me-2"
        />
      </template>
      {{ item.text }}
    </v-chip>
  </v-chip-group>
</template>

<script setup>
import {onMounted, ref, toRaw} from 'vue';

const props = defineProps({
	disabled: {type: Boolean, required: false, default: false},
	// Example: [{id: 0x123, color: red: text: 'some text'}, ...]
	items: {type: Array, required: true},
});

const emit = defineEmits(['changed']);

const selectedChips = ref([0, 1, 2, 3, 4]);

// Hooks
onMounted(() => {
	selectedChips.value = props.items.map((_, i) => i);
});

// Functions
function handleModelChange() {
	// Send raw array not ref
	emit('changed', [...selectedChips.value]);
}

</script>

<style scoped>

</style>
