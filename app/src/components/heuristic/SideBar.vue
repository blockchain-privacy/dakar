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
          <v-icon>{{ icons.mdiCloseCircle }}</v-icon>
        </v-btn>
      </icon-title>
      <v-divider />
      <slot name="body" />
    </v-sheet>
  </v-slide-x-reverse-transition>
</template>

<script>
import IconTitle from '@/components/common/IconTitle.vue';
import {mdiCloseCircle} from '@mdi/js';

export default {
	name: 'SideBar',
	components: {IconTitle},
	props: {
		modelValue: {type: Boolean, required: true},
		title: {type: String, required: true},
		icon: {type: String, required: true},
		maxWidth: {type: String, required: false, default: '600px'},
		minWidth: {type: String, required: false, default: '300px'},
	},
	emits: ['update:modelValue'],
	data() {
		return {
			icons: {
				mdiCloseCircle,
			},
		};
	},
	computed: {
		inputVal: {
			get() {
				return this.modelValue;
			},
			set(val) {
				this.$emit('update:modelValue', val);
			},
		},
	},
};
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
