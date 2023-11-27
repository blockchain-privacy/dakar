<template>
  <v-text-field
    v-model="query"
    style="margin: 23px 0 0 0;min-width:220px"
    variant="outlined"
    density="compact"
    color="primary"
    single-line
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
import {computed, inject, ref, watch} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {useStore} from 'vuex';

const dakar = inject('dakar');
const route = useRoute();
const router = useRouter();
const store = useStore();
const context = {$store: useStore(), $route: useRoute()};

const query = ref('');
let lastQuery = '';

watch(route, () => {
	lastQuery = '';
	newRouting();
});

// Computed
const searchResultType = computed(() => store.getters.getSearchResultType);

const isPushFromUserInput = computed({
	async set(value) {
		await store.dispatch('setPushFromUserInput', value);
	},
	get() {
		return store.getters.getPushFromUserInput;
	},
});

// Functions
function newRouting() {
	const {id} = route.params;
	const pushFromUserInput = isPushFromUserInput.value;

	if (pushFromUserInput) {
		isPushFromUserInput.value = false;
	}

	if (pushFromUserInput || !id
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
			isPushFromUserInput.value = true;
			await router.push({name: ROUTE_NAME_ADDRESS_PAGE, params: {id: trimmedQuery}});
			break;
		case RESPONSE_TYPE_BLOCK:
			isPushFromUserInput.value = true;
			await router.push({name: ROUTE_NAME_BLOCK_PAGE, params: {id: trimmedQuery}});
			break;
		case RESPONSE_TYPE_TRANSACTION:
			isPushFromUserInput.value = true;
			await router.push({name: ROUTE_NAME_TRANSACTION_PAGE, params: {id: trimmedQuery}});
			break;
		default:
			await router.push({name: ROUTE_NAME_NO_RESULTS});
			break;
	}
}

async function storeResult(promise, action) {
	try {
		const response = await promise;
		await store.dispatch(action, response);
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

	await store.dispatch('resetMessages');
	await store.dispatch('setBlockData', null);
	await store.dispatch('setTransactionData', null);
	await store.dispatch('setAddressData', null);

	switch (type) {
		case RESPONSE_TYPE_TRANSACTION:
			await storeResult(dakar.data.txHashGet({hash: trimmedQuery}), 'updateTransactionData');
			break;
		case RESPONSE_TYPE_BLOCK:
			await storeResult(dakar.data.blkHashGet({hash: trimmedQuery}), 'updateBlockData');
			break;
		case RESPONSE_TYPE_ADDRESS:
			await storeResult(dakar.data.addressHashGet({hash: trimmedQuery}), 'updateAddressData');
			break;
		default:
			await storeResult(dakar.data.searchQueryGet({query: trimmedQuery}), 'updateSearchResult');
	}

	return true;
}

function setWarningMessage(msg) {
	store.dispatch('addMessage', {text: msg, type: 'warning', temporary: true, category: route.name});
}

// Initial routing
newRouting();

</script>

<style scoped>

</style>
