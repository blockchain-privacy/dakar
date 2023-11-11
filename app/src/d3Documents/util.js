
export function sleep(ms) {
	// eslint-disable-next-line no-promise-executor-return
	return new Promise(resolve => setTimeout(resolve, ms));
}

// Credits: https://stackoverflow.com/questions/9461621/format-a-number-as-2-5k-if-a-thousand-or-more-otherwise-900
export function abbreviateNumber(number) {
	// What tier? (determines SI symbol)
	// eslint-disable-next-line no-bitwise
	const tier = Math.log10(Math.abs(number)) / 3 | 0;

	// If zero, we don't need a suffix
	if (tier === 0) {
		return number;
	}

	const SI_SYMBOL = ['', 'k', 'M', 'G', 'T', 'P', 'E'];

	// Get suffix and determine scale
	const suffix = SI_SYMBOL[tier];
	const scale = 10 ** (tier * 3);

	// Scale the number
	const scaled = number / scale;

	// Format number and add suffix
	return scaled.toFixed(1) + suffix;
}
