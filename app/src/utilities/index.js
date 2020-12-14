/* eslint-disable no-mixed-operators */
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
