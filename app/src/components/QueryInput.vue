<template>
  <v-text-field style="margin: 23px 0 0 0"
                full-width outlined dense
                label="Search for blocks, transactions and addresses"
                :append-icon="icon.mdiMagnify"
                v-model="query"
                :rules="[isQueryValid]"
                @click:append="handleInput(query, 'user')"
                @keydown.enter="handleInput(query, 'user')"/>
</template>

<script>
import {
  mdiMagnify,
} from '@mdi/js';
import * as Constants from '../constants';
import * as Utility from '../utilities';

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
    searchResultType: {
      get() {
        return this.$store.getters.getSearchResultType;
      },
    },
  },
  methods: {
    async handleInput(q, origin) {
      // template string in case it is a number
      const query = `${q}`.trim();
      // update route only when input is from user and query is different
      if (origin === 'user' && query !== this.lastQuery) {
        if (!this.isValidData(query)) {
          this.warningMsg = 'Input was not valid!';
          return;
        }

        if (!await this.handleQuery(query)) {
          return;
        }

        switch (this.searchResultType) {
          case Constants.RESPONSE_EMPTY:
            this.$router.push({ name: Constants.ROUTE_NAME_NO_RESULTS });
            break;
          case Constants.RESPONSE_TYPE_ADDRESS:
            this.$router.push({
              name: Constants.ROUTE_NAME_ADDRESS_PAGE,
              params: { id: query, pushFromUserInput: true },
            });
            break;
          case Constants.RESPONSE_TYPE_BLOCK:
            this.$router.push({
              name: Constants.ROUTE_NAME_BLOCK_PAGE,
              params: { id: query, pushFromUserInput: true },
            });
            break;
          case Constants.RESPONSE_TYPE_TRANSACTION:
            this.$router.push({
              name: Constants.ROUTE_NAME_TRANSACTION_PAGE,
              params: { id: query, pushFromUserInput: true },
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
    async handleQuery(q, type) {
      // if (origin === 'user' && q !== this.lastQuery) {
      //   // update route only when input is from user and query is different
      //   this.$router.push({name: Constants.ROUTE_NAME_SEARCH_PAGE, params: {id: q}});
      // } else if (origin === 'route') {
      //   // do nothing -> route is already up to date
      // }
      this.query = '';
      // template string in case it is a number
      const query = `${q}`.trim();

      if (this.lastQuery !== '' && this.lastQuery === query) return false;

      this.lastQuery = query;

      Utility.resetData(this);

      if (!this.isValidData(query)) {
        this.warningMsg = 'Input was not valid!';
        return false;
      }

      switch (type) {
        case Constants.RESPONSE_TYPE_TRANSACTION:
          await this.$store.dispatch('updateTransactionData', query);
          break;
        case Constants.RESPONSE_TYPE_BLOCK:
          await this.$store.dispatch('updateBlockData', query);
          break;
        case Constants.RESPONSE_TYPE_ADDRESS:
          await this.$store.dispatch('updateAddressData', query);
          break;
        default:
          await this.$store.dispatch('updateSearchResult', query);
      }

      return true;
    },
    isValidData(str) {
      const inputLen = str.length;
      // 64 -> length of transaction hash and block hash
      if (inputLen === 0 || inputLen > 64) return false;

      // 34 -> address length; if smaller than it must be a block id
      if (inputLen < 34) {
        return Number.isInteger(Number(str));
      }

      return str.match(/^[0-9a-zA-Z]+$/);
    },
    isQueryValid(str) {
      const trimmed = str.trim();

      return trimmed.length === 0 ? true : this.isValidData(trimmed);
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
