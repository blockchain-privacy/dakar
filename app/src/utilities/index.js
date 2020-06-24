export function resetData(context) {
    context.$store.dispatch('resetMsg');
    context.$store.dispatch('setBlockData', null);
    context.$store.dispatch('setTransactionData', null);
    context.$store.dispatch('setAddressData', null);
}
