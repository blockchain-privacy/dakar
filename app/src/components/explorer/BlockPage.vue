<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-container fluid>
    <v-row class="align-center justify-center">
      <v-col
        md="12"
        xl="8"
      >
        <alert :text="errorMsg" />
        <div v-if="block">
          <v-row>
            <v-col>
              <v-card variant="text">
                <icon-title
                  :title="`Block ${block.blockhash}`"
                  :icon="mdiCubeOutline"
                >
                  <mode-chip :blockchain-mode="route.params.blockchainMode" />
                </icon-title>
                <v-card-text>
                  <v-row>
                    <v-col
                      v-if="block.id"
                      cols="12"
                      sm="6"
                    >
                      <icon-item
                        :icon="mdiFormatListNumbered"
                        title="Block Height"
                      >
                        {{ block.id.toLocaleString() }}
                      </icon-item>
                    </v-col>
                    <v-col v-if="block.ts">
                      <icon-item
                        :icon="mdiCalendar"
                        title="Timestamp"
                      >
                        {{ block.ts != null ? new Date(block.ts).toLocaleString() : "" }}
                      </icon-item>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col
                      v-if="block.prevblockhash"
                      cols="12"
                      sm="6"
                    >
                      <icon-item
                        :icon="mdiFormatHeaderPound"
                        title="Previous Block"
                      >
                        <router-link
                          id="block-page-previous-block"
                          :to="{ name: ROUTE_NAME_BLOCK_PAGE,
                                 params: { id: block.prevblockhash, blockchainMode: route.params.blockchainMode }}"
                        >
                          {{ shortenHash(block.prevblockhash) }}
                        </router-link>
                      </icon-item>
                    </v-col>
                    <v-col v-if="block.nextblockhash">
                      <icon-item
                        :icon="mdiFormatHeaderPound"
                        title="Next Block"
                      >
                        <router-link
                          id="block-page-next-block"
                          :to="{ name: ROUTE_NAME_BLOCK_PAGE,
                                 params: { id: block.nextblockhash, blockchainMode: route.params.blockchainMode }}"
                        >
                          {{ shortenHash(block.nextblockhash) }}
                        </router-link>
                      </icon-item>
                    </v-col>
                  </v-row>
                  <v-row v-if="block.txcount">
                    <v-col>
                      <icon-item
                        :icon="mdiPound"
                        title="Number of Transactions"
                      >
                        {{ block.txcount.toLocaleString() }}
                      </icon-item>
                    </v-col>
                  </v-row>
                </v-card-text>
              </v-card>
            </v-col>
            <v-infinite-scroll
              v-if="block.transactions"
              style="width:100%"
              @load="addNewData"
            >
              <template
                v-for="tx in block.transactions"
                :key="tx.txhash+tx.bid"
              >
                <v-col class="px-1 py-3">
                  <transaction
                    :tx="tx"
                    show-title-link
                    :show-heuristic-editor-link="isPrivilegedOrHigher"
                    :show-fingerprint-link="isPrivilegedOrHigher"
                    show-title-bar
                    embed
                  />
                </v-col>
              </template>
              <template #empty>
                <p class="text-label-medium text-grey">
                  End of transaction list reached
                </p>
              </template>
              <template #error>
                <p class="text-red">
                  Error fetching new transactions
                </p>
              </template>
            </v-infinite-scroll>
          </v-row>
        </div>
        <v-skeleton-loader
          v-else
          type="list-item-three-line, list-item-three-line, list-item-three-line"
        />
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import {
	mdiCubeOutline,
	mdiFormatListNumbered,
	mdiCalendar,
	mdiFormatHeaderPound,
	mdiPound,
} from '@mdi/js';
import {
	computed,
	onMounted,
	ref,
	watch,
} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {storeToRefs} from 'pinia';
import IconItem from '../common/IconItem.vue';
import Transaction from './transaction/Transaction.vue';
import {
	getDakarClients,
	isAdminIdentity,
	isPrivilegedIdentity,
	shortenHash,
} from '@/utilities';
import {PAGE_TITLE, ROUTE_NAME_404_PAGE, ROUTE_NAME_BLOCK_PAGE} from '@/constants';
import IconTitle from '@/components/common/IconTitle.vue';
import {useLocalStore} from '@/pinia/local';
import ModeChip from '@/components/common/ModeChip.vue';
import Alert from '@/components/common/Alert.vue';

const route = useRoute();
const router = useRouter();
const {session} = storeToRefs(useLocalStore());
const block = ref(null);
const errorMsg = ref('');

let offset = 0;

const dakarClients = getDakarClients();

// Computed
const isPrivilegedOrHigher = computed(() => isPrivilegedIdentity(session.value, route.params.blockchainMode)
	|| isAdminIdentity(session.value, route.params.blockchainMode));

// Watchers
watch(route, async () => {
	// If route gets changed the component could still be loaded but now with different data.
	// Because of this the internal state has to be reset.
	offset = 0;
	await pullInitialData();
	setPageTitle();
});

// Hooks
onMounted(async () => {
	await pullInitialData();
	setPageTitle();
	// Register scroll handler
	offset = 0;
});

// Functions
function setPageTitle() {
	let id = ' ';
	if (block.value?.id) {
		id = ` ${block.value.id} `;
	}

	document.title = `Block${id}- ${PAGE_TITLE}`;
}

function isResponseValid(resp) {
	return !(!resp.block || !resp.block.transactions || resp.block.transactions.length === 0);
}

async function pullInitialData() {
	if (route.params.id === undefined) {
		return;
	}

	errorMsg.value = '';
	block.value = null;
	try {
		const response = await dakarClients[route.params.blockchainMode].data
			.blockchainBlocksHashGet({hash: route.params.id});
		if (response.block) {
			block.value = response.block;
		}
	} catch (error) {
		if (error.cause?.status === 404) {
			await router.push({name: ROUTE_NAME_404_PAGE, params: {catchAll: 'invalid'}});
		} else {
			errorMsg.value = error.message;
		}
	}
}

async function addNewData({done}) {
	if (!block.value) {
		done('empty');
		return;
	}

	offset += 10;

	// Do nothing if all data is already loaded
	if (offset >= block.value.txcount) {
		done('empty');
		return;
	}

	errorMsg.value = '';
	try {
		const response = await dakarClients[route.params.blockchainMode].data
			.blockchainBlocksHashGet({hash: block.value.blockhash, offset});

		if (isResponseValid(response)) {
			block.value.transactions = [...block.value.transactions, ...response.block.transactions];
		}

		done('ok');
	} catch (error) {
		errorMsg.value = error.message;
		done('error');
	}
}

</script>
