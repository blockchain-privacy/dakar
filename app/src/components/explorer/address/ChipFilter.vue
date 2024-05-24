<template>
  <div class="d-flex align-center">
    <span
      v-if="label"
      class="ms-2"
    >
      {{ label }}
    </span>
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
      class="ms-2"
      @update:model-value="handleModelChange"
    >
      <v-chip
        v-for="item in items"
        :key="item.text"
        rounded
      >
        <template #prepend>
          <v-sheet
            style="width:25px; height:15px"
            rounded
            :color="item.color?item.color:'black'"
            class="me-2"
          />
        </template>
        {{ capitalize( item.text) }}
      </v-chip>
    </v-chip-group>
  </div>
</template>

<script setup>
import {onMounted, ref} from 'vue';
import {capitalize} from '@/utilities';

const props = defineProps({
	disabled: {type: Boolean, required: false, default: false},
	// Example: [{color: red: text: 'some text'}, ...]
	items: {type: Array, required: true},
	label: {type: String, required: false, default: ''},
});

const emit = defineEmits(['changed']);

const selectedChips = ref([]);

// Hooks
onMounted(() => {
	selectedChips.value = props.items.map((_, i) => i);
});

// Functions
function handleModelChange() {
	emit('changed', selectedChips.value.map(d => props.items[d].text));
}

</script>

<style scoped>

</style>
