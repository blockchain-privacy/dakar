<template>
    <v-app>
        <v-app-bar app color="primary" dark>
            <v-container fluid>
                <v-layout row>
                    <v-flex xs3>
                        <v-img
                                alt="Explorer Logo"
                                class="shrink mr-2"
                                contain
                                src="https://cdn.vuetifyjs.com/images/logos/vuetify-logo-dark.png"
                                transition="scale-transition"
                                width="40"/>
                        <h1 class="align-center mt-3">Dash Explorer</h1>
                    </v-flex>
                    <v-flex xs7>
                        <QueryInput/>
                    </v-flex>
                </v-layout>
            </v-container>
        </v-app-bar>
        <v-content>
            <v-container fluid>
                <MsgBox/>
                <transition name="fade">
                    <router-view/>
                </transition>
            </v-container>
            <v-footer class="pa-3" fixed>
                <v-spacer></v-spacer>
                <div>
                    &copy; {{ new Date().getFullYear() }}
                    <b>Dakar project</b> - <a href="http://ntnu.no">NTNU</a>
                </div>
            </v-footer>
        </v-content>
    </v-app>
</template>

<script>
    import QueryInput from "./components/QueryInput";
    import MsgBox from "./components/MsgBox";

    export default {
        name: "App",
        components: {
            MsgBox,
            QueryInput
        }
    };
</script>
