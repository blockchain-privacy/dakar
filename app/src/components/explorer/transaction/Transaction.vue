<template>
  <v-card :variant="embed?undefined:'text'">
    <icon-title
      :title="`Transaction ${tx.txhash}`"
      :icon="icons.mdiTransfer"
      :to="showTitleLink?{ name: routes.ROUTE_NAME_TRANSACTION_PAGE, params: { id: tx.txhash }}:null"
    >
      <privacy-chip
        v-if="tx.privacytype > 0"
        class="mx-3"
        :privacy-type="tx.privacytype"
      />
      <template v-if="isDestination(tx.privacytype)">
        <v-btn
          v-if="showHeuristicEditorLink"
          :id="`btn_open_heuristic_editor_${tx.txhash}`"
          style="margin-right: 0"
          icon
          :color="null"
          variant="text"
          :to="{ name: routes.ROUTE_NAME_HEURISTIC_PAGE,params: { id: tx.txhash } }"
        >
          <v-icon>{{ icons.mdiGraph }}</v-icon>
        </v-btn>
        <v-tooltip
          location="bottom"
          :activator="`#btn_open_heuristic_editor_${tx.txhash}`"
        >
          <span>Open the heuristic editor for this transaction</span>
        </v-tooltip>
        <v-btn
          v-if="showFingerprintLink"
          :id="`btn_find_similar_transactions_${tx.txhash}`"
          style="margin-right: 0"
          icon
          :color="null"
          variant="text"
          @click="showFingerprintDialog = true"
        >
          <v-icon>{{ icons.mdiFingerprint }}</v-icon>
        </v-btn>
        <v-tooltip
          location="bottom"
          :activator="`#btn_find_similar_transactions_${tx.txhash}`"
        >
          <span>Search for similar destination transactions</span>
        </v-tooltip>
      </template>
    </icon-title>
    <v-card-text>
      <v-expand-transition>
        <div v-if="showTransactionDetails">
          <v-row>
            <v-col
              cols="12"
              sm="6"
            >
              <IconItem
                :icon="icons.mdiFormatListNumbered"
                title="Block Height"
              >
                <router-link :to="{ name: routes.ROUTE_NAME_BLOCK_PAGE, params: { id: tx.bid }}">
                  {{ tx.bid.toLocaleString() }}
                </router-link>
              </IconItem>
            </v-col>
            <v-col>
              <IconItem
                :icon="icons.mdiCalendar"
                title="Timestamp"
              >
                {{ new Date(tx.bts).toLocaleString() }}
              </IconItem>
            </v-col>
          </v-row>
          <v-row>
            <v-col
              v-if="(tx.fee || tx.fee === 0) && tx.fee >= 0"
              cols="12"
              sm="6"
            >
              <IconItem
                :icon="icons.mdiCash"
                title="Fee"
              >
                {{ convertAmount(tx.fee) }}
              </IconItem>
            </v-col>
            <v-col>
              <IconItem
                :icon="icons.mdiFormatHeaderPound"
                title="Block"
              >
                <router-link :to="{ name: routes.ROUTE_NAME_BLOCK_PAGE, params: { id: tx.bhash }}">
                  {{ shortenHash(tx.bhash) }}
                </router-link>
              </IconItem>
            </v-col>
          </v-row>
          <v-row v-if="isCoinBaseTx(tx)">
            <v-col>
              <IconItem
                :icon="icons.mdiPickaxe"
                title="Coinbase"
              >
                yes
              </IconItem>
            </v-col>
          </v-row>
          <!-- bottom spacer for transition -->
          <div style="height: 10px" />
        </div>
      </v-expand-transition>
      <v-btn
        variant="text"
        :block="true"
        size="x-small"
        style="margin-top:-16px;"
        @click="showTransactionDetails = !showTransactionDetails"
      >
        <v-icon>{{ showTransactionDetails ? icons.mdiChevronUp : icons.mdiChevronDown }}</v-icon>
      </v-btn>
      <v-row>
        <v-col v-if="tx.inputs">
          <p class="ml-2">
            {{ getLabel(tx.inputs.length, 'Input') }}
          </p>
          <template
            v-for="(i,y) in getInputs"
            :key="i.addresshash + i.inputindex"
          >
            <output-item
              :is-input="true"
              :amount="i.amount"
              :address-hash="i.addresshash"
              :tx-hash="i.txhash"
              :sig-asm="i.sigasm"
              :key-asm="i.keyasm"
              :output-index="i.outputindex"
              :input-index="i.inputindex"
              :timestamp="i.ts"
              :privacy-type="Number(i.privacytype)"
            />
            <v-divider
              v-if="y+1 < getInputs.length"
              :thickness="2"
            />
          </template>
          <!-- split in two for nicer transition -->
          <v-expand-transition>
            <div v-if="showAllOutputs">
              <v-divider :thickness="2" />
              <template
                v-for="(i,y) in getResidualInputs"
                :key="i.addresshash + i.inputindex"
              >
                <OutputItem

                  :is-input="true"
                  :amount="i.amount"
                  :address-hash="i.addresshash"
                  :tx-hash="i.txhash"
                  :sig-asm="i.sigasm"
                  :key-asm="i.keyasm"
                  :output-index="i.outputindex"
                  :input-index="i.inputindex"
                  :timestamp="i.ts"
                  :privacy-type="Number(i.privacytype)"
                />
                <v-divider
                  v-if="y+1 < getResidualInputs.length"
                  :thickness="2"
                />
              </template>
            </div>
          </v-expand-transition>
        </v-col>
        <!-- empty col if no inputs exist -->
        <v-col v-else />
        <v-col v-if="tx.outputs">
          <p class="ml-2">
            {{ getLabel(tx.outputs.length, 'Output') }}
          </p>
          <template
            v-for="(i,y) in getOutputs"
            :key="i.addresshash + i.outputindex"
          >
            <OutputItem
              :is-input="false"
              :amount="i.amount"
              :address-hash="i.addresshash"
              :tx-hash="i.txhash"
              :sig-asm="i.sigasm"
              :key-asm="i.keyasm"
              :output-index="i.outputindex"
              :input-index="i.inputindex"
              :timestamp="i.ts"
              :privacy-type="Number(i.privacytype)"
            />
            <v-divider
              v-if="y+1<getOutputs.length"
              :thickness="2"
            />
          </template>

          <!-- split in two for nicer transition -->
          <v-expand-transition>
            <div v-if="showAllOutputs">
              <v-divider :thickness="2" />
              <template
                v-for="(i,y) in getResidualOutputs"
                :key="i.addresshash + i.outputindex"
              >
                <OutputItem
                  :is-input="false"
                  :amount="i.amount"
                  :address-hash="i.addresshash"
                  :tx-hash="i.txhash"
                  :sig-asm="i.sigasm"
                  :key-asm="i.keyasm"
                  :output-index="i.outputindex"
                  :input-index="i.inputindex"
                  :timestamp="i.ts"
                  :privacy-type="Number(i.privacytype)"
                />
                <v-divider
                  v-if="y+1<getResidualOutputs.length"
                  :thickness="2"
                />
              </template>
            </div>
          </v-expand-transition>
        </v-col>
        <!-- empty col if no outputs exist -->
        <v-col v-else />
      </v-row>
    </v-card-text>
    <v-btn
      v-if="areItemsLimited"
      variant="text"
      :block="true"
      size="x-small"
      @click="showAllOutputs = !showAllOutputs"
    >
      <v-icon>{{ showAllOutputs ? icons.mdiChevronUp : icons.mdiChevronDown }}</v-icon>
    </v-btn>
    <fingerprint-transactions-dialog
      v-model="showFingerprintDialog"
      :transaction-hash="tx.txhash"
    />
  </v-card>
