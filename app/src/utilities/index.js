import {
	BLOCKCHAIN_BTC,
	BLOCKCHAIN_DASH,
	CLUSTER_TYPE_CUSTOM,
	CLUSTER_TYPE_FMI,
	LOCALSTORAGE_FIELD_SESSION,
	LOCALSTORAGE_FIELD_SETTINGS,
	PRIVACY_TYPE_DESTINATION,
	PRIVACY_TYPE_WASABI_2_DESTINATION,
	ROUTE_NAME_LOGIN_PAGE,
	DENOMINATIONS_WASABI2, RESPONSE_TYPE_TRANSACTION, RESPONSE_TYPE_BLOCK, RESPONSE_TYPE_ADDRESS,
} from '@/constants';
import {inject} from 'vue';

export function setLocalSession(sessionData) {
	localStorage.setItem(LOCALSTORAGE_FIELD_SESSION, JSON.stringify(sessionData));
}

export function deleteLocalSession() {
	localStorage.removeItem(LOCALSTORAGE_FIELD_SESSION);
}

export function getLocalSession() {
	let localStorageSessionData = localStorage.getItem(LOCALSTORAGE_FIELD_SESSION);
	if (localStorageSessionData !== null) {
		localStorageSessionData = JSON.parse(localStorageSessionData);
	}

	return localStorageSessionData;
}

export function setLocalSettings(settingsData) {
	localStorage.setItem(LOCALSTORAGE_FIELD_SETTINGS, JSON.stringify(settingsData));
}

export function getLocalSettings() {
	let localStorageSettingsData = localStorage.getItem(LOCALSTORAGE_FIELD_SETTINGS);
	if (localStorageSettingsData !== null) {
		localStorageSettingsData = JSON.parse(localStorageSettingsData);
	}

	return localStorageSettingsData;
}

export function getCoinUnit(mode) {
	switch (mode) {
		case BLOCKCHAIN_DASH: return 'Dash';
		case BLOCKCHAIN_BTC: return 'BTC';
		default: return 'invalid_unit';
	}
}

export function getDakarClient(mode) {
	switch (mode) {
		case BLOCKCHAIN_DASH: return inject('dashdakar');
		case BLOCKCHAIN_BTC: return inject('btcdakar');
		default: throw new Error('invalid blockchain mode:', mode);
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

	return `${hash.substring(0, elementLen)}...${hash.substring(hash.length - elementLen, hash.length)}`;
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
		if (response.status === 400) {
			errMsg = 'invalid request';
		} else if (response.status === 404) {
			errMsg = 'resource not found';
		} else {
			errMsg = `${response.status} ${response.statusText}`;
		}
	}

	throw new Error(errMsg, {cause: response});
}

export function handleError(context, error) {
	let errMsg;
	if (error.cause?.status === 500) {
		errMsg = 'Error requesting data from server. Please try again later.';
	} else {
		errMsg = error.message;
	}

	context.addMessage({
		text: errMsg, type: 'error', temporary: true, category: context.$route.name,
	});
}

export const emailRules = [
	v => Boolean(v) || 'E-mail is required',
	v => (v && v.length < 100) || 'E-mail must be less than 100 characters',
	v => /.+@.+\..+/.test(v) || 'E-mail must be valid',
];

export const fileRule = [v => {
	if (!v) {
		return false;
	}

	return v.length > 0 || 'File is required';
}];

function isRole(session, mode, roleName) {
	switch (mode) {
		case BLOCKCHAIN_BTC: return Boolean(session?.identity?.metadata_public?.roles?.dakar_btc === roleName);
		case BLOCKCHAIN_DASH: return Boolean(session?.identity?.metadata_public?.roles?.dakar_dash === roleName);
		default: return false;
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

// GetClusterTypeLabel translates the cluster shorthand of cluster types to a readable string
export function getClusterTypeLabel(clusterType) {
	switch (clusterType) {
		case CLUSTER_TYPE_FMI:
			return 'Multi-Input Cluster';
		case CLUSTER_TYPE_CUSTOM:
			return 'User-defined Cluster';
		default:
			return clusterType;
	}
}

// Returns true if the provided transaction type is destination
export function isDestination(type) {
	return type === PRIVACY_TYPE_DESTINATION || type === PRIVACY_TYPE_WASABI_2_DESTINATION;
}

// Returns true if the provided argument is a function
export function isFunction(functionToCheck) {
	if (!functionToCheck) {
		return false;
	}

	const fnType = {}.toString.call(functionToCheck);
	return fnType === '[object Function]' || fnType === '[object AsyncFunction]';
}

// Appends an 's' at the end of subject if count is higher than one
export function plural(subject, count) {
	return count === 0 || count > 1 ? `${subject}s` : subject;
}

// Returns a mapping between transaction types and their colors.
// If a blockchain mode is provided, only transaction types of the given mode are returned.
export function getColorMap(mode) {
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
	// Set color for transaction without type
	colorMap.set(undefined, '#607D8B');

	switch (mode) {
		case BLOCKCHAIN_DASH:
			dashTransactionTypes.forEach(t => colorMap.set(t.name, t.color));
			break;
		case BLOCKCHAIN_BTC:
			wasabi2TransactionTypes.forEach(t => colorMap.set(t.name, t.color));
			whirlPoolTransactionTypes.forEach(t => colorMap.set(t.name, t.color));
			break;
		case undefined:
			dashTransactionTypes.forEach(t => colorMap.set(t.name, t.color));
			wasabi2TransactionTypes.forEach(t => colorMap.set(t.name, t.color));
			whirlPoolTransactionTypes.forEach(t => colorMap.set(t.name, t.color));
			break;
		default:
	}

	return colorMap;
}

// Capitalize returns the first letter of each word (separated by a space) in str capitalized
export function capitalize(str) {
	return str.split(' ').map(d => d[0].toUpperCase() + d.slice(1)).join(' ');
}

// IsWasabi2Denomination returns true if the given amount is a wasabi 2.0 denomination
export function isWasabi2Denomination(amount) {
	return DENOMINATIONS_WASABI2.has(amount);
}

// IsUncommonWasabi2Denomination returns true if the given amount is a uncommon wasabi 2.0 denomination
export function isUncommonWasabi2Denomination(amount) {
	return amount % 5000 !== 0 && DENOMINATIONS_WASABI2.has(amount);
}

async function storeResult(promise, piniaAction) {
	try {
		const response = await promise;
		piniaAction(response);
	} catch (e) {
		return e;
	}

	return undefined;
}

export async function handleQuery(q, explorerStore, client, type) {
	explorerStore.setAddress(null);
	explorerStore.setBlock(null);
	explorerStore.setTransaction(null);

	switch (type) {
		case RESPONSE_TYPE_TRANSACTION:
			return await storeResult(client.data.blockchainTransactionsHashGet({hash: q}), explorerStore.updateTransaction);
		case RESPONSE_TYPE_BLOCK:
			return await storeResult(client.data.blockchainBlocksHashGet({hash: q}), explorerStore.updateBlock);
		case RESPONSE_TYPE_ADDRESS:
			return await storeResult(client.data.blockchainAddressesHashGet({hash: q}), explorerStore.updateAddress);
		default:
			return await storeResult(client.data.blockchainSearchQueryGet({query: q}), explorerStore.updateSearchResult);
	}
}
