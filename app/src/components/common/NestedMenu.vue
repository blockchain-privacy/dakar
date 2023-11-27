<!-- source: https://github.com/vuetifyjs/vuetify/issues/1877#issuecomment-593273676  -->
<template>
  <v-menu
    v-model="inputVal"
    :style="{'position':absolute?'absolute':null,
             'left': !!positionX?positionX + 'px':null, 'top': !!positionY?positionY + 'px':null}"
    :offset="isOffset?5:null"
    :open-on-hover="isOpenOnHover"
    :transition="transition"
  >
    <template #activator="item">
      <v-btn
        v-if="icon"
        :color="color"
        v-bind="item.props"
      >
        <v-icon>{{ icon }}</v-icon>
      </v-btn>
      <v-list-item
        v-else-if="isSubMenu"
        class="d-flex justify-space-between"
        v-bind="props"
      >
        {{ name }}
        <v-icon>{{ mdiChevronRight }}</v-icon>
      </v-list-item>
    </template>
    <v-list>
      <template
        v-for="(item, index) in menuItems"
        :key="index"
      >
        <v-divider v-if="item.isDivider" />
        <nested-menu
          v-else-if="item.menu"
          :name="item.title"
          :menu-items="item.menu"
          :is-open-on-hover="false"
          is-offset
          is-sub-menu
          @nested-menu-click="emitClickEvent"
        />
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
import {mdiChevronRight} from '@mdi/js';
import {computed} from 'vue';

const props = defineProps({
	modelValue: Boolean,
	name: {type: String, default: ''},
	icon: {type: String, default: ''},
	menuItems: {type: Array, default: () => []},
	absolute: {type: Boolean, default: false},
	color: {type: String, default: 'secondary'},
	positionX: {type: Number, default: 0},
	positionY: {type: Number, default: 0},
	isOffset: {type: Boolean, default: false},
	isOpenOnHover: {type: Boolean, default: false},
	isSubMenu: {type: Boolean, default: false},
	transition: {type: String, default: 'fade-transition'},
});

const emit = defineEmits(['update:modelValue', 'nested-menu-click']);

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
	// This.closeAllMenus() // Theoretically, create a method that does this as a workaround
	emit('nested-menu-click', item);
}

</script>

<style scoped>

</style>
