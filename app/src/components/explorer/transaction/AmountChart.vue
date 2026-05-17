<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="amountChart d-flex">
    <div
      v-for="(t,i) in amountsPerType"
      :key="t.type"
      v-tooltip="{'text': tooltipText(t.type,t.amount, t.percent), 'location':'top', 'open-delay': 0}"
      class="amountElement"
      :style="`width:${displayPercent[i]}%; background-color:${colorMap.get(t.type)}`"
    />
  </div>
</template>

<script setup>
import {computed} from 'vue';
import {useRoute} from 'vue-router';
import {
	convertAmount,
	getCoinUnit,
	getTransactionColorMap,
	setUndefinedTransactionColor,
} from '@/utilities/index.js';

const props = defineProps({outputs: {type: Array, required: true}});
const noTypeKey = 'no type';
const notSpent = 'not spent';
const route = useRoute();
const colorMap = getTransactionColorMap(route.params.blockchainMode);
setUndefinedTransactionColor(colorMap, noTypeKey);
colorMap.set(notSpent, 'lightgrey');

// Computed
const coinUnit = computed(() => getCoinUnit(route.params.blockchainMode));
const amountsPerType = computed(() => {
	if (!props.outputs) {
		return [];
	}

	const typeMap = new Map();
	let amountSum = 0;
	for (const output of props.outputs) {
		if (!output.amount) {
			continue;
		}

		amountSum += output.amount;

		let t = noTypeKey;
		if (output.txtype) {
			t = output.txtype;
		} else if (!output.ts) {
			t = notSpent;
		}

		let val = typeMap.get(t);
		if (val) {
			val += output.amount;
		} else {
			val = output.amount;
		}

		typeMap.set(t, val);
	}

	return Array.from(typeMap, ([type, amount]) => ({type, amount, percent: amount / amountSum * 100}))
		.toSorted((a, b) => b.amount - a.amount);
});

// Makes sure that each type is represented by at least 1%, so it is easier to see in the chart.
// Changes the larger percentages accordingly.
// Returns an arry with the percent distribution in the same order as amountsPerType.
const displayPercent = computed(() => {
	let newBase = 100;
	const minPercent = 1;

	for (const t of amountsPerType.value) {
		if (t.percent < minPercent) {
			// Reduce base by the difference of t.percent to minPercent
			newBase -= minPercent - t.percent;
		}
	}

	return amountsPerType.value.map(t => {
		if (t.percent >= minPercent) {
			return t.percent / 100 * newBase;
		}

		return minPercent;
	});
});

// Functions
function tooltipText(type, amount, percent) {
	return `${type}: ${convertAmount(amount).toLocaleString()} ${coinUnit.value}, ${percent.toFixed(2).toLocaleString()}%`;
}

</script>

<style scoped>

.amountChart {
  border-radius: 5px;
  overflow: hidden;
}

.amountElement {
  height: 10px;
}

</style>
