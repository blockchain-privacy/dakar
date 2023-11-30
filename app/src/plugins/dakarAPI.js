import {
	AddressExclusionsApi,
	AttributionApi, AuthenticationApi,
	ClusterApi,
	Configuration,
	DataApi,
	HeuristicApi,
	MetaApi, ToolsApi,
} from '@blockchain/dakar';
import {checkResponseStatus} from '@/utilities';
import {useNavStore} from '@/pinia/nav';
import {useLocalStore} from '@/pinia/local';

function newConfig(v) {
	return new Configuration({
		basePath: '/api/v1',
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
			attribution: new AttributionApi(c),
			authentication: new AuthenticationApi(c),
			tools: new ToolsApi(c),
			data: new DataApi(c),
			meta: new MetaApi(c),
			heuristic: new HeuristicApi(c),
			cluster: new ClusterApi(c),
			addressExclusion: new AddressExclusionsApi(c),
		};
	},
};
