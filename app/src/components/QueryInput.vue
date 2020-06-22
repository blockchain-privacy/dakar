<template>
    <v-form v-on:submit.prevent="handleQuery(query,'user')" class="d-flex mx-auto">
        <v-text-field class="d-flex" full-width v-model="query"
                      label="Search for transactions and addresses"/>
    </v-form>
</template>

<script>
    function newRouting(context, id) {
        if (id === undefined) {
            return;
        }
        context.handleQuery(id, 'route');
    }

    export default {
        name: "QueryInput",
        data() {
            return {
                // query is not managed by the vuex store
                // as it only needs to be accessed by this component
                query: "",
                lastQuery: ""
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
            block: {
                get() {
                    return this.$store.getters.getBlockData;
                },
                set(value) {
                    this.$store.dispatch('setBlockData', value);
                }
            },
        },
        methods: {
            handleQuery: function (q, origin) {
                if (origin === 'user' && q !== this.lastQuery) {
                    // update route only when input is from user and query is different
                    this.$router.push({name: 'Search Page', params: {id: q}});
                } else if (origin === 'route') {
                    // do nothing -> route is already up to date
                }

                this.lastQuery = q;

                this.resetData();
                if (!this.isValidData(q)) {
                    this.warningMsg = "Input was not valid!";
                    return;
                }

                this.searchBlock(q).catch(() => {
                    // if block query fails, search for transaction
                    this.searchTx(q).catch(() => {
                        // if transaction query fails, search for address
                        this.searchAddress(q);
                    })
                });
            },
            isValidData: function (str) {
                // TODO: check if str is address or transaction, also calculate checksum
                return str.length >= 34;
            },
            resetData: function () {
                this.query = "";
                this.$store.dispatch('resetMsg');
                this.transaction = null;
                this.address = null;
                this.block = null;
            },
            searchBlock: function (q) {
                console.log("Block search: " + q);
                return fetch("/blk/" + q)
                    .then(response => {
                        if (!response.ok) throw new Error(response.status + " " + response.statusText)
                        return response
                    })
                    .then(response => response.json())
                    .then(data => {
                        this.block = data;
                    })
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
        },
        created: function () {
            newRouting(this, this.$route.params.id);
        },
        watch: {
            '$route'(to) {
                newRouting(this, to.params.id);
            }
        }
    }
</script>

<style scoped>

</style>