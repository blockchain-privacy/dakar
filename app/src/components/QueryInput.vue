<template>
    <v-form v-on:submit.prevent="handleQuery(query)" class="d-flex mx-auto">
        <v-text-field class="d-flex" full-width v-model="query"
                      label="Search for transactions and addresses"/>
    </v-form>
</template>

<script>
    export default {
        name: "QueryInput",
        data() {
            return {
                // query is not managed by the vuex store
                // as it only needs to be accessed by this component
                query: ""
            };
        },
        computed: {
            errorMsg: {
                get() {
                    return this.$store.getters.getErrorMsg;
                },
                set(value) {
                    this.$store.dispatch('setErrorMsg', value);
                }
            },
            warningMsg: {
                get() {
                    return this.$store.getters.getWarningMsg;
                },
                set(value) {
                    this.$store.dispatch('setWarningMsg', value);
                }
            },
            transaction: {
                get() {
                    return this.$store.getters.getTransactionData;
                },
                set(value) {
                    this.$store.dispatch('setTransactionData', value);
                }
            },
            address: {
                get() {
                    return this.$store.getters.getAddressData;
                },
                set(value) {
                    this.$store.dispatch('setAddressData', value);
                }
            },
        },
        methods: {
            handleQuery: function (q) {
                this.resetData();
                if (!this.isValidData(q)) {
                    this.warningMsg = "Input was not valid!";
                    return;
                }

                this.searchTx(q).catch(() => {
                    // if transaction query fails, search for address
                    this.searchAddress(q)
                });
            },
            isValidData: function  (str) {
                // TODO: check if str is address or transaction, also calculate checksum
                return str.length >= 34;
            },
            resetData: function () {
                this.query = ""
                this.$store.dispatch('resetMsg');
                this.transaction = null;
                this.address = null;
            },
            searchTx: function (q) {
                console.log("Tx search: " + q);
                return fetch("/tx/" + q)
                    .then(response => {
                        if (!response.ok) throw new Error(response.status + " " + response.statusText)
                        return response
                    })
                    .then(response => response.json())
                    .then(data => {
                        this.transaction = data;
                    });
            },
            searchAddress: function (q) {
                console.log("Address search: " + q);
                fetch("/address/" + q)
                    .then(response => {
                        if (!response.ok) throw new Error(response.status + " " + response.statusText)
                        return response
                    })
                    .then(response => response.json())
                    .then(data => {
                        this.address = data;
                    })
                    .catch(error => {
                        this.errorMsg = error;
                    });
            }
        }
    }
</script>

<style scoped>

</style>