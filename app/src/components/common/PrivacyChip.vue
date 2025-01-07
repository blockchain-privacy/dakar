<template>
  <wiki-tooltip
    :description-url="transactionTypeWikiPath"
    hide-link
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
import WikiTooltip from '../wiki/WikiTooltip.vue';
import {mdiIncognito} from '@mdi/js';
import {
	PRIVACY_TYPE_CC,
	PRIVACY_TYPE_CP,
	PRIVACY_TYPE_DESTINATION,
	PRIVACY_TYPE_MIXING,
	PRIVACY_TYPE_ORIGIN,
	PRIVACY_TYPE_WASABI_2_DESTINATION,
	PRIVACY_TYPE_WASABI_2_MIXING,
	PRIVACY_TYPE_WASABI_2_ORIGIN,
	PRIVACY_TYPE_WHIRLPOOL_DESTINATION,
	PRIVACY_TYPE_WHIRLPOOL_MIXING,
	PRIVACY_TYPE_WHIRLPOOL_ORIGIN,
} from '@/constants/index.js';
import {computed} from 'vue';

const props = defineProps({
	transactionType: {type: String, required: true},
	size: {type: String, required: false, default: 'default'},
});

const transactionTypeWikiPath = computed(() => {
	const dashDirectory = 'dash';
	const wasabiDirectory = 'wasabi';
	const whirlpoolDirectory = 'whirlpool';

	switch (props.transactionType) {
		case PRIVACY_TYPE_ORIGIN: return `${dashDirectory}/originTransaction.md`;
		case PRIVACY_TYPE_MIXING: return `${dashDirectory}/mixingTransaction.md`;
		case PRIVACY_TYPE_DESTINATION: return `${dashDirectory}/destinationTransaction.md`;
		case PRIVACY_TYPE_CC: return `${dashDirectory}/collateralCreationTransaction.md`;
		case PRIVACY_TYPE_CP: return `${dashDirectory}/collateralPaymentTransaction.md`;
		case PRIVACY_TYPE_WASABI_2_ORIGIN: return `${wasabiDirectory}/originTransaction.md`;
		case PRIVACY_TYPE_WASABI_2_MIXING: return `${wasabiDirectory}/mixingTransaction.md`;
		case PRIVACY_TYPE_WASABI_2_DESTINATION: return `${wasabiDirectory}/destinationTransaction.md`;
		case PRIVACY_TYPE_WHIRLPOOL_ORIGIN: return `${whirlpoolDirectory}/originTransaction.md`;
		case PRIVACY_TYPE_WHIRLPOOL_MIXING: return `${whirlpoolDirectory}/mixingTransaction.md`;
		case PRIVACY_TYPE_WHIRLPOOL_DESTINATION: return `${whirlpoolDirectory}/destinationTransaction.md`;
		default: return '';
	}
});

</script>
<style scoped>

</style>
