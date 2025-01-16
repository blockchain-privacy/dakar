<template>
  <v-text-field
    v-model="query"
    hide-details
    :append-inner-icon="mdiMagnify"
    :variant="variant"
    :density="density"
    color="primary"
    single-line
    :label="label"
    :rules="[isValidQuery]"
    :loading="isLoading"
    :disabled="isLoading"
    @click:append-inner="handleInput(query)"
    @keydown.enter="handleInput(query)"
  />
</template>

<script setup>
import {mdiMagnify} from '@mdi/js';
import {
	RESPONSE_EMPTY, RESPONSE_TYPE_ADDRESS, RESPONSE_TYPE_BLOCK, RESPONSE_TYPE_TRANSACTION,
	ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_NO_RESULTS, ROUTE_NAME_TRANSACTION_PAGE,
} from '@/constants/index.js';
import {
	getDakarClients, handleError, handleQuery,
} from '@/utilities/index.js';
import {computed, ref} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {useExplorerStore} from '@/pinia/explorer.js';
import {useMsgStore} from '@/pinia/msg.js';
import {useNavStore} from '@/pinia/nav.js';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';

const route = useRoute();
const router = useRouter();
const msgStore = useMsgStore();
const {getSettings} = storeToRefs(useLocalStore());
const explorerStore = useExplorerStore();
const {pushFromUserInput} = storeToRefs(useNavStore());
const context = {addMessage: msgStore.addMessage, $route: useRoute()};

defineProps({
	density: {type: String, required: false, default: undefined},
	variant: {type: String, required: false, default: 'solo'},
});

// When the blockchain mode is switched and the current component is not reloaded,
// the dakar client is in the wrong state. As a workaround, get all available dakar
// clients and select the right one when doing a request.
const dakarClients = getDakarClients();
const query = ref('');
const isLoading = ref(false);
const label = 'Search for blocks, transactions and addresses';

// Computed
const searchResultType = computed(() => explorerStore.getSearchResultType);

// Functions
function isValidQueryInput(str) {
	const inputLen = str.length;
	// 64 -> length of transaction hash and block hash
	if (inputLen === 0 || inputLen > 64) {
		return false;
	}

	// 33,34 -> address length; if smaller than it must be a block id
	if (inputLen < 33) {
		return Number.isInteger(Number(str));
	}

	return str.match(/^[\da-zA-Z]+$/) !== null;
}

function isValidQuery(q) {
	// Template string in case it is a number
	const trimmed = `${q}`.trim();
	return trimmed.length === 0 ? true : isValidQueryInput(trimmed);
}

async function handleInput(q) {
	// Template string in case it is a number
	const trimmed = `${q}`.trim();
	if (!trimmed) {
		return;
	}

	if (!isValidQueryInput(trimmed)) {
		setWarningMessage('Query is not valid');
		return;
	}

	msgStore.resetMessages();

	isLoading.value = true;
	const err = await handleQuery(trimmed, explorerStore, dakarClients[getSettings.value.blockchainMode]);
	isLoading.value = false;
	query.value = '';
	if (err) {
		if (err.cause?.status === 404) {
			await router.push({name: ROUTE_NAME_NO_RESULTS});
		} else {
			handleError(context, err);
		}

		return;
	}

	// Route to corresponding page
	switch (searchResultType.value) {
		case RESPONSE_EMPTY:
			await router.push({name: ROUTE_NAME_NO_RESULTS});
			break;
		case RESPONSE_TYPE_ADDRESS:
			pushFromUserInput.value = true;
			await router.push({name: ROUTE_NAME_ADDRESS_PAGE, params: {id: trimmed, blockchainMode: getSettings.value.blockchainMode}});
			break;
		case RESPONSE_TYPE_BLOCK:
			pushFromUserInput.value = true;
			await router.push({name: ROUTE_NAME_BLOCK_PAGE, params: {id: trimmed, blockchainMode: getSettings.value.blockchainMode}});
			break;
		case RESPONSE_TYPE_TRANSACTION:
			pushFromUserInput.value = true;
			await router.push({name: ROUTE_NAME_TRANSACTION_PAGE, params: {id: trimmed, blockchainMode: getSettings.value.blockchainMode}});
			break;
		default:
			await router.push({name: ROUTE_NAME_NO_RESULTS});
			break;
	}
}

function setWarningMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'warning', temporary: true, category: route.name,
	});
}

</script>

<style scoped>

</style>
