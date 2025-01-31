<template>
  <div class="amountChart d-flex">
    <div
      v-for="(t,i) in amountsPerType"
      :key="t.type"
      v-tooltip="{'text': tooltipText(t.type,t.amount, t.percent), 'location':'top', 'open-delay': 400}"
      className="amountElement"
      :style="`width:${displayPercent[i]}%; background-color:${colorMap.get(t.type)}`"
    />
  </div>
</template>

<script setup>

import {computed} from 'vue';
import {
	convertAmount, getCoinUnit, getColorMap, setUndefinedTransactionColor,
} from '@/utilities/index.js';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';

const props = defineProps({outputs: {type: Array, required: true}});
const noTypeKey = 'no type';
const {getSettings} = storeToRefs(useLocalStore());
const colorMap = getColorMap(getSettings.value.blockchainMode);
setUndefinedTransactionColor(colorMap, noTypeKey);

// Computed
const coinUnit = computed(() => getCoinUnit(getSettings.value.blockchainMode));
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

		let t = output.txtype;

		t ||= noTypeKey;

		let val = typeMap.get(t);
		if (val) {
			val += output.amount;
		} else {
			val = output.amount;
		}

		typeMap.set(t, val);
	}

	return Array.from(typeMap, ([type, amount]) => ({type, amount, percent: amount / amountSum * 100})).sort((a, b) => b.amount - a.amount);
});

// Makes sure that each type is represented by at least 1%, so it is easier to see in the chart.
// Changes the larger percentages accordingly.
// Returns an arry with the percent distribution in the same order as amountsPerType.
const displayPercent = computed(() => {
	let newBase = 100.0;
	const minPercent = 1.0;

	for (const t of amountsPerType.value) {
		if (t.percent < minPercent) {
			newBase -= minPercent;
		}
	}

	return amountsPerType.value.map(t => {
		let p;
		if (t.percent >= minPercent) {
			p = t.percent / 100.0 * newBase;
		} else {
			p = minPercent;
		}

		return p;
	});
});

// Functions
function tooltipText(type, amount, percent) {
	return `${type}: ${convertAmount(amount).toLocaleString()} ${coinUnit.value}, ${percent.toFixed(2)}%`;
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
