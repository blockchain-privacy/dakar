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
                    :icon="icon.mdiCubeOutline"
                  />
                  <v-card-text>
                    <v-row>
                      <v-col
                        v-if="data.id"
                        cols="12"
                        sm="6"
                      >
                        <icon-item
                          :icon="icon.mdiFormatListNumbered"
                          title="Block Height"
                        >
                          {{ data.id.toLocaleString() }}
                        </icon-item>
                      </v-col>
                      <v-col v-if="data.ts">
                        <icon-item
                          :icon="icon.mdiCalendar"
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
                          :icon="icon.mdiFormatHeaderPound"
                          title="Previous Block"
                        >
                          <router-link
                            :to="{ name: blockRoute,
                                   params: { id: data.prevblockhash }}"
                          >
                            {{ shortenHash(data.prevblockhash) }}
                          </router-link>
                        </icon-item>
                      </v-col>
                      <v-col v-if="data.nextblockhash">
                        <icon-item
                          :icon="icon.mdiFormatHeaderPound"
                          title="Next Block"
                        >
                          <router-link
                            :to="{ name: blockRoute,
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
                          :icon="icon.mdiPound"
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
                          :show-heuristic-editor-link="showHeuristicEditor"
                          :show-fingerprint-link="showHeuristicEditor"
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

<script>
import {
	mdiCubeOutline, mdiFormatListNumbered, mdiCalendar,
	mdiFormatHeaderPound, mdiTransfer, mdiPound,
} from '@mdi/js';
import {
	handleError, isAdminIdentity, isPrivilegedIdentity, shortenHash,
} from '@/utilities';
import {
	PAGE_TITLE,
	ROUTE_NAME_BLOCK_PAGE,
	ROUTE_NAME_TRANSACTION_PAGE,
} from '@/constants';
import IconItem from '../common/IconItem.vue';
import Transaction from './transaction/Transaction.vue';
import FadeTransition from '../common/FadeTransition.vue';
import IconTitle from '@/components/common/IconTitle.vue';

export default {
	name: 'BlockPage',
	components: {IconTitle, FadeTransition, IconItem, Transaction},
	data() {
		return {
			icon: {
				mdiCubeOutline,
				mdiFormatListNumbered,
				mdiCalendar,
				mdiFormatHeaderPound,
				mdiTransfer,
				mdiPound,
			},
			blockRoute: ROUTE_NAME_BLOCK_PAGE,
			transactionRoute: ROUTE_NAME_TRANSACTION_PAGE,
			offset: 0,
		};
	},
	computed: {
		data() {
			return this.$store.getters.getBlockData;
		},
		session() {
			return this.$store.getters.getSession;
		},
		showHeuristicEditor() {
			return isPrivilegedIdentity(this.session) || isAdminIdentity(this.session);
		},
	},
	watch: {
		$route() {
			// If route gets changed the component could still be loaded but now with different data.
			// Because of this the internal state has to be reset.
			this.offset = 0;
		},
		data() {
			this.setPageTitle();
		},
	},
	mounted() {
		this.setPageTitle();
		// Register scroll handler
		this.offset = 0;
	},
	updated() {
		this.setPageTitle();
	},
	methods: {
		shortenHash,
		setPageTitle() {
			let id = ' ';
			if (this.data && this.data.id) {
				id = ` ${this.data.id} `;
			}

			document.title = `Block${id}- ${PAGE_TITLE}`;
		},
		isResponseValid(data) {
			return !(!data.type || data.type !== 'block' || !data.payload || !data.payload.transactions
          || data.payload.transactions.length === 0);
		},
		async addNewData({done}) {
			if (!this.data) {
				done('empty');
				return;
			}

			this.offset += 10;

			// Do nothing if all data is already loaded
			if (this.offset >= this.data.txcount) {
				done('empty');
				return;
			}

			try {
				const response = await this.dakar.data.blkRangeBlockHashPost({blockHash: this.data.blockhash, offset: {offset: this.offset}});

				if (this.isResponseValid(response)) {
					this.data.transactions = [...this.data.transactions, ...response.payload.transactions];
					this.$store.dispatch('resetMessages');
				}

				done('ok');
			} catch (e) {
				handleError(this, e);
				done('error');
			}
		},
	},
};
</script>
