<template>
    <v-layout row>
        <TxLookup/>
        <AddressLookup/>
        <BlockLookup/>
    </v-layout>
</template>

<script>
    import TxLookup from "./TxLookup";
    import AddressLookup from "./AddressLookup";
    import BlockLookup from "./BlockLookup";

    export default {
        name: "SearchView",
        components: {
            TxLookup,
            AddressLookup,
            BlockLookup
        }
    }
</script>

<style scoped>

</style>