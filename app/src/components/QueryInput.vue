<template>
  <v-text-field @keydown.enter="handleInput(query, 'user')"
                class="d-flex" full-width v-model="query"
                label="Search for blocks, transactions and addresses"/>
</template>

<script>
import * as Constants from '../constants';
import * as Utility from '../utilities';

function newRouting(context) {
1  const { id } = context.$route.params;
  if (id === undefined || context.$route.name !== Constants.ROUTE_NAME_SEARCH_PAGE) {
    return;
  }

  context.handleQuery(id);
}

export default {
  name: 'QueryInput',
  data() {
    return {
      // query is not managed by the vuex store
      // as it only needs to be accessed by this component
      query: '',
      lastQuery: '',
    };
  },
  computed: {
    errorMsg: {
      get() {
        return this.$store.getters.getErrorMsg;
      },
      set(value) {
        this.$store.dispatch('setErrorMsg', value);
      },
    },
    warningMsg: {
      get() {
        return this.$store.getters.getWarningMsg;
      },
      set(value) {
        this.$store.dispatch('setWarningMsg', value);
      },
    },
    transaction: {
      get() {
        return this.$store.getters.getTransactionData;
      },
      set(value) {
        this.$store.dispatch('setTransactionData', value);
      },
    },
    address: {
      get() {
        return this.$store.getters.getAddressData;
      },
      set(value) {
        this.$store.dispatch('setAddressData', value);
      },
    },
    block: {
      get() {
        return this.$store.getters.getBlockData;
      },
      set(value) {
        this.$store.dispatch('setBlockData', value);
      },
    },
  },
  methods: {
    handleInput(q, origin) {
      const query = q.trim();
      if (origin === 'user' && query !== this.lastQuery) {
        // update route only when input is from user and query is different

        if (!this.isValidData(query)) {
          this.warningMsg = 'Input was not valid!';
          return;
        }

        this.$router.push({ name: Constants.ROUTE_NAME_SEARCH_PAGE, params: { id: query } });
      } else if (origin === 'route') {
        // do nothing -> route is already up to date
      }
    },
    handleQuery(q) {
      // if (origin === 'user' && q !== this.lastQuery) {
      //   // update route only when input is from user and query is different
      //   this.$router.push({name: Constants.ROUTE_NAME_SEARCH_PAGE, params: {id: q}});
      // } else if (origin === 'route') {
      //   // do nothing -> route is already up to date
      // }
      const query = q.trim();
      this.lastQuery = query;

      this.query = '';
      Utility.resetData(this);

      if (!this.isValidData(query)) {
        this.warningMsg = 'Input was not valid!';
        return;
      }

      this.searchBlock(query).catch(() => {
        // if block query fails, search for transaction
        this.searchTx(query).catch(() => {
          // if transaction query fails, search for address
          this.searchAddress(query);
        });
      });
    },
    isValidData(str) {
      return str.length > 0 && str.match(/^[0-9a-zA-Z]+$/);
    },
    searchBlock(q) {
      console.log(`Block search: ${q}`);
      return fetch(Constants.ROUTE_BLOCK + q)
        .then((response) => {
          if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
          return response;
        })
        .then((response) => response.json())
        .then((data) => {
          this.block = data;
        });
    },
    searchTx(q) {
      console.log(`Tx search: ${q}`);
      return fetch(Constants.ROUTE_TRANSACTION + q)
        .then((response) => {
          if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
          return response;
        })
        .then((response) => response.json())
        .then((data) => {
          this.transaction = data;
        });
    },
    searchAddress(q) {
      console.log(`Address search: ${q}`);
      fetch(Constants.ROUTE_ADDRESS + q)
        .then((response) => {
          if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
          return response;
        })
        .then((response) => response.json())
        .then((data) => {
          this.address = data;
        })
        .catch((error) => {
          this.errorMsg = error;
        });
    },
  },
  created() {
    newRouting(this);
  },
  watch: {
    $route() {
      this.lastQuery = '';
      newRouting(this);
    },
  },
};
</script>

<style scoped>

</style>
