import { Configuration, V0alpha2Api } from '@ory/client';
import { ORY_KRATOS_BASE_URL } from '../constants';

export default {
  install(Vue) {
    Vue.prototype.ory = new V0alpha2Api(new Configuration({
      basePath: ORY_KRATOS_BASE_URL,
      baseOptions: {
        withCredentials: true,
        headers: {
          Accept: 'application/json',
        },
      },
    }));
  },
};
