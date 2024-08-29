<template>
  <wiki-tooltip
    :description-url="getPrivacyTypeTooltip(transactionType)"
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
      {{ transactionType }}
    </v-chip>
  </wiki-tooltip>
</template>

<script setup>
import {getPrivacyTypeTooltip} from '@/utilities';
import WikiTooltip from '../wiki/WikiTooltip.vue';
import {mdiIncognito} from '@mdi/js';

defineProps({
	transactionType: {type: String, required: true},
	size: {type: String, required: false, default: 'default'},
});

</script>
<style scoped>

</style>
