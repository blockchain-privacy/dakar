<template>
  <v-container
    :fluid="true"
    class="content"
  >
    <v-row
      align="center"
      justify="center"
    >
      <v-col
        cols="12"
        sm="12"
        md="10"
        lg="9"
        xl="8"
      >
        <!-- set width and height so transition looks better -->
        <div
          class="d-flex justify-center mb-2 mx-auto"
          style="width: 105px; height: 187px"
        >
          <v-img
            src="../assets/dakar_dash.svg"
            max-width="105px"
            transition="fade-transition"
          />
        </div>
        <div class="d-flex justify-center">
          <p
            class="text-h2"
            style="position:relative"
          >
            {{ appName }}
          </p>
        </div>
        <v-text-field
          v-model="query"
          class="mt-3"
          :append-inner-icon="icons.mdiMagnify"
          variant="outlined"
          label="Search for blocks, transactions and addresses"
          :rules="[isValidQuery]"
          single-line
          @click:append-inner="handleQuery(query)"
          @keydown.enter="handleQuery(query)"
        />
        <div class="d-flex justify-center ">
          <p
            class="text-h6"
            style="position:relative"
          >
            {{ appSubtitle }}
          </p>
        </div>
      </v-col>
    </v-row>
    <!-- footer -->
    <v-container
      v-if="false"
      class="align-self-end"
    >
      <v-row justify="center">
        <v-col
          md="2"
          class="text-center mx-1"
        >
          <v-btn
            :to="{name: route.aboutPage}"
            variant="plain"
            size="small"
          >
            About
          </v-btn>
        </v-col>
        <v-col
          md="2"
          class="text-center mx-1"
        >
          <v-btn
            :to="{name: route.termsOfUsePage}"
            variant="plain"
            size="small"
          >
            Terms of Use
          </v-btn>
        </v-col>
        <v-col
          md="2"
          class="text-center mx-1"
        >
          <v-btn
            :to="{name: route.privacyPage}"
            variant="plain"
            size="small"
          >
            Privacy Policy
          </v-btn>
        </v-col>
      </v-row>
    </v-container>
  </v-container>
</template>

<script>
import {mdiMagnify, mdiAccount} from '@mdi/js';
import {
	ROUTE_NAME_LOGIN_PAGE, RESPONSE_EMPTY, ROUTE_NAME_NO_RESULTS, RESPONSE_TYPE_ADDRESS,
	ROUTE_NAME_ADDRESS_PAGE, RESPONSE_TYPE_BLOCK, ROUTE_NAME_BLOCK_PAGE, RESPONSE_TYPE_TRANSACTION,
	ROUTE_NAME_TRANSACTION_PAGE, APPLICATION_NAME, APPLICATION_SUBTITLE, ROUTE_NAME_ABOUT,
	ROUTE_NAME_TERMS_OF_USE, ROUTE_NAME_PRIVACY,
} from '@/constants';
import {handleError, isValidQuery, isValidQueryInput} from '@/utilities';

export default {
	name: 'EntryPage',
	data() {
		return {
			query: '',
			route: {
				loginPage: ROUTE_NAME_LOGIN_PAGE,
				aboutPage: ROUTE_NAME_ABOUT,
				termsOfUsePage: ROUTE_NAME_TERMS_OF_USE,
				privacyPage: ROUTE_NAME_PRIVACY,
			},
			icons: {mdiMagnify, mdiAccount},
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
		isPushFromUserInput: {
			async set(value) {
				await this.$store.dispatch('setPushFromUserInput', value);
			},
			get() {
				return this.$store.getters.getPushFromUserInput;
			},
		},
	},
	mounted() {
		document.title = this.appName;
	},
	methods: {
		isValidQuery,
		async executeQuery(query) {
			let ok = false;

			try {
				const response = await this.dakar.data.searchQueryGet({query});

				this.$store.dispatch('updateSearchResult', response);
				ok = response?.type !== RESPONSE_EMPTY;
				if (!ok) {
					this.setWarningMessage('server error');
				}
			} catch (e) {
				handleError(this, e);
			}

			return ok;
		},
		async handleQuery(q) {
			// Template string in case it is a number
			const query = `${q}`.trim();

			// ignore whitespace and empty queries
			if (query.length === 0 || !isValidQueryInput(query) || !await this.executeQuery(query)) {
				return;
			}

			switch (this.searchResultType) {
				case RESPONSE_EMPTY:
					await this.$router.push({name: ROUTE_NAME_NO_RESULTS});
					break;
				case RESPONSE_TYPE_ADDRESS:
					this.isPushFromUserInput = true;
					await this.$router.push({name: ROUTE_NAME_ADDRESS_PAGE, params: {id: query}});
					break;
				case RESPONSE_TYPE_BLOCK:
					this.isPushFromUserInput = true;
					await this.$router.push({name: ROUTE_NAME_BLOCK_PAGE, params: {id: query}});
					break;
				case RESPONSE_TYPE_TRANSACTION:
					this.isPushFromUserInput = true;
					await this.$router.push({name: ROUTE_NAME_TRANSACTION_PAGE, params: {id: query}});
					break;
				default:
					await this.$router.push({name: ROUTE_NAME_NO_RESULTS});
					break;
			}
		},
		setWarningMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'warning', temporary: true, category: this.$route.name});
		},
	},
};
</script>

<style scoped>

.content {
  display: flex;
  flex-direction: column;
  height: 100%
}

:deep(.v-field__outline) {
  border-width: 3px 3px 3px 3px;
  color: #1976d2 !important;
  opacity: 1;
}

:deep(.v-field__outline__start) {
  border-width: 3px 0 3px 3px;
  color: #1976d2 !important;
  opacity: 1;
}
:deep(.v-field__outline__end) {
  border-width: 3px 3px 3px 0;
  color: #1976d2 !important;
  opacity: 1;
}

</style>
