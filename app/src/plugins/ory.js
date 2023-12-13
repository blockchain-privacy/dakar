import {
	Configuration, FrontendApi,
} from '@ory/client';

const config = new Configuration({
	basePath: '/auth',
	baseOptions: {
		withCredentials: true,
		headers: {
			Accept: 'application/json',
		},
	},
});

export default {
	frontend: new FrontendApi(config),
};
