<template>
  <v-card class="elevation-4">
    <v-toolbar :color="tx.privacytype>=0?'purple':'primary'" dark flat>
      <v-toolbar-title>
        <v-icon>{{ icons.mdiTransfer }}</v-icon>
        Transaction
        <router-link v-if="showTitleLink" style="color: inherit;"
                     :to="{ name: routes.ROUTE_NAME_TRANSACTION_PAGE, params: { id: tx.txhash }}">
          {{ tx.txhash }}
        </router-link>
        <template v-else>
          {{ tx.txhash }}
        </template>
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
          id="btn_open_heuristic_editor"
          v-if="isDestination(tx.privacytype) && showHeuristicEditorLink"
          style="margin-right: 0" icon
          :to="{ name: routes.ROUTE_NAME_HEURISTIC_PAGE }">
        <v-icon>{{ icons.mdiGraph }}</v-icon>
      </v-btn>
      <v-tooltip bottom activator="#btn_open_heuristic_editor">
        <span>Open the heuristic editor for this transaction</span>
      </v-tooltip>
      <v-btn
          id="btn_find_similar_transactions"
          v-if="isDestination(tx.privacytype) && showFingerprintLink"
          style="margin-right: 0" icon
          @click="showFingerprintDialog = true">
        <v-icon>{{ icons.mdiFingerprint }}</v-icon>
      </v-btn>
      <v-tooltip bottom activator="#btn_find_similar_transactions">
        <span>Search for similar destination transactions</span>
      </v-tooltip>
    </v-toolbar>
    <v-card-text>
      <v-expand-transition>
        <div v-if="showTransactionDetails">
          <v-row>
            <v-col>
              <IconItem :icon="icons.mdiFormatListNumbered" title="Block Height">
                <router-link :to="{ name: routes.ROUTE_NAME_BLOCK_PAGE, params: { id: tx.bid }}">
                  {{ tx.bid }}
                </router-link>
              </IconItem>
            </v-col>
            <v-col>
              <IconItem :icon="icons.mdiCalendar" title="Timestamp">
                {{ new Date(tx.bts).toLocaleString() }}
              </IconItem>
            </v-col>
          </v-row>
          <v-row>
            <v-col v-if="(tx.fee || tx.fee === 0) && tx.fee >= 0">
              <IconItem :icon="icons.mdiCash" title="Fee">
                {{ convertAmount(tx.fee) }}
              </IconItem>
            </v-col>
            <v-col>
              <IconItem :icon="icons.mdiFormatHeaderPound" title="Block">
                <router-link :to="{ name: routes.ROUTE_NAME_BLOCK_PAGE, params: { id: tx.bhash }}">
                  {{ shortenHash(tx.bhash) }}
                </router-link>
              </IconItem>
            </v-col>
          </v-row>
          <v-row v-if="isCoinBaseTx(tx)">
            <v-col>
              <IconItem :icon="icons.mdiPickaxe" title="Coinbase">
                yes
              </IconItem>
            </v-col>
          </v-row>
          <v-row v-if="getPrivacyTypeLabel(tx.privacytype)">
            <v-col>
              <IconItem :icon="icons.mdiIncognito" title="Privacy Type">
                <WikiTooltip :description-url="getPrivacyTypeTooltip(tx.privacytype)">
                  {{ getPrivacyTypeLabel(tx.privacytype) }}
                </WikiTooltip>
              </IconItem>
            </v-col>
            <v-col v-if="isMixing(tx.privacytype)">
              <IconItem :icon="icons.mdiCircleMultipleOutline" title="Mixing Denomination">
                {{ getMixingLabel(tx.privacytype) }}
              </IconItem>
            </v-col>
          </v-row>
          <!-- bottom spacer for transition -->
          <div style="height: 10px">
          </div>
        </div>
      </v-expand-transition>
      <v-btn text plain block x-small
             @click="showTransactionDetails = !showTransactionDetails"
             style="margin-top:-16px;">
        <v-icon>{{ showTransactionDetails ? icons.mdiChevronUp : icons.mdiChevronDown }}</v-icon>
      </v-btn>
      <v-row>
        <v-col v-if="tx.inputs">
          <p class="ml-2">{{ getLabel(tx.inputs.length, 'Input') }}</p>
          <OutputItem
              v-for="i in getInputs"
              v-bind:key="i.addresshash + i.inputindex"
              :is-input="true"
              :amount="i.amount"
              :address-hash="i.addresshash"
              :tx-hash="i.txhash"
              :sig-asm="i.sigasm"
              :key-asm="i.keyasm"
              :output-index="i.outputindex"
              :input-index="i.inputindex"
              :timestamp="i.ts"
              :privacy-type="i.privacytype"/>
          <!-- split in two for nicer transition -->
          <v-expand-transition>
            <div v-if="showAllOutputs">
              <OutputItem
                  v-for="i in getResidualInputs"
                  v-bind:key="i.addresshash + i.inputindex"
                  :is-input="true"
                  :amount="i.amount"
                  :address-hash="i.addresshash"
                  :tx-hash="i.txhash"
                  :sig-asm="i.sigasm"
                  :key-asm="i.keyasm"
                  :output-index="i.outputindex"
                  :input-index="i.inputindex"
                  :timestamp="i.ts"
                  :privacy-type="i.privacytype"/>
            </div>
          </v-expand-transition>
        </v-col>
        <!-- empty col if no inputs exist -->
        <v-col v-else></v-col>
        <v-col v-if="tx.outputs">
          <p class="ml-2">{{ getLabel(tx.outputs.length, 'Output') }}</p>
          <OutputItem v-for="i in getOutputs"
                      v-bind:key="i.addresshash + i.outputindex"
                      :is-input="false"
                      :amount="i.amount"
                      :address-hash="i.addresshash"
                      :tx-hash="i.txhash"
                      :sig-asm="i.sigasm"
                      :key-asm="i.keyasm"
                      :output-index="i.outputindex"
                      :input-index="i.inputindex"
                      :timestamp="i.ts"
                      :privacy-type="i.privacytype"/>
          <!-- split in two for nicer transition -->
          <v-expand-transition>
            <div v-if="showAllOutputs">
              <OutputItem v-for="i in getResidualOutputs"
                          v-bind:key="i.addresshash + i.outputindex"
                          :is-input="false"
                          :amount="i.amount"
                          :address-hash="i.addresshash"
                          :tx-hash="i.txhash"
                          :sig-asm="i.sigasm"
                          :key-asm="i.keyasm"
                          :output-index="i.outputindex"
                          :input-index="i.inputindex"
                          :timestamp="i.ts"
                          :privacy-type="i.privacytype"/>
            </div>
          </v-expand-transition>
        </v-col>
        <!-- empty col if no outputs exist -->
        <v-col v-else></v-col>
      </v-row>
    </v-card-text>
    <v-btn text plain block x-small v-if="areItemsLimited"
           @click="showAllOutputs = !showAllOutputs">
      <v-icon>{{ showAllOutputs ? icons.mdiChevronUp : icons.mdiChevronDown }}</v-icon>
    </v-btn>
    <fingerprint-transactions v-model="showFingerprintDialog" :transaction-hash="tx.txhash"/>
  </v-card>
