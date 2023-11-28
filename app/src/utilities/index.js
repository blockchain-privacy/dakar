import {
	CLUSTER_TYPE_CUSTOM,
	CLUSTER_TYPE_FMI,
	LOCALSTORAGE_FIELD_SESSION,
	LOCALSTORAGE_FIELD_SETTINGS,
	ROUTE_NAME_LOGIN_PAGE,
} from '@/constants';

export function setLocalSession(sessionData) {
	localStorage.setItem(LOCALSTORAGE_FIELD_SESSION, JSON.stringify(sessionData));
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

export function shortenHash(hash) {
	const elementLen = 17;

	if (hash.length < elementLen * 2 + 3) {
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

// GetCurrentDate returns the current date as a string in the form dd-mm-yyyy
export function getCurrentDate() {
	const now = new Date();
	const dd = String(now.getDate()).padStart(2, '0');
	const mm = String(now.getMonth() + 1).padStart(2, '0'); // January is 0!
	const yyyy = now.getFullYear();
	return `${dd}-${mm}-${yyyy}`;
}

// CheckResponseStatus throws an error depending on the provided response status
export async function checkResponseStatus(context, response) {
	if (response.ok) {
		return;
	}

	if (response.status === 401) {
		handleUnauthorizedRequest(context.$router, context.$store, context.$route);
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

// HandleUnauthorizedRequest directs to the login page
export function handleUnauthorizedRequest(router, store, currentRoute) {
	// Set failed route so we can reroute to it later
	store.dispatch('setFailedRoute', currentRoute);
	store.dispatch('setSession', null);
	router.push({name: ROUTE_NAME_LOGIN_PAGE});
}

// IsSessionExpired returns true if the session has expired
export function isSessionExpired(session) {
	return !session || !session.expires_at
      || new Date() > new Date(session.expires_at);
}

export function handleError(context, error) {
	let errMsg;
	if (error.cause?.status === 500) {
		errMsg = 'Error requesting data from server. Please try again later.';
	} else {
		errMsg = error.message;
	}

	context.$store.dispatch('addMessage', {text: errMsg, type: 'error', temporary: true, category: context.$route.name});
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

// IsValidQueryInput returns true if the input query is valid. This function
// should be used instead of isValidQuery if the input is expected to be trimmed.
export function isValidQueryInput(str) {
	const inputLen = str.length;
	// 64 -> length of transaction hash and block hash
	if (inputLen === 0 || inputLen > 64) {
		return false;
	}

	// 33,34 -> address length; if smaller than it must be a block id
	if (inputLen < 33) {
		return Number.isInteger(Number(str));
	}

	return str.match(/^[\da-zA-Z]+$/) !== null;
}

// IsValidQuery returns true if the input query is valid. This function should
// be used instead of isValidQueryInput if the input is not expected to be trimmed.
export function isValidQuery(str) {
	const trimmed = str.trim();

	return trimmed.length === 0 ? true : isValidQueryInput(trimmed);
}

function isRole(session, roleName) {
	return Boolean(session && session.identity && session.identity.metadata_public
      && session.identity.metadata_public.roles
      && session.identity.metadata_public.roles.some(d => d === roleName));
}

export function isPrivilegedIdentity(session) {
	return isRole(session, 'privileged');
}

export function isAdminIdentity(session) {
	return isRole(session, 'admin');
}

// GetPrivacyTypeLabel translates the integer representation of privacy types to string
export function getPrivacyTypeLabel(privacyType) {
	const t = parseInt(privacyType, 10);

	if (Number.isNaN(t) || t < 0 || t > 499) {
		return '';
	}

	if (t <= 99) {
		return 'mixing';
	}

	if (t <= 199) {
		return 'destination';
	}

	if (t <= 299) {
		return 'origin';
	}

	if (t <= 399) {
		return 'collateral creation';
	}

	if (t <= 499) {
		return 'collateral payment';
	}

	return '';
}

// GetPrivacyTypeTooltip returns the corresponding tooltip path
export function getPrivacyTypeTooltip(privacyType) {
	const t = parseInt(privacyType, 10);
	const folder = 'transactionTypes';
	if (Number.isNaN(t) || t < 0 || t > 499) {
		return '';
	}

	if (t <= 99) {
		return `${folder}/mixingTransaction.md`;
	}

	if (t <= 199) {
		return `${folder}/destinationTransaction.md`;
	}

	if (t <= 299) {
		return `${folder}/originTransaction.md`;
	}

	if (t <= 399) {
		return `${folder}/collateralCreationTransaction.md`;
	}

	if (t <= 499) {
		return `${folder}/collateralPaymentTransaction.md`;
	}

	return '';
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

// IsMixing returns true if the provided privacyType is in the range of mixing transactions
export function isMixing(privacyType) {
	const t = parseInt(privacyType, 10);

	if (Number.isNaN(t) || t < 0) {
		return false;
	}

	return t <= 99;
}

// IsOrigin returns true if the provided privacyType is in the range of origin transactions
export function isOrigin(privacyType) {
	const t = parseInt(privacyType, 10);

	if (Number.isNaN(t) || t < 0) {
		return false;
	}

	return t >= 200 && t <= 299;
}

// IsDestination returns true if the provided privacyType is in the range of
// destination transactions
export function isDestination(privacyType) {
	const t = parseInt(privacyType, 10);

	if (Number.isNaN(t) || t < 0) {
		return false;
	}

	return t >= 100 && t <= 199;
}

// IsCollateralCreation returns true if the provided privacyType is in the range of
// collateral creation transactions
export function isCollateralCreation(privacyType) {
	const t = parseInt(privacyType, 10);

	if (Number.isNaN(t) || t < 0) {
		return false;
	}

	return t >= 300 && t <= 399;
}

// IsCollateralPayment returns true if the provided privacyType is in the range of
// collateral payment transactions
export function isCollateralPayment(privacyType) {
	const t = parseInt(privacyType, 10);

	if (Number.isNaN(t) || t < 0) {
		return false;
	}

	return t >= 400 && t <= 499;
}

// IsFunction returns true if the provided argument is a function
export function isFunction(functionToCheck) {
	if (!functionToCheck) {
		return false;
	}

	const fnType = {}.toString.call(functionToCheck);
	return fnType === '[object Function]' || fnType === '[object AsyncFunction]';
}

// Plural appends an 's' at the end of subject if count is higher than one
export function plural(subject, count) {
	return count > 1 ? `${subject}s` : subject;
}
