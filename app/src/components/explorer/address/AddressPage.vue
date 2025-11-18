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
        lg="9"
        xl="8"
      >
        <address-view
          v-if="address"
          :address-data="address"
          show-title-bar
          show-mode
        />
        <v-skeleton-loader
          v-else
          type="list-item-three-line, list-item-three-line, list-item-three-line"
        />
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import {PAGE_TITLE, ROUTE_NAME_404_PAGE} from '@/constants';
import {onMounted, ref, watch} from 'vue';
import AddressView from '@/components/explorer/address/Address.vue';
import {useRoute, useRouter} from 'vue-router';
import {useMsgStore} from '@/pinia/msg.js';
import {getDakarClients, handleError} from '@/utilities/index.js';

const route = useRoute();
const router = useRouter();
const msgStore = useMsgStore();
const dakarClients = getDakarClients();

const address = ref(null);
const context = {$route: route, addMessage: msgStore.addMessage};

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
	let h = '';

	// Detect if address hash has changed
	if (address.value?.addresshash) {
		h = `${address.value.addresshash} `;
	}

	document.title = `Address ${h}- ${PAGE_TITLE}`;
}

async function pullInitialData() {
	if (route.params.id === undefined) {
		return;
	}

	address.value = null;
	try {
		const response = await dakarClients[route.params.blockchainMode].data
			.blockchainAddressesHashGet({hash: route.params.id});
		if (response.address) {
			address.value = response.address;
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

<style scoped>

</style>
