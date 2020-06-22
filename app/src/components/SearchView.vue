<template>
    <v-layout row>
        <TxLookup/>
        <AddressLookup/>
    </v-layout>
</template>

<script>
    import TxLookup from "./TxLookup";
    import AddressLookup from "./AddressLookup";

    export default {
        name: "SearchView",
        components: {
            TxLookup,
            AddressLookup,
        }
    }
</script>

<style scoped>

</style>