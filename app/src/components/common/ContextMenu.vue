<template>
  <v-menu
    v-model="inputVal"
    :open-on-hover="false"
    transition="fade-transition"
    :target="[positionX,positionY]"
  >
    <v-list>
      <template
        v-for="(item, index) in menuItems"
        :key="index"
      >
        <v-divider v-if="item.isDivider" />
        <v-list-item
          v-else
          :key="index"
          :disabled="item.disabled && !item.disabled()"
          @click="emitClickEvent(item)"
        >
          <template
            v-if="item.icon"
            #prepend
          >
            <v-icon>{{ item.icon }}</v-icon>
          </template>
          <v-list-item-title>{{ item.title }}</v-list-item-title>
        </v-list-item>
      </template>
    </v-list>
  </v-menu>
</template>

<script setup>
import {computed} from 'vue';

const props = defineProps({
	modelValue: Boolean,
	menuItems: {type: Array, default: () => []},
	positionX: {type: Number, default: 0},
	positionY: {type: Number, default: 0},
});

const emit = defineEmits(['update:modelValue']);

// Computed
const inputVal = computed({
	get() {
		return props.modelValue;
	},
	set(val) {
		emit('update:modelValue', val);
	},
});

// Functions
function emitClickEvent(item) {
	if (item.action) {
		item.action();
	}

	inputVal.value = false;
}

</script>

<style scoped>

</style>
