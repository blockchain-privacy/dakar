/* eslint-disable no-mixed-operators */
import {
  LOCALSTORAGE_FIELD_USER,
  LOCALSTORAGE_FIELD_SETTINGS,
  PASSWORD_MAX_CHARACTERS,
  PASSWORD_MIN_CHARACTERS,
  ROUTE_NAME_LOGIN_PAGE, TOKEN_TIMEOUT,
  CLUSTER_TYPE_FMI, CLUSTER_TYPE_CUSTOM,
} from '../constants';

export function resetData(context) {
  context.$store.dispatch('resetMessages');
  context.$store.dispatch('setBlockData', null);
  context.$store.dispatch('setTransactionData', null);
  context.$store.dispatch('setAddressData', null);
}

// setActionDate sets the current time as the last action timestamp of the user data.
// This method should be used each time the frontend interacts with the backend.
export function setActionDate(userData) {
  userData.lastAction = new Date();
  return userData;
}

export function setLocalUser(userData) {
  localStorage.setItem(LOCALSTORAGE_FIELD_USER, JSON.stringify(userData));
}

export function getLocalUser() {
  let localStorageUserData = localStorage.getItem(LOCALSTORAGE_FIELD_USER);
  if (localStorageUserData !== null) localStorageUserData = JSON.parse(localStorageUserData);
  return localStorageUserData;
}

