<template>
  <span>
    <v-tooltip
      location="bottom"
      :text="tooltipText"
    >
      <template #activator="item">
        <v-chip
          :label="true"
          class="overflow-x-auto"
          :color="attribution.isPublic?'primary':null"
          v-bind="item.props"
          :rounded="true"
        >
          {{ attribution.tag }}
        </v-chip>
      </template>
    </v-tooltip>
  </span>
</template>

<script setup>
const props = defineProps({attribution: {type: Object, required: true}});

const privateTooltip = 'This is a private attribution. It is only visible to you.';
const publicTooltip = 'This is a public attribution';
const tooltipText = props.attribution.isPublic ? publicTooltip : privateTooltip;

</script>

<style scoped>

</style>
