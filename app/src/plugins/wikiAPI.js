import {
	DefaultApi,
	Configuration,
} from '@neondark/wikiapi';
import {checkResponseStatus} from '@/utilities';

function newConfig(v) {
	return new Configuration({
		basePath: '/wikiapi/',
		credentials: 'include',
		middleware: [{
			async post(d) {
				await checkResponseStatus(v, d.response);
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
