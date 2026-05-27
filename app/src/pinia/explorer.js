// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {defineStore} from 'pinia';

export const useExplorerStore = defineStore('explorer', {
	state: () => ({
		highlightWasabi2Denominations: false,
	}),
	getters: {
		getHighlightWasabi2Denominations: state => state.highlightWasabi2Denominations,
	},
	actions: {},
});
