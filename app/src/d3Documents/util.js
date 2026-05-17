// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

export function sleep(ms) {
	// eslint-disable-next-line no-promise-executor-return
	return new Promise(resolve => setTimeout(resolve, ms));
}

// Credits: https://stackoverflow.com/questions/9461621/format-a-number-as-2-5k-if-a-thousand-or-more-otherwise-900
export function abbreviateNumber(number) {
	// What tier? (determines SI symbol)
	const tier = Math.trunc(Math.log10(Math.abs(number)) / 3);

	// If zero, we don't need a suffix
	if (tier === 0 || Number.isNaN(tier) || !Number.isFinite(tier)) {
		return number;
	}

	const SI_SYMBOL = ['', 'k', 'M', 'G', 'T', 'P', 'E'];

	// Get suffix and determine scale
	const suffix = SI_SYMBOL[tier];
	const scale = 10 ** (tier * 3);

	// Scale the number
	const scaled = number / scale;

	// Format number and add suffix
	return Number(scaled.toFixed(1)).toLocaleString() + suffix;
}

// Returns the ratio of a shortened line
export function getRatio(source, target, nodeRadius) {
	const c = Math.hypot(target.x - source.x, target.y - source.y);
	// 10 is the marker width
	// c2 must not be negative
	const c2 = Math.max(c - nodeRadius - 10, 0);

	return c2 / c;
}

// Returns a new reduced y coordinate of the target point.
// To switch the direction, switch source and target arguments
export function reduceY(source, target, nodeRadius) {
	const dy = (target.y - source.y) * getRatio(source, target, nodeRadius);
	return source.y + dy;
}

// Returns a new reduced x coordinate of the target point.
// To switch the direction, switch source and target arguments
export function reduceX(source, target, nodeRadius) {
	const dx = (target.x - source.x) * getRatio(source, target, nodeRadius);
	return source.x + dx;
}
