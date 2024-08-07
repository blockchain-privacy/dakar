import {
	DefaultApi,
	Configuration,
} from '@neondark/wikiapi';
import {checkResponseStatus} from '@/utilities';
import {useNavStore} from '@/pinia/nav';
import {useLocalStore} from '@/pinia/local';

function newConfig(v) {
	return new Configuration({
		basePath: '/wikiapi',
		credentials: 'include',
		middleware: [{
			async post(d) {
				await checkResponseStatus(v, useNavStore(), useLocalStore(), d.response);
			},
		}],
	});
}

export default {
	setup(v) {
		const c = newConfig(v);
		return {
			default: new DefaultApi(c),
		};
	},
};
