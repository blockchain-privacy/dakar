import {
	Configuration, OAuth2Api, IdentityApi, FrontendApi,
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
	identity: new IdentityApi(config),
	frontend: new FrontendApi(config),
	oauth2: new OAuth2Api(config),
};
