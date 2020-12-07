<template>
    <v-layout row>
        <TxLookup/>
        <AddressLookup/>
        <BlockLookup/>
    </v-layout>
</template>

<script>
import TxLookup from './TxLookup.vue';
import AddressLookup from './AddressLookup.vue';
import BlockLookup from './BlockLookup.vue';

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
