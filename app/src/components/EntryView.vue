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
            full-width
            outlined
            class="search-field v-input--is-focused"
            label="Search for blocks, transactions and addresses"
            :rules="[isValidQuery]"
            :background-color="this.$vuetify.theme.dark?'black':'white'"
            @keydown.enter="handleQuery(query)">
          <template v-slot:append-outer>
            <v-btn
                outlined
                class="search-btn"
                color="primary"
                @click="handleQuery(query)">
              <v-icon> {{ icons.mdiMagnify }}</v-icon>
            </v-btn>
          </template>
        </v-text-field>
        <div class="d-flex justify-center ">
          <p class="text-h6" style="position:relative; z-index: 5">
            Blockchain transaction analytics
          </p>
        </div>
      </v-col>
    </v-row>
    <div class="hidden-sm-and-down">
      <svg class="bg-svg" v-for="i in 10" :key="i +'a'"
           width="25.946mm" height="25.946mm" version="1.1" viewBox="0 0 25.946 25.946"
           xmlns="http://www.w3.org/2000/svg">
        <g transform="translate(-79.261 -114.16)">
          <g transform="rotate(45 134.89 -96.187)">
            <path d="m252.01 91.883h21.525" fill="none" stroke="#000" stroke-width=".25829px"/>
            <circle cx="248.96" cy="91.883" r="3.2982" fill="#008ee5" stroke-width=".074983"/>
            <circle cx="276.32" cy="91.883" r="3.2982" fill="#008ee5" stroke-width=".074983"/>
          </g>
        </g>
      </svg>
      <svg class="bg-svg" v-for="i in 10" :key="i +'b'"
           width="30.295mm" height="28.285mm" version="1.1" viewBox="0 0 30.295 28.285"
           xmlns="http://www.w3.org/2000/svg">
        <g transform="translate(-54.613 -103.28)">
          <g transform="translate(-192.88 29.538)">
            <path d="m253.43 86.567 18.641 10.762" fill="none" stroke="#000" stroke-width=".25823"/>
            <path d="m251.5 85.577 15.622-9.0193" fill="none" stroke="#000" stroke-width=".25823"/>
            <g fill="#008ee5" stroke-width=".074983">
              <circle transform="rotate(30)" cx="259.71" cy="-51.746" r="3.2982"/>
              <circle transform="rotate(30)" cx="287.07" cy="-51.746" r="3.2982"/>
              <circle transform="rotate(30)" cx="270.07" cy="-66.973" r="3.2982"/>
            </g>
          </g>
        </g>
      </svg>
      <svg class="bg-svg" v-for="i in 10" :key="i +'c'"
           width="58.463mm" height="47.275mm" version="1.1" viewBox="0 0 58.463 47.275"
           xmlns="http://www.w3.org/2000/svg">
        <g transform="translate(-45.709 -108.98)">
          <g transform="translate(-147.34 34.424)">
            <path d="m220.92 95.044 15.22 15.22" fill="none" stroke="#000" stroke-width=".25829px"/>
            <circle transform="rotate(45)" cx="247.74" cy="-89.01" r="3.2982"
                    fill="#008ee5" stroke-width=".074983"/>
            <path d="m220.92 95.044 5.571 20.791" fill="none"
                  stroke="#000" stroke-width=".25829px"/>
            <circle transform="rotate(75)" cx="173.3" cy="-188.8" r="3.2982"
                    fill="#008ee5" stroke-width=".074983"/>
            <path d="m220.92 95.044 15.22-15.22" fill="none" stroke="#000" stroke-width=".25829px"/>
            <circle transform="rotate(-45)" cx="113.32" cy="223.42" r="3.2982"
                    fill="#008ee5" stroke-width=".074983"/>
            <path d="m199.4 94.751h21.525" fill="none" stroke="#000" stroke-width=".25829px"/>
            <circle cx="196.35" cy="94.751" r="3.2982" fill="#008ee5" stroke-width=".074983"/>
            <path d="m223.9 94.751h21.525" fill="none" stroke="#000" stroke-width=".25829px"/>
            <circle cx="220.85" cy="94.751" r="3.2982" fill="#008ee5" stroke-width=".074983"/>
            <circle cx="248.22" cy="94.751" r="3.2982" fill="#008ee5" stroke-width=".074983"/>
          </g>
        </g>
      </svg>
      <svg class="bg-svg" v-for="i in 10" :key="i +'d'"
           width="33.961mm" height="6.5963mm" version="1.1" viewBox="0 0 33.961 6.5963"
           xmlns="http://www.w3.org/2000/svg">
        <g transform="translate(-92.84 -68.534)">
          <g transform="translate(-152.82 -20.051)">
            <path d="m252.01 91.883h21.525" fill="none" stroke="#000" stroke-width=".25829px"/>
            <circle cx="248.96" cy="91.883" r="3.2982" fill="#008ee5" stroke-width=".074983"/>
            <circle cx="276.32" cy="91.883" r="3.2982" fill="#008ee5" stroke-width=".074983"/>
          </g>
        </g>
      </svg>
    </div>
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
  RESPONSE_TYPE_TRANSACTION, ROUTE_NAME_TRANSACTION_PAGE, APPLICATION_NAME,
} from '../constants';
import '../style.scss';
import { isValidQuery, isValidQueryInput } from '../utilities';

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
    async executeQuery(query) {
      await this.$store.dispatch('updateSearchResult', query);
      return true;
    },
    async handleQuery(q) {
      // template string in case it is a number
      const query = `${q}`.trim();

      if (!isValidQueryInput(query)) {
        this.setWarningMessage('Input was not valid');
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
  beforeMount() {

  },
};
</script>

<style scoped>

.search-field {
  z-index: 5;
  border-bottom-right-radius: 0;
  border-top-right-radius: 0;
}

>>> .search-field fieldset {
  border-width: 3px 0 3px 3px;
  border-color: #1976d2;
}

>>> .v-input--is-focused {
  transform: none;
}

.search-btn {
  /*background-color: white;*/
  margin-left: -10px;
  margin-top: -19px;
  height: 57px !important;
  border-bottom-left-radius: 0;
  border-top-left-radius: 0;
  border-width: 3px 3px 3px 0;
}

</style>
