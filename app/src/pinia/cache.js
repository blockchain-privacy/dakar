// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {defineStore} from 'pinia';
import {ref} from 'vue';

const maxElements = 50;
const localStoreKey = 'cacheStore';
let hasCacheChanged = false;

async function persistCache(m) {
	if (!hasCacheChanged || !isBase64Supported()) {
		return;
	}

	// Compress map and store it encoded in base64
	const compressedResponse = await new Response(new Blob([JSON.stringify([...m])], {
		type: 'application/json',
	}).stream().pipeThrough(new CompressionStream('gzip')));
	const blob = await compressedResponse.blob();
	const buffer = await blob.arrayBuffer();

	localStorage.setItem(localStoreKey, new Uint8Array(buffer).toBase64());

	hasCacheChanged = false;
}

async function visibilityChanged(m) {
	if (document.visibilityState === 'visible') {
		return;
	}

	await persistCache(m);
}

// Todo: remove check after mid 2026
function isBase64Supported() {
	return Uint8Array.fromBase64 !== undefined;
}

async function loadCache() {
	if (!isBase64Supported()) {
		return new Map();
	}

	const localItem = localStorage.getItem(localStoreKey);
	if (localItem !== null) {
		// Decompress map
		const blob = new Blob([Uint8Array.fromBase64(localItem)]);
		const ds = new DecompressionStream('gzip');
		const decompressedStream = blob.stream().pipeThrough(ds);
		const parsedItem = await new Response(decompressedStream).json();
		const filledMap = new Map(parsedItem);

		// Convert timestamp strings to dateTime objects
		for (const [key, item] of filledMap) {
			// Todo remove else branch, just kept for compatibility for cache items without ts
			// eslint-disable-next-line unicorn/prefer-ternary
			if (item.ts) {
				item.ts = new Date(item.ts);
			} else {
				// Set old date, so cache items get replaced eventually
				item.ts = new Date('January 1, 1970');
			}

			filledMap.set(key, item);
		}

		return filledMap;
	}

	return new Map();
}

// Removes all items from the given map that are expired.
function removeExpiredItems(m) {
	for (const [key, item] of m) {
		if (item.ts && (Date.now() - item.ts) > item.ttl) {
			m.delete(key);
			hasCacheChanged = true;
		}
	}
}

const initialMap = await loadCache();
// Cache item structure: {ts: dateTime, value: any}
export const useCacheStore = defineStore('cache', () => {
	const cache = ref(initialMap);
	// 1000 * 60 = 60000 milliseconds = 1 minute
	setInterval(() => removeExpiredItems(cache.value), 60_000);

	setInterval(() => persistCache(cache.value), 90_000);
	document.addEventListener('visibilitychange', () => visibilityChanged(cache.value));

	// Set with ttl in minutes
	function setTTL(key, value, ttl) {
		if (ttl > 0) {
			cache.value.set(key, {ts: new Date(), value, ttl: ttl * 60_000});
		} else {
			cache.value.set(key, {ts: new Date(), value});
		}

		hasCacheChanged = true;
		// Remove first (oldest) element when map has become to large
		if (cache.value.size > maxElements) {
			cache.value.delete(cache.value.keys().next().value);
		}
	}

	function set(key, value) {
		setTTL(key, value, 0);
	}

	function get(key) {
		return cache.value.get(key)?.value;
	}

	function getWithMetadata(key) {
		return cache.value.get(key);
	}

	return {
		cache, set, setTTL, get, getWithMetadata,
	};
});
