<template>
  <v-container :fluid="true">
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
          <div v-if="data">
            <v-row>
              <v-col>
                <v-card variant="text">
                  <icon-title
                    :title="`Block ${data.blockhash}`"
                    :icon="mdiCubeOutline"
                  />
                  <v-card-text>
                    <v-row>
                      <v-col
                        v-if="data.id"
                        cols="12"
                        sm="6"
                      >
                        <icon-item
                          :icon="mdiFormatListNumbered"
                          title="Block Height"
                        >
                          {{ data.id.toLocaleString() }}
                        </icon-item>
                      </v-col>
                      <v-col v-if="data.ts">
                        <icon-item
                          :icon="mdiCalendar"
                          title="Timestamp"
                        >
                          {{ data.ts != null ? new Date(data.ts).toLocaleString() : "" }}
                        </icon-item>
                      </v-col>
                    </v-row>
                    <v-row>
                      <v-col
                        v-if="data.prevblockhash"
                        cols="12"
                        sm="6"
                      >
                        <icon-item
                          :icon="mdiFormatHeaderPound"
                          title="Previous Block"
                        >
                          <router-link
                            :to="{ name: ROUTE_NAME_BLOCK_PAGE,
                                   params: { id: data.prevblockhash }}"
                          >
                            {{ shortenHash(data.prevblockhash) }}
                          </router-link>
                        </icon-item>
                      </v-col>
                      <v-col v-if="data.nextblockhash">
                        <icon-item
                          :icon="mdiFormatHeaderPound"
                          title="Next Block"
                        >
                          <router-link
                            :to="{ name: ROUTE_NAME_BLOCK_PAGE,
                                   params: { id: data.nextblockhash }}"
                          >
                            {{ shortenHash(data.nextblockhash) }}
                          </router-link>
                        </icon-item>
                      </v-col>
                    </v-row>
                    <v-row>
                      <v-col>
                        <icon-item
                          :icon="mdiPound"
                          title="Number of Transactions"
                        >
                          {{ data.txcount.toLocaleString() }}
                        </icon-item>
                      </v-col>
                    </v-row>
                  </v-card-text>
                </v-card>
              </v-col>
              <template v-if="data.transactions">
                <v-divider />
                <v-col>
                  <v-infinite-scroll @load="addNewData">
                    <template
                      v-for="tx in data.transactions"
                      :key="tx.txhash+tx.bid"
                    >
                      <v-col>
                        <transaction
                          :tx="tx"
                          show-title-link
                          :show-heuristic-editor-link="isPrivilegedOrHigher"
                          :show-fingerprint-link="isPrivilegedOrHigher"
                          :embed="true"
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
                </v-col>
              </template>
            </v-row>
          </div>
          <v-skeleton-loader
            v-else
            class="mx-auto"
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
import {computed, inject, onMounted, onUpdated, watch} from 'vue';
import {useRoute} from 'vue-router';
import {useStore} from 'vuex';

const dakar = inject('dakar');
const route = useRoute();
const store = useStore();
const context = {$store: store, $route: route};

let offset = 0;

// Computed
const data = computed(() => store.getters.getBlockData);

const session = computed(() => store.getters.getSession);
const isPrivilegedOrHigher = computed(() => isPrivilegedIdentity(session.value) || isAdminIdentity(session.value));

// Watchers
watch(route, () => {
	// If route gets changed the component could still be loaded but now with different data.
	// Because of this the internal state has to be reset.
	offset = 0;
});

watch(data, () => {
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
	if (data.value && data.value.id) {
		id = ` ${data.value.id} `;
	}

	document.title = `Block${id}- ${PAGE_TITLE}`;
}

function 	isResponseValid(data) {
	return !(!data.type || data.type !== 'block' || !data.payload || !data.payload.transactions
      || data.payload.transactions.length === 0);
}

async function addNewData({done}) {
	if (!data.value) {
		done('empty');
		return;
	}

	offset += 10;

	// Do nothing if all data is already loaded
	if (offset >= data.value.txcount) {
		done('empty');
		return;
	}

	try {
		const response = await dakar.data.blkRangeBlockHashPost({blockHash: data.value.blockhash, offset: {offset}});

		if (isResponseValid(response)) {
			data.value.transactions = [...data.value.transactions, ...response.payload.transactions];
			await store.dispatch('resetMessages');
		}

		done('ok');
	} catch (e) {
		handleError(context, e);
		done('error');
	}
}

</script>
