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
            <v-form v-on:submit.prevent="handleQuery(query)" class="d-flex mx-auto">
              <v-text-field class="d-flex" full-width v-model="query" label="Search for transactions and addresses"/>
            </v-form>
          </v-flex>
        </v-layout>
      </v-container>
    </v-app-bar>
    <v-content>
      <v-container fluid>
        <v-layout row>
          <v-flex>
            <v-alert xs6 :value="errorMsg && errorMsg !== ''" type="error">
              {{ errorMsg }}
            </v-alert>
          </v-flex>
        </v-layout>
        <v-layout row>
          <TxLookup v-if="transaction" :data="transaction"/>
          <AddressLookup v-if="address" :data="address"/>
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
  import AddressLookup from "./components/AddressLookup";
  export default {
    name: "App",

    components: {
      TxLookup,
      AddressLookup,
    },

    data: function () {
      return {
        query: "",
        errorMsg: null,
        transaction: null,
        address: null
      };
    },

    methods: {
      handleQuery: function(q){
          this.query = ""
          this.errorMsg = null
          this.transaction = null
          this.address = null
         this.searchTx(q).catch(() => {
             // if transaction query fails, search for address
             this.searchAddress(q)
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
                  this.errorMsg = error
                });
            }
    }
  };
</script>
