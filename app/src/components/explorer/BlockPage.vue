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
        <fade-transition>
          <div v-if="block">
            <v-row>
              <v-col>
                <v-card variant="text">
                  <icon-title
                    :title="`Block ${block.blockhash}`"
                    :icon="mdiCubeOutline"
                  />
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
                                   params: { id: block.prevblockhash, blockchainMode: getSettings.blockchainMode }}"
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
                                   params: { id: block.nextblockhash, blockchainMode: getSettings.blockchainMode }}"
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
              <template v-if="block.transactions">
                <v-divider />
                <v-container class="pa-0">
                  <v-infinite-scroll @load="addNewData">
                    <template
                      v-for="tx in block.transactions"
                      :key="tx.txhash+tx.bid"
                    >
                      <v-col class="px-0">
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
                      <p class="text-overline text-grey">
                        End of transaction list reached
                      </p>
                    </template>
                    <template #error>
                      <p class="text-h5 text-red">
                        Error fetching new transactions
                      </p>
                    </template>
                  </v-infinite-scroll>
                </v-container>
              </template>
            </v-row>
          </div>
          <v-skeleton-loader
            v-else
            type="list-item-three-line, list-item-three-line, list-item-three-line"
          />
        </fade-transition>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import {
	mdiCubeOutline, mdiFormatListNumbered, mdiCalendar,
	mdiFormatHeaderPound, mdiPound,
} from '@mdi/js';
import {
	handleError, isAdminIdentity, isPrivilegedIdentity, shortenHash,
} from '@/utilities';
import {PAGE_TITLE, ROUTE_NAME_BLOCK_PAGE} from '@/constants';
import IconItem from '../common/IconItem.vue';
import Transaction from './transaction/Transaction.vue';
import FadeTransition from '../common/FadeTransition.vue';
import IconTitle from '@/components/common/IconTitle.vue';
import {
	computed, onMounted, onUpdated, watch,
} from 'vue';
import {useRoute} from 'vue-router';
import {useExplorerStore} from '@/pinia/explorer';
import {storeToRefs} from 'pinia';
import {useMsgStore} from '@/pinia/msg';
import {useLocalStore} from '@/pinia/local';
import {getDakarClient} from '@/utilities';

const route = useRoute();
const msgStore = useMsgStore();
const context = {$route: route, addMessage: msgStore.addMessage};
const {block} = storeToRefs(useExplorerStore());
const {session, getSettings} = storeToRefs(useLocalStore());

let offset = 0;

const dakar = getDakarClient(getSettings.value.blockchainMode);

// Computed
const isPrivilegedOrHigher = computed(() => isPrivilegedIdentity(session.value, getSettings.value.blockchainMode)
	|| isAdminIdentity(session.value, getSettings.value.blockchainMode));

// Watchers
watch(route, () => {
	// If route gets changed the component could still be loaded but now with different data.
	// Because of this the internal state has to be reset.
	offset = 0;
});

watch(block, () => {
	setPageTitle();
});

// Hooks
onMounted(() => {
	setPageTitle();
	// Register scroll handler
	offset = 0;
});

onUpdated(() => {
	setPageTitle();
});

// Functions
function setPageTitle() {
	let id = ' ';
	if (block.value && block.value.id) {
		id = ` ${block.value.id} `;
	}

	document.title = `Block${id}- ${PAGE_TITLE}`;
}

function isResponseValid(reponse) {
	return !(!reponse.block || !reponse.block.transactions || reponse.block.transactions.length === 0);
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

	try {
		const response = await dakar.data.blockchainBlocksHashGet({hash: block.value.blockhash, offset});

		if (isResponseValid(response)) {
			block.value.transactions = [...block.value.transactions, ...response.block.transactions];
			msgStore.resetMessages();
		}

		done('ok');
	} catch (e) {
		handleError(context, e);
		done('error');
	}
}

</script>
