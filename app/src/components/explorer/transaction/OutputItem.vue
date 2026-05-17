<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-card
    :variant="highlight?'outlined':'text'"
    class="my-2"
    :ripple="false"
  >
    <v-card-text style="min-height: 90px">
      <v-row>
        <v-col style="min-height: 26px">
          <div class="d-flex justify-space-between align-center">
            <workspace-link
              v-if="addressHash"
              style="max-width: 350px"
              :to="{ name: ROUTE_NAME_ADDRESS_PAGE, params: { id: addressHash, blockchainMode: route.params.blockchainMode }}"
            >
              {{ addressHash }}
            </workspace-link>
            <div
              v-else
              style="min-width: 200px"
            >
              No address available
            </div>
            <div
              :class="{'text-no-wrap':true, 'ms-1':true, 'pt-1':true, 'amount':true, 'px-1':true,
                       'amountHighlighted':isWasabi2Amount}"
            >
              {{ convertAmount(amount) }} {{ coinUnit }}
            </div>
          </div>
        </v-col>
      </v-row>
      <v-row>
        <v-col
          v-if="txHash !== '' && (isInput || inputIndex >= 0)"
          style="min-height: 26px"
        >
          <div
            class="d-flex justify-space-between align-center"
            style="height: 100%"
          >
            <div
              class="text-body-small d-flex align-center text-no-wrap me-2 shorten"
              style="width: 100%"
            >
              <div class="flex-shrink-0">
                <workspace-link
                  v-tooltip="{'text': txHash, 'open-delay': 400,'location':'top'}"
                  :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: txHash, blockchainMode: route.params.blockchainMode }}"
                >
                  {{ isInput ? 'created' : 'spent' }}
                </workspace-link>
              </div>
              <span
                v-tooltip="{'text': new Date(timestamp).toLocaleString(), 'open-delay': 400, 'location': 'top'}"
                class="shorten"
              >on {{ timestamp ? new Date(timestamp).toLocaleString() : '' }}</span>
            </div>
            <privacy-chip
              v-if="transactionType"
              :transaction-type="transactionType"
              size="small"
            />
          </div>
        </v-col>
        <!-- set min-height so this col is as high as the other one -->
        <v-col
          v-else-if="!isInput"
          class="text-body-small d-flex align-center"
          style="min-height: 26px"
        >
          not spent
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script setup>
import {computed} from 'vue';
import {storeToRefs} from 'pinia';
import {useRoute} from 'vue-router';
import {convertAmount, getCoinUnit, isWasabi2Denomination} from '@/utilities';
import {
	ROUTE_NAME_ADDRESS_PAGE,
	ROUTE_NAME_TRANSACTION_PAGE,
} from '@/constants';
import PrivacyChip from '@/components/common/PrivacyChip.vue';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';
import {useExplorerStore} from '@/pinia/explorer.js';

const props = defineProps({
	isInput: {type: Boolean, required: true},
	addressHash: {type: String, required: true},
	amount: {type: Number, required: true},
	inputIndex: {type: Number, required: false, default: -1},
	outputIndex: {type: Number, required: false, default: -1},
	txHash: {type: String, required: false, default: ''},
	timestamp: {type: String, required: false, default: ''},
	transactionType: {type: String, required: false, default: ''},
	highlight: {type: Boolean, required: false},
});

const route = useRoute();
const {getHighlightWasabi2Denominations} = storeToRefs(useExplorerStore());

// Computed
const coinUnit = computed(() => getCoinUnit(route.params.blockchainMode));
const isWasabi2Amount = computed(() => getHighlightWasabi2Denominations.value && isWasabi2Denomination(props.amount));

</script>

<style scoped>

.amount {
  /* hide border, for proper sizing */
  border: 2px solid transparent;
  border-radius: 10px;
}
.amountHighlighted {
  border-color: #76C408;
}

</style>
