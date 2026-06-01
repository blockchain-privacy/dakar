<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <wiki-tooltip
    :description-url="transactionTypeWikiPath"
    hide-link
  >
    <v-chip
      v-bind="$attrs"
      rounded
      :color="color"
      :size="size"
      style="cursor: pointer"
      class="text-capitalize"
      variant="flat"
    >
      <v-icon
        :icon="mdiIncognito"
        start
      />
      {{ transactionType }}
    </v-chip>
  </wiki-tooltip>
</template>

<script setup>
import {mdiIncognito} from '@mdi/js';
import {computed} from 'vue';
import {useRoute} from 'vue-router';
import WikiTooltip from '../wiki/WikiTooltip.vue';
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
import {getTransactionColorMap} from '@/utilities/index.js';

const props = defineProps({
	transactionType: {type: String, required: true},
	size: {type: String, required: false, default: 'default'},
});
const route = useRoute();

const colorMap = getTransactionColorMap(route.params.blockchainMode);

const color = computed(() => colorMap.get(props.transactionType));

const transactionTypeWikiPath = computed(() => {
	const dashDirectory = 'dash';
	const wasabiDirectory = 'wasabi';
	const whirlpoolDirectory = 'whirlpool';

	switch (props.transactionType) {
		case PRIVACY_TYPE_ORIGIN: {return `${dashDirectory}/originTransaction.md`;}

		case PRIVACY_TYPE_MIXING: {return `${dashDirectory}/mixingTransaction.md`;}

		case PRIVACY_TYPE_DESTINATION: {return `${dashDirectory}/destinationTransaction.md`;}

		case PRIVACY_TYPE_CC: {return `${dashDirectory}/collateralCreationTransaction.md`;}

		case PRIVACY_TYPE_CP: {return `${dashDirectory}/collateralPaymentTransaction.md`;}

		case PRIVACY_TYPE_WASABI_2_ORIGIN: {return `${wasabiDirectory}/originTransaction.md`;}

		case PRIVACY_TYPE_WASABI_2_MIXING: {return `${wasabiDirectory}/mixingTransaction.md`;}

		case PRIVACY_TYPE_WASABI_2_DESTINATION: {return `${wasabiDirectory}/destinationTransaction.md`;}

		case PRIVACY_TYPE_WHIRLPOOL_ORIGIN: {return `${whirlpoolDirectory}/originTransaction.md`;}

		case PRIVACY_TYPE_WHIRLPOOL_MIXING: {return `${whirlpoolDirectory}/mixingTransaction.md`;}

		case PRIVACY_TYPE_WHIRLPOOL_DESTINATION: {return `${whirlpoolDirectory}/destinationTransaction.md`;}

		default: {return '';}
	}
});

</script>
<style scoped>

</style>