export function removeLocalUser() {
  return localStorage.removeItem(LOCALSTORAGE_FIELD_USER);
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

export function removeLocalSettings() {
  return localStorage.removeItem(LOCALSTORAGE_FIELD_SETTINGS);
}

export function resetLocal() {
  removeLocalUser();
  removeLocalSettings();
}

export function shortenHash(hash) {
  const elementLen = 17;

  if (hash.length < elementLen * 2 + 3) {
    return hash;
  }

  return `${hash.substring(0, elementLen)}...${hash.substring(hash.length - elementLen, hash.length)}`;
}

export function convertAmount(val) {
  return val / 1e8;
}

// getCurrentDate returns the current date as a string in the form dd-mm-yyyy
export function getCurrentDate() {
  const now = new Date();
  const dd = String(now.getDate()).padStart(2, '0');
  const mm = String(now.getMonth() + 1).padStart(2, '0'); // January is 0!
  const yyyy = now.getFullYear();
  return `${dd}-${mm}-${yyyy}`;
}

// isInvalidTokenMsg checks if the page should be rerouted to the login page
function isInvalidTokenMsg(msg, router, store) {
  if (msg.invalidToken !== undefined && msg.invalidToken === true) {
    // set failed route so we can reroute to it later
    store.dispatch('setFailedRoute', router.history.current);
    store.dispatch('setActiveUser', null);
    router.push({ name: ROUTE_NAME_LOGIN_PAGE });
    return true;
  }
  return false;
}

// isTokenTimedOut returns true if the user session has timed out
export function isTokenTimedOut(userData) {
  return !userData || !userData.lastAction
      || new Date() - new Date(userData.lastAction) >= TOKEN_TIMEOUT;
}

export function doPost(route, router, store, body, parameter) {
  let para = '';
  if (parameter !== undefined) para = parameter;
  return fetch(route + para, {
    method: 'POST',
    credentials: 'same-origin',
    redirect: 'error',
    referrerPolicy: 'no-referrer',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  }).then((response) => {
    if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
    return response;
  }).then((response) => response.json())
    .then((data) => {
      if (isInvalidTokenMsg(data, router, store)) throw Error('Please login again.');
      // update last action time stamp
      const userData = store.getters.getActiveUser;
      if (userData) {
        store.dispatch('setActiveUser', setActionDate(userData));
      }

      return data;
    });
}

export function doPostUpload(route, router, store, body, parameter) {
  let para = '';
  if (parameter !== undefined) para = parameter;
  return fetch(route + para, {
    method: 'POST',
    credentials: 'same-origin',
    redirect: 'error',
    referrerPolicy: 'no-referrer',
    body,
  }).then((response) => {
    if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
    return response;
  }).then((response) => response.json())
    .then((data) => {
      if (isInvalidTokenMsg(data, router, store)) throw Error('Please login again.');
      // update last action time stamp
      const userData = store.getters.getActiveUser;
      if (userData) {
        store.dispatch('setActiveUser', setActionDate(userData));
      }

      return data;
    });
}

export function doPostBlob(route, router, store, body, parameter) {
  let para = '';
  if (parameter !== undefined) para = parameter;
  return fetch(route + para, {
    method: 'POST',
    credentials: 'same-origin',
    redirect: 'error',
    referrerPolicy: 'no-referrer',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  }).then((response) => {
    if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
    return response;
  }).then((response) => response.blob())
    .then((data) => {
      if (isInvalidTokenMsg(data, router, store)) throw Error('Please login again.');
      // update last action time stamp
      const userData = store.getters.getActiveUser;
      if (userData) {
        store.dispatch('setActiveUser', setActionDate(userData));
      }

      return data;
    });
}

export function doGet(route, router, store, parameter) {
  let para = '';
  if (parameter !== undefined) para = parameter;
  return fetch(route + para, {
    credentials: 'same-origin',
  })
    .then((response) => {
      if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
      return response;
    })
    .then((response) => response.json())
    .then((data) => {
      if (isInvalidTokenMsg(data, router, store)) throw Error('Please login again.');
      // update last action time stamp
      const userData = store.getters.getActiveUser;
      if (userData) {
        store.dispatch('setActiveUser', setActionDate(userData));
      }

      return data;
    });
}

export function doGetBlob(route, router, store, parameter) {
  let para = '';
  if (parameter !== undefined) para = parameter;
  return fetch(route + para, {
    credentials: 'same-origin',
  })
    .then((response) => {
      if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
      return response;
    })
    .then((response) => response.blob())
    .then((data) => {
      if (isInvalidTokenMsg(data, router, store)) throw Error('Please login again.');
      // update last action time stamp
      const userData = store.getters.getActiveUser;
      if (userData) {
        store.dispatch('setActiveUser', setActionDate(userData));
      }

      return data;
    });
}

export function handleError(context, error) {
  let errMsg;
  if (error.message === '500 Internal Server Error') {
    errMsg = 'Server is not reachable';
  } else {
    errMsg = error.message;
  }

  context.dispatch('addMessage', { text: errMsg, type: 'error', temporary: true });
}

export const emailRules = [
  (v) => !!v || 'E-mail is required',
  (v) => (v && v.length < 100) || 'E-mail must be less than 100 characters',
  (v) => /.+@.+\..+/.test(v) || 'E-mail must be valid',
];

const notAllowedWhitespaceCharacters = [
  '\b', '\t', '\n', '\v', '\f', '\r',
  '\u0008', '\u0009', '\u000A', '\u000B', '\u000C',
  '\u000D', '\u0022', '\u0027', '\u005C',
  '\u00A0', '\u2028', '\u2029', '\uFEFF'];

// hasWhitespace checks if the given string
// contains any of the characters in notAllowedWhitespaceCharacters
// credit: https://stackoverflow.com/questions/1731190/check-if-a-string-has-white-space
const hasWhitespace = (char) => notAllowedWhitespaceCharacters.some(
  (w) => char.indexOf(w) > -1,
  notAllowedWhitespaceCharacters,
);

export const passwordRules = [
  (v) => !!v || 'Password is required',
  (v) => !hasWhitespace(v) || 'Password contains white space characters',
  (v) => v.length >= PASSWORD_MIN_CHARACTERS || `At least ${PASSWORD_MIN_CHARACTERS} characters`,
  (v) => (v && v.length < PASSWORD_MAX_CHARACTERS)
        || `Password must be less than ${PASSWORD_MAX_CHARACTERS} characters`,
];

// isValidQueryInput returns true if the input query is valid. This function
// should be used instead of isValidQuery if the input is expected to be trimmed.
export function isValidQueryInput(str) {
  const inputLen = str.length;
  // 64 -> length of transaction hash and block hash
  if (inputLen === 0 || inputLen > 64) return false;

  // 33,34 -> address length; if smaller than it must be a block id
  if (inputLen < 33) {
    return Number.isInteger(Number(str));
  }

  return str.match(/^[0-9a-zA-Z]+$/) !== null;
}

// isValidQuery returns true if the input query is valid. This function should
// be used instead of isValidQueryInput if the input is not expected to be trimmed.
export function isValidQuery(str) {
  const trimmed = str.trim();

  return trimmed.length === 0 ? true : isValidQueryInput(trimmed);
}

function isRole(userData, roleName) {
  return userData && userData.roles && userData.roles.some((d) => d.name === roleName);
}

export function isPrivilegedUser(userData) {
  return isRole(userData, 'privileged');
}

export function isAdminUser(userData) {
  return isRole(userData, 'admin');
}

// getPrivacyTypeLabel translates the integer representation of privacy types to string
export function getPrivacyTypeLabel(privacyType) {
  const t = parseInt(privacyType, 10);

  if (Number.isNaN(t) || t < 0 || t > 499) return '';
  if (t <= 99) return 'mixing';
  if (t <= 199) return 'destination';
  if (t <= 299) return 'origin';
  if (t <= 399) return 'collateral creation';
  if (t <= 499) return 'collateral payment';

  return '';
}

// getMixingDenomination translates the integer representation of privacy types to string
export function getMixingLabel(privacyType) {
  const t = parseInt(privacyType, 10);

  if (Number.isNaN(t) || t < 0 || t > 99) return -1;
  if (t <= 5) return 0;
  if (t <= 10) return 1;
  if (t <= 15) return 2;
  if (t <= 20) return 3;
  if (t <= 25) return 4;

  return -1;
}

// getClusterTypeLabel translates the cluster shorthand of cluster types to a readable string
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

// isMixing returns true if the provided privacyType is in the range of mixing transactions
export function isMixing(privacyType) {
  const t = parseInt(privacyType, 10);

  if (Number.isNaN(t) || t < 0) return false;
  return t <= 99;
}

// isOrigin returns true if the provided privacyType is in the range of origin transactions
export function isOrigin(privacyType) {
  const t = parseInt(privacyType, 10);

  if (Number.isNaN(t) || t < 0) return false;
  return t >= 200 && t <= 299;
}

// isDestination returns true if the provided privacyType is in the range of
// destination transactions
export function isDestination(privacyType) {
  const t = parseInt(privacyType, 10);

  if (Number.isNaN(t) || t < 0) return false;
  return t >= 100 && t <= 199;
}

// isCollateralCreation returns true if the provided privacyType is in the range of
// collateral creation transactions
export function isCollateralCreation(privacyType) {
  const t = parseInt(privacyType, 10);

  if (Number.isNaN(t) || t < 0) return false;
  return t >= 300 && t <= 399;
}

// isCollateralPayment returns true if the provided privacyType is in the range of
// collateral payment transactions
export function isCollateralPayment(privacyType) {
  const t = parseInt(privacyType, 10);

  if (Number.isNaN(t) || t < 0) return false;
  return t >= 400 && t <= 499;
}

// uuidv4 generates a pseudo random unique id
// credits:
// - https://stackoverflow.com/questions/105034/how-to-create-a-guid-uuid
// - https://gist.github.com/jed/982883
// replace with https://developer.mozilla.org/en-US/docs/Web/API/Crypto/randomUUID#browser_compatibility
// when more browser provide compatibility
export function uuidv4() {
  return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, (c) => (
  // eslint-disable-next-line no-bitwise
    c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4)
    .toString(16));
}
