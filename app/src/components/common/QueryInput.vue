<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-menu
    :open-on-click="false"
    open-on-focus
    scroll-strategy="none"
    max-width="0"
  >
    <template #activator="{props}">
      <v-text-field
        v-model="query"
        v-bind="props"
        :class="$attrs.class"
        :style="$attrs.style"
        hide-details
        :variant="variant"
        :density="density"
        color="primary"
        single-line
        :label="label"
        :rules="[isValidQuery]"
        :loading="isLoading"
        type="input"
        :append-inner-icon="mdiMagnify"
        @update:model-value="queueSearch"
        @click:append-inner="handleDirectSearch"
        @keydown.enter="handleDirectSearch"
      />
    </template>
    <template v-if="!isLoading">
      <query-input-results
        v-if="!query && !hideHistory && getSearchHistory.length > 0"
        :items="getSearchHistory"
        @item-clicked="handleResultItemClick"
      />
      <v-list v-else-if="resultItems.empty">
        <v-list-item title="No results" />
      </v-list>
      <query-input-results
        v-else-if="resultItems.length > 0"
        :items="resultItems"
        @item-clicked="handleResultItemClick"
      />
    </template>
  </v-menu>
</template>

<script setup>
import {
	BLOCKCHAIN_ATTRIBUTES,
	ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_TRANSACTION_PAGE,
} from '@/constants/index.js';
import {getDakarClients} from '@/utilities/index.js';
import {ref} from 'vue';
import {useRouter} from 'vue-router';
import {useMsgStore} from '@/pinia/msg.js';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';
import {mdiMagnify} from '@mdi/js';
import QueryInputResults from '@/components/common/QueryInputResults.vue';

const localStore = useLocalStore();
const {getSearchHistory} = storeToRefs(localStore);
const router = useRouter();
const msgStore = useMsgStore();

defineProps({
	density: {type: String, required: false, default: undefined},
	variant: {type: String, required: false, default: 'solo'},
});

const dakarClients = getDakarClients();
const isLoading = ref(false);
const resultItems = ref([]);
const hideHistory = ref(false);
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

	lastQuery = trimmed;

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

function handleResultItemClick(item) {
	// The menu should be hidden by removing focus (blur()) from the text field.
	// However, when calling blur it shows the history list for a split second.
	// To work around this, set a flag to false and set it to true shortly after
	hideHistory.value = true;
	setTimeout(() => {
		hideHistory.value = false;
	}, 500);

	document.activeElement.blur();
	localStore.addSearchHistoryItem(item);
	router.push(getResultNavigation(item));
	query.value = '';
	resultItems.value = [];
}

</script>

<style scoped>

/*
VMenu makes the input have the wrong cursor
*/
:deep( .v-field__input ) {
  cursor: text !important;
}

</style>
