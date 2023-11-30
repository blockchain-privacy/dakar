import {defineStore} from 'pinia';

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
			this.failedRoute = payload;
		},
	},
});
