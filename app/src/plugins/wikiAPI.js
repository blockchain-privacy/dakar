// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {
	DefaultApi,
	Configuration,
} from '@blockchain-privacy/wikiapi';
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
