<template>
    <v-layout row v-if="data">
        <v-flex xs12 sm8 offset-sm2>
            <v-card>
                <v-card-title>
                    <v-icon>mdi-database</v-icon>
                    Database Information
                </v-card-title>
                <v-list-item>
                    <v-list-item-avatar>
                        <v-icon>mdi-progress-wrench</v-icon>
                    </v-list-item-avatar>
                    <v-list-item-content>
                        <v-list-item-title>Database status</v-list-item-title>
                        <v-list-item-subtitle>
                            {{data.status}}
                        </v-list-item-subtitle>
                    </v-list-item-content>
                </v-list-item>
                <v-list two-line subheader>
                    <v-list-item>
                        <v-list-item-avatar>
                            <v-icon>mdi-format-header-pound</v-icon>
                        </v-list-item-avatar>
                        <v-list-item-content>
                            <v-list-item-title>Last block hash</v-list-item-title>
                            <v-list-item-subtitle>
                                <router-link :to="'search/' + data.lastblockhash">{{data.lastblockhash}}</router-link>
                            </v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                    <v-list-item>
                        <v-list-item-avatar>
                            <v-icon>mdi-pound</v-icon>
                        </v-list-item-avatar>
                        <v-list-item-content>
                            <v-list-item-title>Last block ID</v-list-item-title>
                            <v-list-item-subtitle>
                                {{ data.lastblockid}}
                            </v-list-item-subtitle>
                        </v-list-item-content>
                        <v-list-item-content>
                            <v-list-item-title>Stop block ID</v-list-item-title>
                            <v-list-item-subtitle>
                                {{ data.stopblockid}}
                            </v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                    <v-list-item>
                        <v-list-item-avatar>
                            <v-icon>mdi-map-marker-distance</v-icon>
                        </v-list-item-avatar>
                        <v-list-item-content>
                            <v-list-item-title>Range up</v-list-item-title>
                            <v-list-item-subtitle>
                                {{data.rangeup}}
                            </v-list-item-subtitle>
                        </v-list-item-content>
                        <v-list-item-content>
                            <v-list-item-title>Range down</v-list-item-title>
                            <v-list-item-subtitle>
                                {{data.rangedown}}
                            </v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                    <v-list-item>
                        <v-list-item-avatar>
                            <v-icon>mdi-pound</v-icon>
                        </v-list-item-avatar>
                        <v-list-item-content>
                            <v-list-item-title>Global block count</v-list-item-title>
                            <v-list-item-subtitle>
                                {{data.globalblockcount}}
                            </v-list-item-subtitle>
                        </v-list-item-content>
                        <v-list-item-content>
                            <v-list-item-title>Global transaction count</v-list-item-title>
                            <v-list-item-subtitle>
                                {{data.globaltxcount}}
                            </v-list-item-subtitle>
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
        name: 'EntryView',
        computed: {
            data() {
                return this.$store.getters.getMetaData;
            },
        },
        created() {
            this.$store.dispatch('updateMetaData');
        }
    }
</script>

<style scoped>

</style>