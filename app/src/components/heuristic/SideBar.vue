<template>
  <v-slide-x-reverse-transition>
    <v-sheet
      v-if="inputVal"
      class="sidebar"
      elevation="4"
      :style="`max-width:${maxWidth}; min-width:${minWidth}`"
    >
      <icon-title
        :title="title"
        :icon="icon"
        :one-line="true"
      >
        <slot name="actions" />
        <v-btn
          icon
          variant="text"
          color="grey"
          @click="inputVal=false"
        >
          <v-icon :icon="mdiCloseCircle" />
        </v-btn>
      </icon-title>
      <v-divider />
      <slot name="body" />
    </v-sheet>
  </v-slide-x-reverse-transition>
</template>

<script setup>
import IconTitle from '@/components/common/IconTitle.vue';
import {mdiCloseCircle} from '@mdi/js';
import {computed} from 'vue';

const props = defineProps({
	modelValue: {type: Boolean, required: true},
	title: {type: String, required: true},
	icon: {type: String, required: true},
	maxWidth: {type: String, required: false, default: '600px'},
	minWidth: {type: String, required: false, default: '300px'},
});

const emit = defineEmits(['update:modelValue']);

const inputVal = computed({
	get() {
		return props.modelValue;
	},
	set(val) {
		emit('update:modelValue', val);
	},
});

</script>

<style scoped>
.sidebar {
  position: absolute;
  top: 0;
  right: 0;
  height: 100%;
  /* Heuristic toolbar a z-index of 1004, therefore set z-index to the same so top shadow is not visible */
  z-index: 1004;
  overflow: auto;
}
</style>