</template>

<script>
import {
	mdiTransfer, mdiGraph, mdiFormatListNumbered, mdiCalendar,
	mdiCash, mdiFormatHeaderPound, mdiCircleMultipleOutline,
	mdiChevronDown, mdiChevronUp, mdiPickaxe, mdiFingerprint,
} from '@mdi/js';
import OutputItem from './OutputItem.vue';
import {shortenHash, convertAmount, isDestination} from '@/utilities';
import {ROUTE_NAME_HEURISTIC_PAGE, ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import IconItem from '../../common/IconItem.vue';
import FingerprintTransactionsDialog from './FingerprintTransactionsDialog.vue';
import {isProxy, toRaw} from 'vue';
import PrivacyChip from '@/components/common/PrivacyChip.vue';
import IconTitle from '@/components/common/IconTitle.vue';

export default {
	name: 'Transaction',
	components: {IconTitle, PrivacyChip, FingerprintTransactionsDialog, OutputItem, IconItem},
	props: {
		tx: {type: Object, required: true},
		showHeuristicEditorLink: {type: Boolean, required: true},
		showFingerprintLink: {type: Boolean, required: true},
		showDetails: {type: Boolean, required: false, default: false},
		showTitleLink: {type: Boolean, required: false, default: false},
		embed: {type: Boolean, required: false, default: false},
	},
	data() {
		return {
			icons: {
				mdiTransfer,
				mdiGraph,
				mdiFormatListNumbered,
				mdiCalendar,
				mdiCash,
				mdiFormatHeaderPound,
				mdiCircleMultipleOutline,
				mdiChevronDown,
				mdiChevronUp,
				mdiPickaxe,
				mdiFingerprint,
			},
			routes: {
				ROUTE_NAME_HEURISTIC_PAGE,
				ROUTE_NAME_BLOCK_PAGE,
				ROUTE_NAME_TRANSACTION_PAGE,
			},
			showTransactionDetails: this.showDetails,
			showAllOutputs: false,
			showFingerprintDialog: false,
			maxOutputs: 3,
		};
	},
	computed: {
		getInputs() {
			return this.getLimitedItems(this.sortByTimestamp(this.tx.inputs));
		},
		getResidualInputs() {
			return this.getResidualItems(this.sortByTimestamp(this.tx.inputs));
		},
		getOutputs() {
			return this.getLimitedItems(this.sortByTimestamp(this.tx.outputs));
		},
		getResidualOutputs() {
			return this.getResidualItems(this.sortByTimestamp(this.tx.outputs));
		},
		areItemsLimited() {
			if (!this.tx) {
				return false;
			}

			if (this.tx.inputs && this.tx.inputs.length > this.maxOutputs) {
				return true;
			}

			return Boolean(this.tx.outputs && this.tx.outputs.length > this.maxOutputs);
		},
	},
	methods: {
		shortenHash,
		convertAmount,
		isDestination,
		getLabel(count, label) {
			if (count > 1) {
				return `${count} ${label}s`;
			}

			return `${count} ${label}`;
		},
		sortByTimestamp(outputs) {
			if (!outputs) {
				return [];
			}

			let copiedOutputs;

			if (isProxy(outputs)) {
				copiedOutputs = structuredClone(toRaw(outputs));
			} else {
				copiedOutputs = structuredClone(outputs);
			}

			return copiedOutputs.sort((a, b) => {
				if (!a.ts || !b.ts) {
					return true;
				}

				return new Date(a.ts) - new Date(b.ts);
			});
		},
		isCoinBaseTx(tx) {
			if (!tx || !tx.outputs) {
				return false;
			}

			return !tx.inputs || tx.inputs.length === 0;
		},
		getLimitedItems(items) {
			if (!items) {
				return [];
			}

			return items.slice(0, this.maxOutputs);
		},
		getResidualItems(items) {
			if (!items) {
				return [];
			}

			if (items.length <= this.maxOutputs) {
				return [];
			}

			if (this.showAllOutputs) {
				return items.slice(this.maxOutputs);
			}

			return [];
		},
	},
};
</script>

<style scoped>

</style>
