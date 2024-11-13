<template>
  <v-card
    :variant="highlight?'outlined':'text'"
    class="my-2"
    :ripple="false"
  >
    <v-card-text
      style="min-height: 90px"
    >
      <v-row>
        <v-col>
          <div class="d-flex justify-space-between align-center">
            <workspace-link
              v-if="addressHash"
              :to="{ name: ROUTE_NAME_ADDRESS_PAGE, params: { id: addressHash, blockchainMode: getSettings.blockchainMode }}"
            >
              {{ addressHash }}
            </workspace-link>
            <div
              v-else
              style="min-width: 200px"
            >
              No Address available
            </div>
            <div class="text-no-wrap ms-1">
              {{ convertAmount(amount) }} {{ coinUnit }}
            </div>
          </div>
        </v-col>
      </v-row>
      <v-row>
        <v-col v-if="isInput || inputIndex >= 0">
          <div
            class="d-flex justify-space-between align-center"
          >
            <div class="text-caption d-flex align-center text-no-wrap me-2">
              <workspace-link :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: txHash, blockchainMode: getSettings.blockchainMode }}">
                {{ isInput ? 'created' : 'spent' }}
              </workspace-link>
              on {{ timestamp ? new Date(timestamp).toLocaleString() : '' }}
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
          class="text-caption py-1"
          style="min-height: 50px"
        >
          not spent
        </v-col>
      </v-row>
      <v-expand-transition>
        <v-row v-if="expanded">
          <v-col>
            <v-text-field
              v-if="keyAsm"
              hide-details
              density="compact"
              label="Key script"
              class="mb-3"
              variant="outlined"
              readonly
              :model-value="keyAsm"
            >
              <template #append>
                <v-tooltip
                  v-if="keyAsm"
                  location="bottom"
                  text="Toggle ASCII encoding of key script"
                >
                  <template #activator="{props}">
                    <v-btn
                      v-bind="props"
                      variant="text"
                      icon
                      @click="showAscii = !showAscii"
                    >
                      <v-icon>{{ mdiFormatColorText }}</v-icon>
                    </v-btn>
                  </template>
                </v-tooltip>
              </template>
            </v-text-field>
            <v-text-field
              v-if="keyAsm && showAscii && scriptToAscii(keyAsm)"
              hide-details
              density="compact"
              label="Key script"
              class="mb-3"
              variant="outlined"
              readonly
              :model-value="scriptToAscii(keyAsm)"
            />
            <v-text-field
              v-if="sigAsm"
              hide-details
              density="compact"
              label="Signature script"
              variant="outlined"
              readonly
              :model-value="sigAsm"
            />
          </v-col>
        </v-row>
      </v-expand-transition>
    </v-card-text>
    <v-btn
      v-if="keyAsm || sigAsm"
      variant="text"
      block
      size="x-small"
      @click="expanded = !expanded"
    >
      <v-icon>{{ expanded ? mdiChevronUp : mdiChevronDown }}</v-icon>
    </v-btn>
  </v-card>
</template>

<script setup>
import {mdiChevronUp, mdiChevronDown, mdiFormatColorText} from '@mdi/js';
import {convertAmount, getCoinUnit} from '@/utilities';
import {
	ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_TRANSACTION_PAGE,
} from '@/constants';
import PrivacyChip from '@/components/common/PrivacyChip.vue';
import {computed, ref} from 'vue';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';

defineProps({
	isInput: {type: Boolean, required: true},
	addressHash: {type: String, required: true},
	amount: {type: Number, required: true},
	keyAsm: {type: String, required: false, default: ''},
	sigAsm: {type: String, required: false, default: ''},
	inputIndex: {type: Number, required: false, default: -1},
	outputIndex: {type: Number, required: false, default: -1},
	txHash: {type: String, required: false, default: ''},
	timestamp: {type: String, required: false, default: ''},
	transactionType: {type: String, required: false, default: ''},
	highlight: {type: Boolean, required: false},
});

const {getSettings} = storeToRefs(useLocalStore());

const expanded = ref(false);
const showAscii = ref(false);

// Computed
const coinUnit = computed(() => getCoinUnit(getSettings.value.blockchainMode));

// Functions
const isHex = str => /^[A-F\d]+$/i.test(str);

function hex2Ascii(hex) {
	const hexString = hex.toString();// Force conversion
	let str = '';
	for (let i = 0; i < hexString.length; i += 2) {
		str += String.fromCharCode(parseInt(hexString.substring(i, i + 2), 16));
	}

	return str;
}

function scriptToAscii(script) {
	const hex = script.split(' ').find(d => isHex(d));

	if (hex === undefined) {
		return '';
	}

	return hex2Ascii(hex);
}

</script>

<style scoped>

</style>
