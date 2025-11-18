// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {defineStore} from 'pinia';
import {toRaw} from 'vue';

export const useNavStore = defineStore('nav', {
	state: () => ({
		// FailedRoute is filled with the route which the user wanted
		// to access but did for some reason (e.g. invalid credentials) fail
		failedRoute: null,
		// PushFromUserInput is true if a data route
		// navigated to by using router.push() instead of browser navigation
		pushFromUserInput: false,
	}),
	actions: {
		setFailedRoute(payload) {
			const rawRoute = {};
			// It is not enough to use toRaw, need to also use Object.assign
			// to get a route object which does not change when the current route changes
			Object.assign(rawRoute, toRaw(payload));
			this.failedRoute = rawRoute;
		},
	},
});
