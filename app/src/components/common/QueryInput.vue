<template>
  <v-menu
    open-on-click
    open-on-focus
    close-on-content-click
    max-width="0"
  >
    <template #activator="{props}">
      <v-text-field
        v-model="query"
        v-bind="props"
        :class="$attrs.class"
        :style="$attrs.style"
        hide-details
        :append-inner-icon="mdiMagnify"
        :variant="variant"
        :density="density"
        color="primary"
        single-line
        :label="label"
        :rules="[isValidQuery]"
        :loading="isLoading"
        type="input"
        @update:model-value="queueSearch"
      />
    </template>
    <v-list v-if="!isLoading && resultItems.empty">
      <v-list-item title="No results" />
    </v-list>
    <v-list
      v-else-if="!isLoading && resultItems.length > 0"
    >
      <v-list-item
        v-for="(item, index) in resultItems"
        :key="index"
        @click="handleResultItemClick(item.response,item.mode)"
      >
        <template #prepend>
          <v-icon
            :icon="BLOCKCHAIN_ATTRIBUTES[item.mode].logo"
            :color="BLOCKCHAIN_ATTRIBUTES[item.mode].color"
          />
        </template>
        <template #append>
          <v-chip v-if="getResultType(item.response.type)">
            {{ getResultType(item.response.type) }}
          </v-chip>
        </template>
        <div class="shorten">
          {{ getResultTitle(item.response) }}
        </div>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script setup>
import {mdiMagnify} from '@mdi/js';
import {
	BLOCKCHAIN_ATTRIBUTES,
	BLOCKCHAIN_BTC, BLOCKCHAIN_DASH, ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_TRANSACTION_PAGE,
} from '@/constants/index.js';
import {getDakarClients} from '@/utilities/index.js';
import {ref} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {useMsgStore} from '@/pinia/msg.js';

const router = useRouter();
const route = useRoute();
const msgStore = useMsgStore();

defineProps({
	density: {type: String, required: false, default: undefined},
	variant: {type: String, required: false, default: 'solo'},
});

// When the blockchain mode is switched and the current component is not reloaded,
// the dakar client is in the wrong state. As a workaround, get all available dakar
// clients and select the right one when doing a request.
const dakarClients = getDakarClients();
const isLoading = ref(false);
const resultItems = ref([]);
const query = ref('');
const label = 'Search for blocks, transactions and addresses';
let searchTimer = null;

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

async function search(q) {
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
	resultItems.value = [];

	try {
		const btcResponse = await dakarClients[BLOCKCHAIN_BTC].data.blockchainSearchQueryGet({query: trimmed});
		resultItems.value.push({
			mode: BLOCKCHAIN_BTC, response: btcResponse, value: q, title: q,
		});
	} catch (_) {
		// Just do nothing
	}

	try {
		const dashResponse = await dakarClients[BLOCKCHAIN_DASH].data.blockchainSearchQueryGet({query: trimmed});
		resultItems.value.push({
			mode: BLOCKCHAIN_DASH, response: dashResponse, value: q, title: q,
		});
	} catch (_) {
		// Just do nothing
	}

	if (resultItems.value.length === 0) {
		// Both request returned not data
		resultItems.value = {empty: true};
	}

	isLoading.value = false;
}

function queueSearch(q) {
	if (searchTimer !== null) {
		clearTimeout(searchTimer);
	}

	searchTimer = setTimeout(search, 700, q);
}

function setWarningMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'warning', temporary: true, category: route.name,
	});
}

function getResultTitle(item) {
	switch (item.type) {
		case 'tx': return item.payload[0].txhash;
		case 'block': return item.payload.id;
		case 'addr': return item.payload.addresshash;
		default:
			return '';
	}
}

function getResultNavigation(item, mode) {
	switch (item.type) {
		case 'tx': return {name: ROUTE_NAME_TRANSACTION_PAGE, params: {id: item.payload[0].txhash, blockchainMode: mode}};
		case 'block': return {name: ROUTE_NAME_BLOCK_PAGE, params: {id: item.payload.id, blockchainMode: mode}};
		case 'addr': return {name: ROUTE_NAME_ADDRESS_PAGE, params: {id: item.payload.id, blockchainMode: mode}};
		default:
			return {};
	}
}

function getResultType(type) {
	switch (type) {
		case 'tx': return 'Transaction';
		case 'block': return 'Block';
		case 'addr': return 'Address';
		default:
			return '';
	}
}

function handleResultItemClick(item, mode) {
	router.push(getResultNavigation(item, mode));
	query.value = '';
	resultItems.value = [];
}

</script>

<style scoped>

</style>
