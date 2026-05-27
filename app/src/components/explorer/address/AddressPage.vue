<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-container fluid>
    <v-row class="align-center justify-center">
      <v-col
        md="12"
        xl="8"
      >
        <alert :text="errorMsg" />
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
import {onMounted, ref, watch} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {PAGE_TITLE, ROUTE_NAME_404_PAGE} from '@/constants';
import AddressView from '@/components/explorer/address/Address.vue';
import {getDakarClients} from '@/utilities/index.js';
import Alert from '@/components/common/Alert.vue';

const route = useRoute();
const router = useRouter();
const dakarClients = getDakarClients();

const address = ref(null);
const errorMsg = ref('');

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
	errorMsg.value = '';
	try {
		const response = await dakarClients[route.params.blockchainMode].data
			.blockchainAddressesHashGet({hash: route.params.id});
		if (response.address) {
			address.value = response.address;
		}
	} catch (error) {
		if (error.cause?.status === 404) {
			await router.push({name: ROUTE_NAME_404_PAGE, params: {catchAll: 'invalid'}});
		} else {
			errorMsg.value = error.message;
		}
	}
}

</script>

<style scoped>

</style>
