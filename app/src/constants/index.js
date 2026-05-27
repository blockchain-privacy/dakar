// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

// Search responses
import {bitcoinLogo, dashLogo} from '@/customIcons/index.js';

// Frontend route names
export const ROUTE_NAME_ENTRY_PAGE = 'Entry Page';
export const ROUTE_NAME_STATUS_PAGE = 'Status Page';
export const ROUTE_NAME_404_PAGE = 'Page not found';
export const ROUTE_NAME_NO_RESULTS = 'No results found';
export const ROUTE_NAME_ERROR = 'Error';
export const ROUTE_NAME_LOGIN_PAGE = 'Login Page';
export const ROUTE_NAME_OAUTH_LOGIN_PAGE = 'OAuth Login Page';
export const ROUTE_NAME_OAUTH_CONSENT_PAGE = 'OAuth Consent Page';
export const ROUTE_NAME_OAUTH_VERIFICATION_PAGE = 'OAuth Verification Page';
export const ROUTE_NAME_OAUTH_SUCCESS_PAGE = 'OAuth Success Page';
export const ROUTE_NAME_OAUTH_ERROR_PAGE = 'OAuth Error Page';
export const ROUTE_NAME_ACCOUNT_RECOVERY = 'Account Recovery Page';
export const ROUTE_NAME_USER_ADMIN_PAGE = 'User Administration Page';
export const ROUTE_NAME_USER_PROFILE_PAGE = 'User Profile Page';
export const ROUTE_NAME_UPDATE_PAGE = 'Update Page';
export const ROUTE_NAME_WORKSPACE_PAGE = 'Workspace Editor';
export const ROUTE_NAME_BLOCK_PAGE = 'Block Page';
export const ROUTE_NAME_TRANSACTION_PAGE = 'Transaction Page';
export const ROUTE_NAME_ADDRESS_PAGE = 'Address Page';
export const ROUTE_NAME_WORKSPACES_PAGE = 'Workspace Lookup Page';
export const ROUTE_NAME_CLUSTER_OVERVIEW = 'Custom Clusters Page';
export const ROUTE_NAME_ATTRIBUTIONS = 'Attributions Page';
export const ROUTE_NAME_WIKI_ROOT = 'Wiki Root Page';
export const ROUTE_NAME_WIKI = 'Wiki Page';
export const ROUTE_NAME_ABOUT = 'About Page';
export const ROUTE_NAME_TERMS_OF_USE = 'Terms of Use Page';
export const ROUTE_NAME_PRIVACY = 'Privacy Policy Page';

// Application
export const PAGE_TITLE = 'Dakar';
export const APPLICATION_NAME = 'Dakar';
export const LOCALSTORAGE_FIELD_SETTINGS = 'settings';
export const LOCALSTORAGE_FIELD_SESSION = 'session';
export const LOCALSTORAGE_FIELD_SEARCH_HISTORY = 'searchHistory';
export const LOCALSTORAGE_FIELD_UPDATE_ID = 'updateID';

// Blockchain
export const BLOCKCHAIN_DASH = 'dash';
export const BLOCKCHAIN_BTC = 'btc';
export const BLOCKCHAIN_ATTRIBUTES = {
	[BLOCKCHAIN_BTC]: {
		color: '#FF9315', title: 'Bitcoin', icon: bitcoinLogo, mode: BLOCKCHAIN_BTC,
	},
	[BLOCKCHAIN_DASH]: {
		color: '#008CE4', title: 'Dash', icon: dashLogo, mode: BLOCKCHAIN_DASH,
	},
};

// Cluster types
export const CLUSTER_TYPE_FMI = 'fmi';
export const CLUSTER_TYPE_CUSTOM = 'custom';

// Transaction types
export const PRIVACY_TYPE_DESTINATION = 'destination';
export const PRIVACY_TYPE_CC = 'collateral creation';
export const PRIVACY_TYPE_CP = 'collateral payment';
export const PRIVACY_TYPE_ORIGIN = 'origin';
export const PRIVACY_TYPE_MIXING = 'mixing';

export const PRIVACY_TYPE_WASABI_2_ORIGIN = 'wasabi 2.0 origin';
export const PRIVACY_TYPE_WASABI_2_MIXING = 'wasabi 2.0 mixing';
export const PRIVACY_TYPE_WASABI_2_DESTINATION = 'wasabi 2.0 destination';

export const PRIVACY_TYPE_WHIRLPOOL_ORIGIN = 'whirlpool origin';
export const PRIVACY_TYPE_WHIRLPOOL_MIXING = 'whirlpool mixing';
export const PRIVACY_TYPE_WHIRLPOOL_DESTINATION = 'whirlpool destination';

// Workspace node type
export const WORKSPACE_NODE_TYPE_TRANSACTION = 'transaction';
export const WORKSPACE_NODE_TYPE_CLUSTER = 'cluster';
export const WORKSPACE_NODE_TYPE_SELECTOR = 'selector';
export const WORKSPACE_NODE_TYPE_NOTE = 'note';

// Selector type
export const SELECTOR_TYPE_HEURISTIC = 'heuristic';
export const SELECTOR_TYPE_TX_PROP = 'transactionProperties';
export const SELECTOR_TYPE_TX_GRAPH = 'transactionGraph';

// Selector status
export const SELECTOR_STATUS_WAITING = 'waiting';
export const SELECTOR_STATUS_ERROR = 'error';
export const SELECTOR_STATUS_SUCCESS = 'success';

export const SELECTOR_RESULT_LIMIT = 20_000;

// Selector error code
export const SELECTOR_ERROR_CODE_RESULT_LIMIT_EXCEEDED = 'result_limit_exceeded';

export const SELECTOR_MAX_ITEMS = 200;
export const CLUSTER_MAX_OUTPUTS = 200_000;

// Wasabi 2.0 denominations
/* eslint-disable @stylistic/indent, @stylistic/array-element-newline -- disable so multiple number can stay on the same line */
export const DENOMINATIONS_WASABI2 = new Set([5000, 6561, 8192, 10_000, 13_122, 16_384, 19_683, 20_000,
  32_768, 39_366, 50_000, 59_049, 65_536, 100_000, 118_098, 131_072, 177_147, 200_000, 262_144, 354_294, 500_000,
  524_288, 531_441, 1_000_000, 1_048_576, 1_062_882, 1_594_323, 2_000_000, 2_097_152, 3_188_646, 4_194_304, 4_782_969,
  5_000_000, 8_388_608, 9_565_938, 10_000_000, 14_348_907, 16_777_216, 20_000_000, 28_697_814, 33_554_432, 43_046_721,
  50_000_000, 67_108_864, 86_093_442, 100_000_000, 129_140_163, 134_217_728, 200_000_000, 258_280_326, 268_435_456,
  387_420_489, 500_000_000, 536_870_912, 774_840_978, 1_000_000_000, 1_073_741_824, 1_162_261_467, 2_000_000_000,
  2_147_483_648, 2_324_522_934, 3_486_784_401, 4_294_967_296, 5_000_000_000, 6_973_568_802, 8_589_934_592,
  10_000_000_000, 10_460_353_203, 17_179_869_184, 20_000_000_000, 20_920_706_406, 31_381_059_609, 34_359_738_368,
  50_000_000_000, 62_762_119_218, 68_719_476_736, 94_143_178_827, 100_000_000_000, 137_438_953_472]);
/* eslint-enable @stylistic/indent, @stylistic/array-element-newline */

// ID of the most recent update. This will be stored in localstorage to detect if a user has viewed the most recent update.
export const UPDATE_ID = 1;
