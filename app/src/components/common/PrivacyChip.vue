<template>
  <v-chip
    :rounded="true"
    color="purple"
    :size="size"
  >
    <v-icon
      :icon="icons.mdiIncognito"
      class="me-2"
    />
    <wiki-tooltip
      :description-url="getPrivacyTypeTooltip(privacyType)"
      :show-link="false"
    >
      {{ getPrivacyTypeLabel(privacyType) }}
    </wiki-tooltip>
  </v-chip>
</template>

<script>
import {getPrivacyTypeLabel, getPrivacyTypeTooltip} from '@/utilities';
import WikiTooltip from '../wiki/WikiTooltip.vue';
import {mdiIncognito} from '@mdi/js';

export default {
	name: 'PrivacyChip',
	components: {WikiTooltip},
	props: {
		privacyType: {type: Number, required: true},
		size: {type: String, required: false, default: 'default'},
	},
	data() {
		return {
			icons: {
				mdiIncognito,
			},
		};
	},
	methods: {getPrivacyTypeTooltip, getPrivacyTypeLabel},
};
</script>
<style scoped>

</style>
