<template>
  <v-text-field
    id="query-input"
    v-model="query"
    :hide-details="true"
    style="min-width:220px"
    variant="outlined"
    density="compact"
    color="primary"
    :single-line="true"
    label="Search for blocks, transactions and addresses"
    :append-inner-icon="mdiMagnify"
    :rules="[isValidQuery]"
    @click:append-inner="handleInput(query)"
    @keydown.enter="handleInput(query)"
  />
</template>

<script setup>
import {mdiMagnify} from '@mdi/js';
import {
	RESPONSE_EMPTY, RESPONSE_TYPE_ADDRESS, RESPONSE_TYPE_BLOCK, RESPONSE_TYPE_TRANSACTION,
	ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_NO_RESULTS, ROUTE_NAME_TRANSACTION_PAGE,
} from '@/constants';
import {handleError, isValidQuery, isValidQueryInput} from '@/utilities';
import {
	computed, inject, ref, watch,
} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {useExplorerStore} from '@/pinia/explorer';
import {useMsgStore} from '@/pinia/msg';
import {useNavStore} from '@/pinia/nav';
import {storeToRefs} from 'pinia';

const dakar = inject('dakar');
const route = useRoute();
const router = useRouter();
const msgStore = useMsgStore();
const explorerStore = useExplorerStore();
const {pushFromUserInput} = storeToRefs(useNavStore());
const context = {addMessage: msgStore.addMessage, $route: useRoute()};

const query = ref('');
let lastQuery = '';

watch(route, () => {
	lastQuery = '';
	newRouting();
});

// Computed
const searchResultType = computed(() => explorerStore.getSearchResultType);

// Functions
function newRouting() {
	const {id} = route.params;
	const isPushFromUserInput = pushFromUserInput.value;

	if (isPushFromUserInput) {
		pushFromUserInput.value = false;
	}

	if (isPushFromUserInput || !id
      || !(route.name === ROUTE_NAME_BLOCK_PAGE
          || route.name === ROUTE_NAME_ADDRESS_PAGE
          || route.name === ROUTE_NAME_TRANSACTION_PAGE)) {
		return;
	}

	switch (route.name) {
		case ROUTE_NAME_TRANSACTION_PAGE:
			handleQuery(id, RESPONSE_TYPE_TRANSACTION);
			break;
		case ROUTE_NAME_BLOCK_PAGE:
			handleQuery(id, RESPONSE_TYPE_BLOCK);
			break;
		case ROUTE_NAME_ADDRESS_PAGE:
			handleQuery(id, RESPONSE_TYPE_ADDRESS);
			break;
		default:
			handleQuery(id);
	}
}

async function handleInput(q) {
	// Template string in case it is a number
	const trimmedQuery = `${q}`.trim();
	// Update route only when input is from user and query is different
	// ignore whitespace and empty queries
	// Get data for route
	if (trimmedQuery.length === 0 || trimmedQuery === lastQuery || !isValidQueryInput(trimmedQuery) || !await handleQuery(trimmedQuery)) {
		setWarningMessage('Input was not valid');
		return;
	}

	// Route to corresponding page
	switch (searchResultType.value) {
		case RESPONSE_EMPTY:
			await router.push({name: ROUTE_NAME_NO_RESULTS});
			break;
		case RESPONSE_TYPE_ADDRESS:
			pushFromUserInput.value = true;
			await router.push({name: ROUTE_NAME_ADDRESS_PAGE, params: {id: trimmedQuery}});
			break;
		case RESPONSE_TYPE_BLOCK:
			pushFromUserInput.value = true;
			await router.push({name: ROUTE_NAME_BLOCK_PAGE, params: {id: trimmedQuery}});
			break;
		case RESPONSE_TYPE_TRANSACTION:
			pushFromUserInput.value = true;
			await router.push({name: ROUTE_NAME_TRANSACTION_PAGE, params: {id: trimmedQuery}});
			break;
		default:
			await router.push({name: ROUTE_NAME_NO_RESULTS});
			break;
	}
}

async function storeResult(promise, action, piniaAction) {
	try {
		const response = await promise;
		piniaAction(response);
	} catch (e) {
		handleError(context, e);
	}
}

async function handleQuery(q, type) {
	query.value = '';
	// Template string in case it is a number
	const trimmedQuery = `${q}`.trim();

	if (lastQuery !== '' && lastQuery === trimmedQuery) {
		return false;
	}

	lastQuery = trimmedQuery;

	if (!isValidQueryInput(trimmedQuery)) {
		setWarningMessage('Input was not valid');
		return false;
	}

	// Reset messages here
	msgStore.resetMessages();
	explorerStore.setAddress(null);
	explorerStore.setBlock(null);
	explorerStore.setTransaction(null);

	switch (type) {
		case RESPONSE_TYPE_TRANSACTION:
			await storeResult(dakar.data.blockchainTransactionsHashGet({hash: trimmedQuery}), 'updateTransactionData', explorerStore.updateTransaction);
			break;
		case RESPONSE_TYPE_BLOCK:
			await storeResult(dakar.data.blockchainBlocksHashGet({hash: trimmedQuery}), 'updateBlockData', explorerStore.updateBlock);
			break;
		case RESPONSE_TYPE_ADDRESS:

			await storeResult(dakar.data.blockchainAddressesHashGet({hash: trimmedQuery}), 'updateAddressData', explorerStore.updateAddress);
			break;
		default:
			await storeResult(dakar.data.blockchainSearchQueryGet({query: trimmedQuery}), 'updateSearchResult', explorerStore.updateSearchResult);
	}

	return true;
}

function setWarningMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'warning', temporary: true, category: route.name,
	});
}

// Initial routing
newRouting();

</script>

<style scoped>

</style>
