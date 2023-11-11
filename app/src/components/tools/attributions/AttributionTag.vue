<template>
  <span>
    <v-tooltip
      location="bottom"
      :text="tooltipText"
    >
      <template #activator="{ props }">
        <v-chip
          :label="true"
          class="overflow-x-auto"
          :color="attribution.isPublic?'primary':null"
          v-bind="props"
          :rounded="true"
        >
          {{ attribution.tag }}
        </v-chip>
      </template>
    </v-tooltip>
  </span>
</template>

<script>
export default {
	name: 'AttributionTag',
	props: {attribution: {type: Object, required: true}},
	data() {
		return {
			privateTooltip: 'This is a private attribution. It is only visible to you.',
			publicTooltip: 'This is a public attribution',
		};
	},
	computed: {
		tooltipText() {
			return this.attribution.isPublic ? this.publicTooltip : this.privateTooltip;
		},
	},
};
</script>

<style scoped>

</style>
