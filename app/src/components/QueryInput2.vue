<template>
  <v-text-field @keydown.enter="handleInput(query, 'user')"
                class="d-flex" full-width v-model="query"
                label="Search for blocks, transactions and addresses"/>
</template>

<script>
import * as Constants from '../constants';
import * as Utility from '../utilities';

function newRouting(context) {
  const { id } = context.$route.params;

  if (id === undefined || !(context.$route.name === Constants.ROUTE_NAME_BLOCK_PAGE
      || context.$route.name === Constants.ROUTE_NAME_ADDRESS_PAGE
      || context.$route.name === Constants.ROUTE_NAME_TRANSACTION_PAGE)) {
    return;
  }
  console.log('new routing', context.$route.name);
  context.handleQuery(id);
}

export default {
  name: 'QueryInput2',
  data() {
    return {
      // query is not managed by the vuex store
      // as it only needs to be accessed by this component
      query: '',
      lastQuery: '',
      lastRoute: '',
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
    searchResult: {
      get() {
        return this.$store.getters.getSearchResult;
      },
    },
  },
  methods: {
    async handleInput(q, origin) {
      const query = q.trim();
      // update route only when input is from user and query is different
      if (origin === 'user' && query !== this.lastQuery) {
        if (!this.isValidData(query)) {
          this.warningMsg = 'Input was not valid!';
          return;
        }

        if (!await this.handleQuery(query)) {
          return;
        }

        switch (this.searchResult.type) {
          case Constants.RESPONSE_EMPTY:
            this.$router.push({ name: Constants.ROUTE_NAME_NO_RESULTS });
            break;
          case Constants.RESPONSE_TYPE_ADDRESS:
            this.$router.push({
              name: Constants.ROUTE_NAME_ADDRESS_PAGE,
              params: { id: query },
            });
            break;
          case Constants.RESPONSE_TYPE_BLOCK:
            this.$router.push({
              name: Constants.ROUTE_NAME_BLOCK_PAGE,
              params: { id: query },
            });
            break;
          case Constants.RESPONSE_TYPE_TRANSACTION:
            this.$router.push({
              name: Constants.ROUTE_NAME_TRANSACTION_PAGE,
              params: { id: query },
            });
            break;
          default:
            this.$router.push({ name: Constants.ROUTE_NAME_NO_RESULTS });
            break;
        }

        // this.$router.push({ name: Constants.ROUTE_NAME_SEARCH_PAGE, params: { id: query } });
      } else if (origin === 'route') {
        // do nothing -> route is already up to date
      }
    },
    async handleQuery(q) {
      // if (origin === 'user' && q !== this.lastQuery) {
      //   // update route only when input is from user and query is different
      //   this.$router.push({name: Constants.ROUTE_NAME_SEARCH_PAGE, params: {id: q}});
      // } else if (origin === 'route') {
      //   // do nothing -> route is already up to date
      // }
      this.query = '';
      const query = q.trim();
      if (this.lastQuery !== '' && this.lastQuery === query) return false;

      this.lastQuery = query;

      Utility.resetData(this);

      if (!this.isValidData(query)) {
        this.warningMsg = 'Input was not valid!';
        return false;
      }

      await this.$store.dispatch('updateSearchResult', query);
      return true;
    },
    isValidData(str) {
      return str.length > 0 && str.match(/^[0-9a-zA-Z]+$/);
    },
  },
  created() {
    newRouting(this);
  },
  watch: {
    $route() {
      if (this.lastQuery !== this.$route.name) this.lastQuery = '';
      newRouting(this);
    },
  },
};
</script>

<style scoped>

</style>
