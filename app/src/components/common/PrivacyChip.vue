<template>
  <wiki-tooltip
    :description-url="getPrivacyTypeTooltip(privacyType)"
    :show-link="false"
  >
    <v-chip
      v-bind="$attrs"
      rounded
      color="purple"
      :size="size"
      style="cursor: pointer"
      class="text-capitalize"
    >
      <v-icon
        :icon="mdiIncognito"
        class="me-2"
      />
      {{ getPrivacyTypeLabel(privacyType) }}
    </v-chip>
  </wiki-tooltip>
</template>

<script setup>
import {getPrivacyTypeLabel, getPrivacyTypeTooltip} from '@/utilities';
import WikiTooltip from '../wiki/WikiTooltip.vue';
import {mdiIncognito} from '@mdi/js';

defineProps({
	privacyType: {type: Number, required: true},
	size: {type: String, required: false, default: 'default'},
});

</script>
<style scoped>

</style>
