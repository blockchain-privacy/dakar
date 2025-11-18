<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-container fluid>
    <v-row
      align="center"
      justify="center"
    >
      <v-col
        cols="12"
        sm="12"
        md="12"
        lg="10"
        xl="8"
      >
        <template v-if="transactions.length > 0">
          <fade-transition
            v-for="t in transactions"
            :key="t.txhash+t.bid"
          >
            <!-- duplicate transaction hashes can exist -> loop through all results
               (e.g. d5d27987d2a3dfc724e359870c6644b40e497bdc0589a033220fe15429d88599 in Bitcoin) -->
            <transaction
              :tx="t"
              :show-heuristic-editor-link="isPrivilegedOrHigher"
              :show-fingerprint-link="isPrivilegedOrHigher"
              show-title-bar
              show-details
              show-mode
            />
          </fade-transition>
        </template>
        <v-skeleton-loader
          v-else
          type="list-item-three-line, list-item-three-line, list-item-three-line"
        />
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import Transaction from './Transaction.vue';
import {PAGE_TITLE, ROUTE_NAME_404_PAGE} from '@/constants';
import {
	getDakarClients, handleError, isAdminIdentity, isPrivilegedIdentity,
} from '@/utilities';
import {
	computed, onMounted, ref, watch,
} from 'vue';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local';
import FadeTransition from '@/components/common/FadeTransition.vue';
import {useRoute, useRouter} from 'vue-router';
import {useMsgStore} from '@/pinia/msg.js';

const {session} = storeToRefs(useLocalStore());
const route = useRoute();
const router = useRouter();
const msgStore = useMsgStore();
const dakarClients = getDakarClients();

const transactions = ref([]);
const context = {$route: route, addMessage: msgStore.addMessage};

// Computed
const isPrivilegedOrHigher = computed(() => isPrivilegedIdentity(session.value, route.params.blockchainMode)
	|| isAdminIdentity(session.value, route.params.blockchainMode));

// Watchers
watch(route, async () => {
	await pullInitialData();
	setPageTitle();
});

// Hooks
onMounted(async () => {
	await pullInitialData();
	setPageTitle();
});

// Functions
function setPageTitle() {
	let h = ' ';
	if (transactions.value && transactions.value.length > 0 && transactions.value[0].txhash) {
		h = ` ${transactions.value[0].txhash} `;
	}

	document.title = `Transaction${h}- ${PAGE_TITLE}`;
}

async function pullInitialData() {
	if (route.params.id === undefined) {
		return;
	}

	transactions.value = [];
	try {
		const response = await dakarClients[route.params.blockchainMode].data
			.blockchainTransactionsHashGet({hash: route.params.id});
		if (response.transactions) {
			transactions.value = response.transactions;
		}
	} catch (e) {
		if (e.cause?.status === 404) {
			await router.push({name: ROUTE_NAME_404_PAGE, params: {catchAll: 'invalid'}});
		} else {
			handleError(context, e);
		}
	}
}
</script>
