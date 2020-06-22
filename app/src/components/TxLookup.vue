<template>
    <v-layout row v-if="data">
        <v-flex xs12 sm8 offset-sm2>
            <v-card>
                <v-card-title>
                    <v-icon>mdi-bank-transfer</v-icon>
                    Transaction
                </v-card-title>
                <v-list two-line subheader>
                    <v-list-item>
                        <v-list-item-avatar>
                            <v-icon>mdi-format-header-pound</v-icon>
                        </v-list-item-avatar>
                        <v-list-item-content>
                            <v-list-item-title>HASH</v-list-item-title>
                            <v-list-item-subtitle>
                                <router-link :to="data.tx.hash">{{data.tx.hash}}</router-link>
                            </v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                    <v-list-item>
                        <v-list-item-avatar>
                            <v-icon>mdi-calendar</v-icon>
                        </v-list-item-avatar>
                        <v-list-item-content>
                            <v-list-item-title>DATE</v-list-item-title>
                            <v-list-item-subtitle>ts: {{data.tx.bts}} - {{new Date(data.tx.bts)}}</v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                    <v-list-item>
                        <v-list-item-avatar>
                            <v-icon></v-icon>
                        </v-list-item-avatar>
                        <v-list-item-content>
                            <v-list-item-title>Confirmations</v-list-item-title>
                            <v-list-item-subtitle>{{data.confirmations}}</v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                    <v-list-item>
                        <v-list-item-avatar>
                            <v-icon></v-icon>
                        </v-list-item-avatar>
                        <v-list-item-content>
                            <v-list-item-title>Block</v-list-item-title>
                            <v-list-item-subtitle>
                                {{data.bheight}} -
                                <router-link :to="data.bhash">{{data.bhash}}</router-link>
                            </v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                    <v-list-item>
                        <v-list-item-avatar>
                            <v-icon></v-icon>
                        </v-list-item-avatar>
                        <v-list-item-content>
                            <v-list-item-title>Inputs</v-list-item-title>
                            <v-list-item-subtitle>{{data.tx.inputs.length}}</v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                    <v-list-item>
                        <v-list-item-avatar>
                            <v-icon></v-icon>
                        </v-list-item-avatar>
                        <v-list-item-content>
                            <v-list-item-title>Outputs</v-list-item-title>
                            <v-list-item-subtitle>{{data.tx.outputs.length}}</v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                </v-list>
            </v-card>
            <v-card><!-- TODO hack, to leave space at the bottom, such that content is not hidden by the footer.-->
                <div style="height: 40pt">
                </div>
            </v-card>
        </v-flex>
    </v-layout>
</template>

<script>
    export default {
        name: 'TxLookup',
        computed: {
            data() {
                return this.$store.getters.getTransactionData;
            }
        },
    }
</script>
