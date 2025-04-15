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
        @click:append-inner="handleDirectSearch"
        @keydown.enter="handleDirectSearch"
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
        @click="handleResultItemClick(item)"
      >
        <template #prepend>
          <v-icon
            :icon="BLOCKCHAIN_ATTRIBUTES[item.mode].icon"
            :color="BLOCKCHAIN_ATTRIBUTES[item.mode].color"
          />
        </template>
        <template #append>
          <v-chip v-if="getResultType(item.type)">
            {{ getResultType(item.type) }}
          </v-chip>
        </template>
        <div class="shorten">
          {{ item.title }}
        </div>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script setup>
import {mdiMagnify} from '@mdi/js';
import {
	BLOCKCHAIN_ATTRIBUTES,
	ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_TRANSACTION_PAGE,
} from '@/constants/index.js';
import {getDakarClients} from '@/utilities/index.js';
import {ref} from 'vue';
import {useRouter} from 'vue-router';
import {useMsgStore} from '@/pinia/msg.js';

const router = useRouter();
const msgStore = useMsgStore();

defineProps({
	density: {type: String, required: false, default: undefined},
	variant: {type: String, required: false, default: 'solo'},
});

const dakarClients = getDakarClients();
const isLoading = ref(false);
const resultItems = ref([]);
const query = ref('');
const label = 'Search for blocks, transactions and addresses';
let searchTimer = null;
let lastQuery = '';

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

	if (lastQuery === trimmed) {
		return;
	}

	if (!isValidQueryInput(trimmed)) {
		setNoResults();
		return;
	}

	msgStore.resetMessages();

	isLoading.value = true;

	resultItems.value = [];
	const searchResults = [];
	const blockchainKeys = Object.keys(BLOCKCHAIN_ATTRIBUTES);
	const resolved = await Promise.allSettled(blockchainKeys.map(chain => dakarClients[chain].data
		.blockchainSearchQueryGet({query: trimmed})));

	for (const [index, response] of resolved.entries()) {
		if (response.status === 'rejected') {
			// ignore error
			continue;
		}

		searchResults.push({
			mode: blockchainKeys[index], type: response.value.type, value: trimmed, title: trimmed,
		});
	}

	if (searchResults.length === 0) {
		// Both request returned not data
		setNoResults();
	} else {
		resultItems.value = searchResults;
	}

	lastQuery = trimmed;

	isLoading.value = false;
}

function setNoResults() {
	resultItems.value = {empty: true};
}

function queueSearch(q) {
	if (searchTimer !== null) {
		clearTimeout(searchTimer);
	}

	searchTimer = setTimeout(search, 700, q);
}

async function handleDirectSearch() {
	if (`${query.value.trim()}` !== lastQuery) {
		// Results are not recent
		if (searchTimer !== null) {
			clearTimeout(searchTimer);
		}

		await search(query.value);
	}

	if (resultItems.value.empty || resultItems.value.length === 0) {
		return;
	}

	const item = resultItems.value[0];
	handleResultItemClick(item);
}

function getResultNavigation(item) {
	switch (item.type) {
		case 'tx': return {name: ROUTE_NAME_TRANSACTION_PAGE, params: {id: item.title, blockchainMode: item.mode}};
		case 'block': return {name: ROUTE_NAME_BLOCK_PAGE, params: {id: item.title, blockchainMode: item.mode}};
		case 'addr': return {name: ROUTE_NAME_ADDRESS_PAGE, params: {id: item.title, blockchainMode: item.mode}};
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

function handleResultItemClick(item) {
	router.push(getResultNavigation(item));
	query.value = '';
	resultItems.value = [];
}

</script>

<style scoped>

</style>
