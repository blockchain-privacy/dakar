// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

import {inject} from 'vue';
import {
	BLOCKCHAIN_BTC,
	BLOCKCHAIN_DASH,
	CLUSTER_TYPE_CUSTOM,
	CLUSTER_TYPE_FMI,
	DENOMINATIONS_WASABI2,
	PRIVACY_TYPE_DESTINATION,
	PRIVACY_TYPE_WASABI_2_DESTINATION,
	PRIVACY_TYPE_WHIRLPOOL_DESTINATION,
	ROUTE_NAME_LOGIN_PAGE,
	SELECTOR_STATUS_SUCCESS,
	SELECTOR_TYPE_HEURISTIC,
	SELECTOR_TYPE_TX_GRAPH,
	SELECTOR_TYPE_TX_PROP,
	WORKSPACE_NODE_TYPE_CLUSTER,
	WORKSPACE_NODE_TYPE_SELECTOR,
	WORKSPACE_NODE_TYPE_TRANSACTION,
} from '@/constants';

export function getCoinUnit(mode) {
	switch (mode) {
		case BLOCKCHAIN_DASH: {return 'Dash';}

		case BLOCKCHAIN_BTC: {return 'BTC';}

		default: {return 'invalid_unit';}
	}
}

export function getDakarClient(mode) {
	switch (mode) {
		case BLOCKCHAIN_DASH: {return inject('dashdakar');}

		case BLOCKCHAIN_BTC: {return inject('btcdakar');}

		default: {throw new Error(`invalid blockchain mode: ${mode}`);}
	}
}

export function getDakarClients() {
	return {
		dash: inject('dashdakar'),
		btc: inject('btcdakar'),
	};
}

export function shortenHash(hash) {
	const elementLen = 17;

	if (hash.length < (elementLen * 2) + 3) {
		return hash;
	}

	return `${hash.slice(0, Math.max(0, elementLen))}...${hash.slice(hash.length - elementLen)}`;
}

// ConvertAmount returns the given integer divided by 100 000 000 and localized
export function convertAmount(val) {
	return (val / 1e8).toLocaleString(undefined, {
		maximumFractionDigits: 10,
	});
}

// AmountToIntegers returns the given number multipled by 100 000 000 and localized
export function amountToIntegers(val) {
	return Math.trunc(val * 1e8);
}

// GetCurrentDate returns the current date as a string in the form dd-mm-yyyy
export function getCurrentDate() {
	const now = new Date();
	const dd = String(now.getDate()).padStart(2, '0');
	const mm = String(now.getMonth() + 1).padStart(2, '0'); // January is 0!
	const yyyy = now.getFullYear();
	return `${dd}-${mm}-${yyyy}`;
}

// CheckResponseStatus throws an error depending on the provided response status
export async function checkResponseStatus(context, navStore, localStore, response) {
	if (response.ok) {
		return;
	}

	if (response.status === 401) {
		navStore.setFailedRoute(context.$route);
		localStore.setSession(null);
		context.$router.push({name: ROUTE_NAME_LOGIN_PAGE});
		throw new Error('Please login again.', {cause: response});
	}

	let errMsg = '';
	for (const e of response.headers.entries()) {
		if (e[0] === 'content-type') {
			if (e[1] === 'application/json') {
				// eslint-disable-next-line no-await-in-loop
				const jsonResponse = await response.json();
				if (jsonResponse.msg) {
					errMsg = jsonResponse.msg;
				}
			}

			break;
		}
	}

	if (errMsg === '') {
		switch (response.status) {
			case 400: {
				errMsg = 'Your request was invalid. Please try again.';
				break;
			}

			case 404: {
				errMsg = 'The resource you are trying to access is unavailable.';
				break;
			}

			case 500:
			case 502: {
				errMsg = 'Error requesting data from server. Please try again later.';
				break;
			}

			default: {
				errMsg = `${response.status} ${response.statusText}`;
			}
		}
	}

	throw new Error(errMsg, {cause: response});
}

export function handleError(context, error) {
	context.addMessage({
		text: error.message, type: 'error', temporary: true, category: context.$route.name,
	});
}

