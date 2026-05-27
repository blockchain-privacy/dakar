// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {
	ClusterApi,
	Configuration,
	DataApi,
	MetaApi,
	ToolsApi,
	WorkspaceApi,
} from '@blockchain/dakar';
import {checkResponseStatus} from '@/utilities';
import {useNavStore} from '@/pinia/nav';
import {useLocalStore} from '@/pinia/local';

// BashPathPrefix should have a leading slash: /someprefix
function newConfig(v, basePathPrefix) {
	return new Configuration({
		basePath: basePathPrefix + '/api/v1',
		credentials: 'include',
		middleware: [{
			async post(d) {
				await checkResponseStatus(v, useNavStore(), useLocalStore(), d.response);
			},
		}],
	});
}

export default {
	setup(v, basePathPrefix) {
		if (!basePathPrefix) {
			throw new Error('prefix for dakar client not set');
		}

		const c = newConfig(v, basePathPrefix);
		return {
			tools: new ToolsApi(c),
			data: new DataApi(c),
			meta: new MetaApi(c),
			cluster: new ClusterApi(c),
			workspace: new WorkspaceApi(c),
		};
	},
};
