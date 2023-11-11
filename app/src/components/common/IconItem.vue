<template>
  <v-list-item>
    <template #prepend>
      <v-icon
        v-if="!isColor"
        size="xx-large"
      >
        {{ icon }}
      </v-icon>
      <v-icon
        v-if="isColor"
        :class="{ 'green-icon': !isRed, 'red-icon': isRed }"
        size="xx-large"
      >
        {{ icon }}
      </v-icon>
    </template>
    <div>
      <v-list-item-title>
        {{ title }}
        <v-tooltip
          v-if="tooltip"
          :text="tooltip"
          location="right"
        >
          <template #activator="{ props }">
            <v-icon
              size="x-small"
              v-bind="props"
              :icon="icons.mdiHelpCircleOutline"
            />
          </template>
        </v-tooltip>
      </v-list-item-title>
      <v-list-item-subtitle>
        <slot />
      </v-list-item-subtitle>
    </div>
  </v-list-item>
</template>

<script>
import {mdiHelpCircleOutline} from '@mdi/js';

export default {
	name: 'IconItem',
	props: {
		title: {type: String, required: true},
		icon: {type: String, required: true},
		tooltip: {type: String, default: ''},
		isColor: {type: Boolean, default: false},
		isRed: {type: Boolean, default: false},
	},
	data() {
		return {
			icons: {mdiHelpCircleOutline},
		};
	},
};
</script>

<style scoped>
:deep(.green-icon .v-icon__svg) {
  fill: green;
}

:deep(.red-icon .v-icon__svg) {
  fill: red;
}

</style>