export const isValidEmail = v => /^\S[^\s@]*@\S[^\s.]*\.\S+$/v.test(v);
export const emailRules = [
	v => Boolean(v) || 'E-mail is required',
	v => (v && v.length < 100) || 'E-mail must be less than 100 characters',
	v => isValidEmail(v) || 'E-mail must be valid',
];

export const fileRule = [v => {
	if (!v) {
		return false;
	}

	return v.length > 0 || v instanceof File || 'File is required';
}];

function isRole(session, mode, roleName) {
	switch (mode) {
		case BLOCKCHAIN_BTC: {return Boolean(session?.identity?.metadata_public?.roles?.dakar_btc === roleName);}

		case BLOCKCHAIN_DASH: {return Boolean(session?.identity?.metadata_public?.roles?.dakar_dash === roleName);}

		default: {return false;}
	}
}

export function isModeBTC(mode) {
	return mode === BLOCKCHAIN_BTC;
}

export function isPrivilegedIdentity(session, mode) {
	return isRole(session, mode, 'privileged');
}

export function isAdminIdentity(session, mode) {
	return isRole(session, mode, 'admin');
}

export function isAnyPrivilegedIdentity(session) {
	return isPrivilegedIdentity(session, BLOCKCHAIN_BTC) || isPrivilegedIdentity(session, BLOCKCHAIN_DASH);
}

export function isAnyAdminIdentity(session) {
	return isAdminIdentity(session, BLOCKCHAIN_BTC) || isAdminIdentity(session, BLOCKCHAIN_DASH);
}

// GetClusterTypeLabel translates the cluster shorthand of cluster types to a readable string
export function getClusterTypeLabel(clusterType) {
	switch (clusterType) {
		case CLUSTER_TYPE_FMI: {
			return 'Multi-Input Cluster';
		}

		case CLUSTER_TYPE_CUSTOM: {
			return 'User-defined Cluster';
		}

		default: {
			return clusterType;
		}
	}
}

// Returns all descriptors which can be applied to at least one of the provided nodes
// if validate is true (default), then an error is thrown if an invalid node is passed
export function filterDescriptors(descriptors, nodes, validate = true) {
	if (!descriptors?.length || !nodes?.length) {
		return [];
	}

	const allowedDescriptors = [];

	for (const d of descriptors) {
		for (const n of nodes) {
			let nodeType;
			switch (n.type) {
				case WORKSPACE_NODE_TYPE_SELECTOR: {
					if (n.selectorType !== SELECTOR_TYPE_HEURISTIC || n.selectorStatus !== SELECTOR_STATUS_SUCCESS) {
						// eslint-disable-next-line max-depth
						if (validate) {
							throw new Error('invalid node type');
						}

						continue;
					}

					nodeType = n.heuristicOptions.type;
					break;
				}

				case WORKSPACE_NODE_TYPE_TRANSACTION: {
					nodeType = n.txtype;
					break;
				}

				default: {
					if (validate) {
						throw new Error('invalid node type');
					}
				}
			}

			if (d.allowedParents.includes(nodeType)) {
				allowedDescriptors.push(d);
				break;
			}
		}
	}

	return allowedDescriptors;
}

// Returns true if the provided transaction type is a destination
export function isDestination(type) {
	return type === PRIVACY_TYPE_DESTINATION || type === PRIVACY_TYPE_WASABI_2_DESTINATION
		|| type === PRIVACY_TYPE_WHIRLPOOL_DESTINATION;
}

// Returns the caption of the given heuristic type
export function getCoinJoinTypeCaption(heuristicType) {
	if (heuristicType.startsWith('whirlpool')) {
		return 'Whirlpool';
	}

	if (heuristicType.startsWith('wasabi2')) {
		return 'Wasabi 2.0';
	}

	return 'Dash';
}

// Returns true if the provided argument is a function
export function isFunction(functionToCheck) {
	if (!functionToCheck) {
		return false;
	}

	const fnType = Object.prototype.toString.call(functionToCheck);
	return fnType === '[object Function]' || fnType === '[object AsyncFunction]';
}

// Appends an 's' at the end of subject if count is higher than one
export function plural(subject, count) {
	return count === 0 || count > 1 ? `${subject}s` : subject;
}

// Appends an 's' at the end of subject if count is higher than one
export function pluralIrregular(subject, pluralVersion, count) {
	return count === 0 || count > 1 ? pluralVersion : subject;
}

