/* eslint-disable no-mixed-operators */
export function resetData(context) {
  context.$store.dispatch('resetMsg');
  context.$store.dispatch('setBlockData', null);
  context.$store.dispatch('setTransactionData', null);
  context.$store.dispatch('setAddressData', null);
  context.$store.dispatch('setHeuristicData', null);
}

export function shortenHash(hash) {
  const elementLen = 17 * 2 + 3;

  if (hash.length < elementLen) {
    return hash;
  }

  return `${hash.substring(0, elementLen)}...${hash.substring(hash.length - elementLen, hash.length)}`;
}

export function convertAmount(val) {
  return val / 1e8;
}
