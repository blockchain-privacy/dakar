import {
  Configuration, OAuth2Api, IdentityApi, FrontendApi,
} from '@ory/client';
import { ORY_KRATOS_PATH_PREFIX } from '../constants';

const config = new Configuration({
  basePath: ORY_KRATOS_PATH_PREFIX,
  baseOptions: {
    withCredentials: true,
    headers: {
      Accept: 'application/json',
    },
  },
});

export default {
  install(Vue) {
    Vue.prototype.ory = {
      identity: new IdentityApi(config),
      frontend: new FrontendApi(config),
      oauth2: new OAuth2Api(config),
    };
  },
};