// Returns a mapping between transaction types and their colors.
// If a blockchain mode is provided, only transaction types of the given mode are returned.
export function getTransactionColorMap(mode) {
	// Colors from https://sashamaps.net/docs/resources/20-colors/
	const dashTransactionTypes = [
		{name: 'origin', color: '#800000'},
		{name: 'mixing', color: '#e6194b'},
		{name: 'destination', color: '#fabed4'},
		{name: 'collateral creation', color: '#3cb44b'},
		{name: 'collateral payment', color: '#bfef45'},
	];

	const whirlPoolTransactionTypes = [
		{name: 'whirlpool origin', color: '#800000'},
		{name: 'whirlpool mixing', color: '#e6194b'},
		{name: 'whirlpool destination', color: '#fabed4'},
	];

	const wasabi2TransactionTypes = [
		{name: 'wasabi 2.0 origin', color: '#3cb44b'},
		{name: 'wasabi 2.0 mixing', color: '#bfef45'},
		{name: 'wasabi 2.0 destination', color: '#45ef87'},
	];

	const colorMap = new Map();

	switch (mode) {
		case BLOCKCHAIN_DASH: {
			for (const t of dashTransactionTypes) {
				colorMap.set(t.name, t.color);
			}

			break;
		}

		case BLOCKCHAIN_BTC: {
			for (const t of wasabi2TransactionTypes) {
				colorMap.set(t.name, t.color);
			}

			for (const t of whirlPoolTransactionTypes) {
				colorMap.set(t.name, t.color);
			}

			break;
		}

		case undefined: {
			for (const t of dashTransactionTypes) {
				colorMap.set(t.name, t.color);
			}

			for (const t of wasabi2TransactionTypes) {
				colorMap.set(t.name, t.color);
			}

			for (const t of whirlPoolTransactionTypes) {
				colorMap.set(t.name, t.color);
			}

			break;
		}

		default:
	}

	return colorMap;
}

// Returns a mapping between all graph nodes and their colors.
// If a blockchain mode is provided, only transaction types of the given mode are returned.
export function getGraphColorMap(mode) {
	const colorMap = getTransactionColorMap(mode);
	colorMap.set(WORKSPACE_NODE_TYPE_CLUSTER, '#ffe119');
	setUndefinedTransactionColor(colorMap, WORKSPACE_NODE_TYPE_TRANSACTION);

	colorMap.set(SELECTOR_TYPE_HEURISTIC, '#4363d8');
	colorMap.set(SELECTOR_TYPE_TX_GRAPH, '#42d4f4');
	colorMap.set(SELECTOR_TYPE_TX_PROP, '#000075');
	return colorMap;
}

// SetUndefinedTransactionColor adds the color for transactions without transaction type for the given key
export function setUndefinedTransactionColor(colorMap, key) {
	// Set color for transaction without type
	colorMap.set(key, '#607D8B');
}

// Capitalize returns the first letter of each word (separated by a space) in str capitalized
export function capitalize(str) {
	return str.split(' ').map(d => d[0].toUpperCase() + d.slice(1)).join(' ');
}

// IsWasabi2Denomination returns true if the given amount is a wasabi 2.0 denomination
export function isWasabi2Denomination(amount) {
	return DENOMINATIONS_WASABI2.has(amount);
}

// IsUncommonWasabi2Denomination returns true if the given amount is an uncommon wasabi 2.0 denomination
export function isUncommonWasabi2Denomination(amount) {
	return amount % 5000 !== 0 && DENOMINATIONS_WASABI2.has(amount);
}

// Returns an array containing all bitcoin addresses and transaction hashes in text
export function extractEntities(text) {
	if (!text || !text.trim()) {
		return [];
	}

	const regexp = /[13X7][a-km-zA-HJ-NP-Z1-9]{25,34}|bc1[a-zA-HJ-NP-Z0-9]{39,59}|[a-fA-F0-9]{64}/gv;
	const matches = [...text.matchAll(regexp)].map(r => r[0]);
	return [...new Set(matches)];
}

export function isMaxLargerThanMin(obj) {
	// Compare undefined so zero check is handled at least line
	if (obj.min === undefined || obj.max === undefined) {
		return true;
	}

	return obj.max >= obj.min;
}
