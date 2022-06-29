<template>
  <v-container fluid class="fill-height">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="12" md="10" lg="9" xl="8">
        <div class="d-flex justify-center mb-2">
          <v-img src="../assets/dakar_dash.svg" max-width="105px" style="z-index: 5"/>
        </div>
        <div class="d-flex justify-center ">
          <p class="text-h2" style="position:relative; z-index: 5">{{ appName }}</p>
        </div>
        <v-text-field
            v-model="query"
            :append-icon="icons.mdiMagnify"
            full-width
            outlined
            class="search-field v-input--is-focused"
            label="Search for blocks, transactions and addresses"
            :rules="[isValidQuery]"
            :background-color="$vuetify.theme.dark?'black':'white'"
            @click:append="handleQuery(query)"
            @keydown.enter="handleQuery(query)">
        </v-text-field>
        <div class="d-flex justify-center ">
          <p class="text-h6" style="position:relative; z-index: 5">
            {{ appSubtitle }}
          </p>
        </div>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import {
  mdiMagnify, mdiAccount,
} from '@mdi/js';
import * as d3 from 'd3';
import {
  ROUTE_NAME_LOGIN_PAGE, RESPONSE_EMPTY, ROUTE_NAME_NO_RESULTS,
  RESPONSE_TYPE_ADDRESS, ROUTE_NAME_ADDRESS_PAGE, RESPONSE_TYPE_BLOCK, ROUTE_NAME_BLOCK_PAGE,
  RESPONSE_TYPE_TRANSACTION, ROUTE_NAME_TRANSACTION_PAGE, APPLICATION_NAME, ROUTE_SEARCH,
  APPLICATION_SUBTITLE,
} from '../constants';
import {
  doGet, handleError, isValidQuery, isValidQueryInput,
} from '../utilities';

export default {
  name: 'EntryView',
  data() {
    return {
      query: '',
      route: {
        loginPage: ROUTE_NAME_LOGIN_PAGE,
      },
      icons: {
        mdiMagnify, mdiAccount,
      },
      isMenuVisible: false,
      appName: APPLICATION_NAME,
      appSubtitle: APPLICATION_SUBTITLE,
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
    handleThemeChange(isDark) {
      d3.selectAll('.bg-svg')
        .selectAll('path')
        .attr('stroke', () => (isDark ? 'white' : 'black'));
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
    async executeQuery(query) {
      await this.execQuery(ROUTE_SEARCH, 'updateSearchResult', query);
      return true;
    },
    async handleQuery(q) {
      // template string in case it is a number
      const query = `${q}`.trim();

      // ignore whitespace and empty queries
      if (query.length === 0) return;

      if (!isValidQueryInput(query)) {
        this.setWarningMessage('Query is invalid');
        return;
      }

      if (!await this.executeQuery(query)) {
        return;
      }

      switch (this.searchResultType) {
        case RESPONSE_EMPTY:
          await this.$router.push({ name: ROUTE_NAME_NO_RESULTS });
          break;
        case RESPONSE_TYPE_ADDRESS:
          await this.$router.push({
            name: ROUTE_NAME_ADDRESS_PAGE,
            params: { id: query, pushFromUserInput: true },
          });
          break;
        case RESPONSE_TYPE_BLOCK:
          await this.$router.push({
            name: ROUTE_NAME_BLOCK_PAGE,
            params: { id: query, pushFromUserInput: true },
          });
          break;
        case RESPONSE_TYPE_TRANSACTION:
          await this.$router.push({
            name: ROUTE_NAME_TRANSACTION_PAGE,
            params: { id: query, pushFromUserInput: true },
          });
          break;
        default:
          await this.$router.push({ name: ROUTE_NAME_NO_RESULTS });
          break;
      }
    },
    setWarningMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'warning', temporary: true });
    },
  },
  mounted() {
    document.title = this.appName;
    this.handleThemeChange(this.$vuetify.theme.dark);
    // add attributes to root svg
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
      this.handleThemeChange(e.matches);
    });
  },
  watch: {
    // eslint-disable-next-line func-names
    '$vuetify.theme.dark': function (isDark) {
      this.handleThemeChange(isDark);
    },
  },
};
</script>

<style scoped>

>>> .search-field fieldset {
  border-width: 3px 3px 3px 3px;
  border-color: #1976d2;
}

>>> .v-input--is-focused {
  transform: none;
}

</style>
