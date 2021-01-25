<template>
    <v-layout row>
        <TxLookup/>
        <AddressLookup/>
        <BlockLookup/>
    </v-layout>
</template>

<script>
import TxLookup from './data/TxLookup.vue';
import AddressLookup from './data/AddressLookup.vue';
import BlockLookup from './data/BlockLookup.vue';

export default {
  name: 'SearchView',
  components: {
    TxLookup,
    AddressLookup,
    BlockLookup,
  },
};
</script>

<style scoped>

</style>
