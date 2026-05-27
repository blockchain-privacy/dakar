// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

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
