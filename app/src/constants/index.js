// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

// Search responses
import {bitcoinLogo, dashLogo} from '@/customIcons/index.js';

export const RESPONSE_EMPTY = 'response_empty';
export const RESPONSE_TYPE_TRANSACTION = 'tx';
export const RESPONSE_TYPE_ADDRESS = 'addr';
export const RESPONSE_TYPE_BLOCK = 'block';

// Frontend route names
export const ROUTE_NAME_ENTRY_PAGE = 'Entry Page';
export const ROUTE_NAME_STATUS_PAGE = 'Status Page';
export const ROUTE_NAME_404_PAGE = 'Page not found';
export const ROUTE_NAME_NO_RESULTS = 'No results found';
export const ROUTE_NAME_ERROR = 'Error';
export const ROUTE_NAME_LOGIN_PAGE = 'Login Page';
export const ROUTE_NAME_ACCOUNT_RECOVERY = 'Account Recovery Page';
export const ROUTE_NAME_USER_ADMIN_PAGE = 'User Administration Page';
export const ROUTE_NAME_USER_PROFILE_PAGE = 'User Profile Page';
export const ROUTE_NAME_WORKSPACE_PAGE = 'Workspace Editor';
export const ROUTE_NAME_BLOCK_PAGE = 'Block Page';
export const ROUTE_NAME_TRANSACTION_PAGE = 'Transaction Page';
export const ROUTE_NAME_ADDRESS_PAGE = 'Address Page';
export const ROUTE_NAME_WORKSPACES_PAGE = 'Workspace Lookup Page';
export const ROUTE_NAME_CLUSTER_OVERVIEW = 'Custom Clusters Page';
export const ROUTE_NAME_ATTRIBUTIONS = 'Attributions Page';
export const ROUTE_NAME_ADDRESS_EXCLUSIONS = 'Address Exclusions Page';
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

export const SELECTOR_MAX_ITEMS = 200;
export const CLUSTER_MAX_OUTPUTS = 200_000;

// Wasabi 2.0 denominations
/* eslint-disable @stylistic/indent, @stylistic/array-element-newline */
export const DENOMINATIONS_WASABI2 = new Set([5000, 6561, 8192, 10000, 13122, 16384, 19683, 20000,
  32768, 39366, 50000, 59049, 65536, 100000, 118098, 131072, 177147, 200000, 262144, 354294, 500000, 524288, 531441,
  1000000, 1048576, 1062882, 1594323, 2000000, 2097152, 3188646, 4194304, 4782969, 5000000, 8388608, 9565938,
  10000000, 14348907, 16777216, 20000000, 28697814, 33554432, 43046721, 50000000, 67108864, 86093442, 100000000,
  129140163, 134217728, 200000000, 258280326, 268435456, 387420489, 500000000, 536870912, 774840978, 1000000000,
  1073741824, 1162261467, 2000000000, 2147483648, 2324522934, 3486784401, 4294967296, 5000000000, 6973568802,
  8589934592, 10000000000, 10460353203, 17179869184, 20000000000, 20920706406, 31381059609, 34359738368, 50000000000,
  62762119218, 68719476736, 94143178827, 100000000000, 137438953472]);
/* eslint-enable @stylistic/indent, @stylistic/array-element-newline */