</template>

<script>
import {
  mdiTransfer, mdiGraph, mdiFormatListNumbered, mdiCalendar,
  mdiCash, mdiFormatHeaderPound, mdiIncognito, mdiCircleMultipleOutline,
  mdiChevronDown, mdiChevronUp, mdiPickaxe, mdiFingerprint,
} from '@mdi/js';
import OutputItem from './OutputItem.vue';
import {
  shortenHash, convertAmount, isDestination, getPrivacyTypeLabel, isMixing,
  getMixingLabel, getPrivacyTypeTooltip,
} from '../../utilities';
import { ROUTE_NAME_HEURISTIC_PAGE, ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_TRANSACTION_PAGE } from '../../constants';
import IconItem from '../common/IconItem.vue';
import FingerprintTransactions from '../dialogs/FingerprintTransactions.vue';
import WikiTooltip from '../wiki/WikiTooltip.vue';

export default {
  name: 'Transaction',
  components: {
    WikiTooltip, FingerprintTransactions, OutputItem, IconItem,
  },
  props: {
    tx: { type: Object, required: true },
    showHeuristicEditorLink: { type: Boolean, required: true },
    showFingerprintLink: { type: Boolean, required: true },
    showDetails: { type: Boolean, required: false, default: false },
    showTitleLink: { type: Boolean, required: false, default: false },
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
        mdiIncognito,
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
      if (!this.tx) return false;
      if (this.tx.inputs && this.tx.inputs.length > this.maxOutputs) return true;
      return !!(this.tx.outputs && this.tx.outputs.length > this.maxOutputs);
    },
  },
  methods: {
    shortenHash,
    convertAmount,
    isDestination,
    getPrivacyTypeLabel,
    getPrivacyTypeTooltip,
    isMixing,
    getMixingLabel,
    getLabel(count, label) {
      if (count > 1) {
        return `${count} ${label}s`;
      }
      return `${count} ${label}`;
    },
    sortByTimestamp(outputs) {
      if (outputs == null) return [];
      const copiedOutputs = JSON.parse(JSON.stringify(outputs));

      return copiedOutputs.sort((a, b) => {
        if (!a.ts || !b.ts) return true;
        return new Date(a.ts) > new Date(b.ts);
      });
    },
    isCoinBaseTx(tx) {
      if (!tx || !tx.outputs) {
        return false;
      }
      return !tx.inputs || tx.inputs.length === 0;
    },
    getLimitedItems(items) {
      if (!items) return [];
      return items.slice(0, this.maxOutputs);
    },
    getResidualItems(items) {
      if (!items) return [];
      if (items.length <= this.maxOutputs) return [];
      if (this.showAllOutputs) return items.slice(this.maxOutputs);
      return [];
    },
  },
};
</script>

<style scoped>

</style>
