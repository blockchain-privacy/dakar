<template>
  <v-text-field style="margin: 23px 0 0 0"
                full-width outlined dense
                label="Search for blocks, transactions and addresses"
                :append-icon="icon.mdiMagnify"
                v-model="query"
                :rules="[isValidQuery]"
                @click:append="handleInput(query, 'user')"
                @keydown.enter="handleInput(query, 'user')"/>
</template>

<script>
import {
  mdiMagnify,
} from '@mdi/js';
import * as Constants from '../constants';
import {
  doGet, handleError, isValidQuery, isValidQueryInput, resetData,
} from '../utilities';

function newRouting(context) {
  const { id, pushFromUserInput } = context.$route.params;

  if (pushFromUserInput !== undefined || id === undefined
      || !(context.$route.name === Constants.ROUTE_NAME_BLOCK_PAGE
          || context.$route.name === Constants.ROUTE_NAME_ADDRESS_PAGE
          || context.$route.name === Constants.ROUTE_NAME_TRANSACTION_PAGE)) {
    return;
  }

  switch (context.$route.name) {
    case Constants.ROUTE_NAME_TRANSACTION_PAGE:
      context.handleQuery(id, Constants.RESPONSE_TYPE_TRANSACTION);
      break;
    case Constants.ROUTE_NAME_BLOCK_PAGE:
      context.handleQuery(id, Constants.RESPONSE_TYPE_BLOCK);
      break;
    case Constants.ROUTE_NAME_ADDRESS_PAGE:
      context.handleQuery(id, Constants.RESPONSE_TYPE_ADDRESS);
      break;
    default:
      context.handleQuery(id);
  }
}

export default {
  name: 'QueryInput',
  data() {
    return {
      // query is not managed by the vuex store
      // as it only needs to be accessed by this component
      query: '',
      lastQuery: '',
      icon: {
        mdiMagnify,
      },
    };
  },
  computed: {
    searchResultType: {
      get() {
        return this.$store.getters.getSearchResultType;
      },
    },
  },
  methods: {
    isValidQuery,
    async handleInput(q, origin) {
      // template string in case it is a number
      const query = `${q}`.trim();
      // update route only when input is from user and query is different
      if (origin === 'user' && query !== this.lastQuery) {
        if (!isValidQueryInput(query)) {
          this.setWarningMessage('Input was not valid');
          return;
        }

        // get data for route
        if (!await this.handleQuery(query)) {
          return;
        }

        // route to corresponding page
        switch (this.searchResultType) {
          case Constants.RESPONSE_EMPTY:
            await this.$router.push({ name: Constants.ROUTE_NAME_NO_RESULTS });
            break;
          case Constants.RESPONSE_TYPE_ADDRESS:
            await this.$router.push({
              name: Constants.ROUTE_NAME_ADDRESS_PAGE,
              params: { id: query, pushFromUserInput: true },
            });
            break;
          case Constants.RESPONSE_TYPE_BLOCK:
            await this.$router.push({
              name: Constants.ROUTE_NAME_BLOCK_PAGE,
              params: { id: query, pushFromUserInput: true },
            });
            break;
          case Constants.RESPONSE_TYPE_TRANSACTION:
            await this.$router.push({
              name: Constants.ROUTE_NAME_TRANSACTION_PAGE,
              params: { id: query, pushFromUserInput: true },
            });
            break;
          default:
            await this.$router.push({ name: Constants.ROUTE_NAME_NO_RESULTS });
            break;
        }
      } else if (origin === 'route') {
        // do nothing -> route is already up to date
      }
    },
    execQuery(route, action, parameter) {
      return doGet(route, this.$router, this.$store, parameter).then((data) => {
        this.$store.dispatch(action, data);
        this.$store.dispatch('resetMessages');
      }).catch((e) => {
        handleError(this.$store, e);
        return e;
      });
    },
    async handleQuery(q, type) {
      this.query = '';
      // template string in case it is a number
      const query = `${q}`.trim();

      if (this.lastQuery !== '' && this.lastQuery === query) return false;

      this.lastQuery = query;

      resetData(this);

      if (!isValidQueryInput(query)) {
        this.setWarningMessage('Input was not valid');
        return false;
      }

      switch (type) {
        case Constants.RESPONSE_TYPE_TRANSACTION:
          await this.execQuery(Constants.ROUTE_TRANSACTION, 'updateTransactionData', query);
          break;
        case Constants.RESPONSE_TYPE_BLOCK:
          await this.execQuery(Constants.ROUTE_BLOCK, 'updateBlockData', query);
          break;
        case Constants.RESPONSE_TYPE_ADDRESS:
          await this.execQuery(Constants.ROUTE_ADDRESS, 'updateAddressData', query);
          break;
        default:
          await this.execQuery(Constants.ROUTE_SEARCH, 'updateSearchResult', query);
      }

      return true;
    },
    setWarningMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'warning', temporary: true });
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
