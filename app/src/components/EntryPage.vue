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
            src="../assets/dakar_animated.svg"
            max-width="105px"
            transition="fade-transition"
          />
        </div>
        <div class="d-flex justify-center">
          <p
            class="text-h2"
            style="position:relative"
          >
            {{ APPLICATION_NAME }}
          </p>
        </div>
        <v-text-field
          id="entry-page-search"
          v-model="query"
          class="mt-3"
          :append-inner-icon="mdiMagnify"
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
            {{ APPLICATION_SUBTITLE }}
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
            :to="{name: ROUTE_NAME_ABOUT}"
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
            :to="{name: ROUTE_NAME_TERMS_OF_USE}"
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
            :to="{name: ROUTE_NAME_PRIVACY}"
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

<script setup>
import {mdiMagnify} from '@mdi/js';
import {
	RESPONSE_EMPTY, ROUTE_NAME_NO_RESULTS, RESPONSE_TYPE_ADDRESS,
	ROUTE_NAME_ADDRESS_PAGE, RESPONSE_TYPE_BLOCK, ROUTE_NAME_BLOCK_PAGE, RESPONSE_TYPE_TRANSACTION,
	ROUTE_NAME_TRANSACTION_PAGE, APPLICATION_NAME, APPLICATION_SUBTITLE, ROUTE_NAME_ABOUT,
	ROUTE_NAME_TERMS_OF_USE, ROUTE_NAME_PRIVACY,
} from '@/constants';
import {handleError, isValidQuery, isValidQueryInput} from '@/utilities';
import {computed, inject, onMounted, ref} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import {useExplorerStore} from '@/pinia/explorer';
import {storeToRefs} from 'pinia';
import {useNavStore} from '@/pinia/nav';

const dakar = inject('dakar');
const router = useRouter();
const route = useRoute();
const msgStore = useMsgStore();
const explorerStore = useExplorerStore();
const {pushFromUserInput} = storeToRefs(useNavStore());
const context = {$route: route, addMessage: msgStore.addMessage};

const query = ref('');

// Computed
const searchResultType = computed(() => explorerStore.getSearchResultType);

// Hooks
onMounted(() => {
	document.title = APPLICATION_NAME;
});

// Functions
function setWarningMessage(msg) {
	msgStore.addMessage({text: msg, type: 'warning', temporary: true, category: route.name});
}

async function executeQuery(query) {
	let ok = false;

	try {
		const response = await dakar.data.searchQueryGet({query});

		explorerStore.updateSearchResult(response);
		ok = response?.type !== RESPONSE_EMPTY;
		if (!ok) {
			setWarningMessage('server error');
		}
	} catch (e) {
		handleError(context, e);
	}

	return ok;
}

async function handleQuery(q) {
	// Template string in case it is a number
	const query = `${q}`.trim();

	// ignore whitespace and empty queries
	if (query.length === 0 || !isValidQueryInput(query) || !await executeQuery(query)) {
		return;
	}

	switch (searchResultType.value) {
		case RESPONSE_EMPTY:
			await router.push({name: ROUTE_NAME_NO_RESULTS});
			break;
		case RESPONSE_TYPE_ADDRESS:
			pushFromUserInput.value = true;
			await router.push({name: ROUTE_NAME_ADDRESS_PAGE, params: {id: query}});
			break;
		case RESPONSE_TYPE_BLOCK:
			pushFromUserInput.value = true;
			await router.push({name: ROUTE_NAME_BLOCK_PAGE, params: {id: query}});
			break;
		case RESPONSE_TYPE_TRANSACTION:
			pushFromUserInput.value = true;
			await router.push({name: ROUTE_NAME_TRANSACTION_PAGE, params: {id: query}});
			break;
		default:
			await router.push({name: ROUTE_NAME_NO_RESULTS});
			break;
	}
}

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
