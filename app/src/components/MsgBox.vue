<template>
    <v-layout row>
        <v-flex>
            <v-alert xs6 :value="infoMsg && infoMsg !== ''" type="info">
                {{ infoMsg }}
            </v-alert>
            <v-alert xs6 :value="errorMsg && errorMsg !== ''" type="error">
                {{ errorMsg }}
            </v-alert>
            <v-alert xs6 :value="successMsg && successMsg !== ''" type="success">
                {{ successMsg }}
            </v-alert>
            <v-alert xs6 :value="warningMsg && warningMsg !== ''" type="warning">
                {{ warningMsg }}
            </v-alert>
        </v-flex>
    </v-layout>
</template>

<script>
    export default {
        name: 'MsgBox',
        computed: {
            errorMsg() {
                return this.$store.getters.getErrorMsg;
            },
            infoMsg() {
                return this.$store.getters.getInfoMsg;
            },
            successMsg() {
                return this.$store.getters.getSuccessMsg;
            },
            warningMsg() {
                return this.$store.getters.getWarningMsg;
            }
        }
    }
</script>

<style scoped>

</style>