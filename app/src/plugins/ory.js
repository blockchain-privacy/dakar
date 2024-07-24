import {Configuration, FrontendApi} from '@ory/client-fetch';

const c = new Configuration({
	basePath: '/auth',
	credentials: 'include',
	headers: {
		Accept: 'application/json',
	},
	middleware: [{
		async post(d) {
			// Decode JSON of error
			if (!d.response.ok) {
				return await d.response.json();
			}
		},
	}],
});

export default {
	frontend: new FrontendApi(c),
};
