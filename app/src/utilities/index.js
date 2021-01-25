/* eslint-disable no-mixed-operators */
import { ROUTE_NAME_LOGIN_PAGE } from '../constants';

export function resetData(context) {
  context.$store.dispatch('resetMsg');
  context.$store.dispatch('setBlockData', null);
  context.$store.dispatch('setTransactionData', null);
  context.$store.dispatch('setAddressData', null);
  context.$store.dispatch('setHeuristicData', null);
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
export function isInvalidTokenMsg(msg, router) {
  if (msg.invalidToken !== undefined && msg.invalidToken === true) {
    router.push({ name: ROUTE_NAME_LOGIN_PAGE });
    return true;
  }
  return false;
}

export function doPost(route, router, body, parameter) {
  return fetch(route + parameter, {
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
      if (isInvalidTokenMsg(data, router)) throw Error();
      return data;
    });
}

export function doGet(route, router, parameter) {
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
      if (isInvalidTokenMsg(data, router)) throw Error();
      return data;
    });
}

export function handleError(context, error) {
  let errMsg;
  if (error.message === '500 Internal Server Error') {
    errMsg = 'Server is not reachable';
  } else {
    errMsg = `Error getting data: ${error}`;
  }

  context.dispatch('setErrorMsg', errMsg);
}

export const emailRules = [
  (v) => !!v || 'E-mail is required',
  (v) => (v && v.length < 100) || 'E-mail must be less than 100 characters',
  (v) => /.+@.+\..+/.test(v) || 'E-mail must be valid',
];
