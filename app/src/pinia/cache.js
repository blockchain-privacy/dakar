// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {defineStore} from 'pinia';

const maxElements = 30;
export const useCacheStore = defineStore('cache', {
	state: () => ({
		cache: new Map(),
	}),
	getters: {
		getCache: state => state.cache,
	},
	actions: {
		setValue(key, value) {
			this.cache.set(key, value);
			// Remove first (oldest) element when map has become to large
			if (this.cache.size > maxElements) {
				this.cache.delete(this.cache.keys().next().value);
			}
		},
		getValue(key) {
			return this.cache.get(key);
		},
		resetCache() {
			this.cache.clear();
		},
		removeValue(key) {
			this.cache.delete(key);
		},
	},
});
