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
            <v-form v-on:submit.prevent="searchTx(query)" class="d-flex mx-auto">
              <v-text-field class="d-flex" full-width v-model="query" label="Blockchain search"/>
            </v-form>
          </v-flex>
        </v-layout>
      </v-container>
    </v-app-bar>
    <v-content>
      <v-container fluid>
        <v-layout row>
          <v-flex>
            <v-alert xs6 :value="errorMsg && errorMsg != ''" type="error">
              {{ errorMsg }}
            </v-alert>
          </v-flex>
        </v-layout>
        <v-layout row>
          <TxLookup v-if="transaction" :data="transaction"/>
        </v-layout>
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
  import TxLookup from "./components/TxLookup";

  export default {
    name: "App",

    components: {
      TxLookup
    },

    data: function () {
      return {
        query: "",
        errorMsg: null,
        transaction: null
      };
    },

    methods: {
      searchTx: function (q) {
        this.query = ""
        this.errorMsg = null
        console.log("Searching clicked: " + q);
        fetch("/tx/" + q)
          .then(response => {
            if (!response.ok) throw new Error(response.status + " " + response.statusText)
            return response
          })
          .then(response => response.json())
          .then(data => {
            this.transaction = data;
          })
          .catch(error => {
            this.errorMsg = error
          });
      }
    }
  };
</script>
