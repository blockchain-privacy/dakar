<template>
  <v-card
    :variant="highlight?'outlined':'text'"
    class="my-2"
    :ripple="false"
  >
    <v-card-text style="min-height: 90px">
      <v-row>
        <v-col>
          <div class="d-flex justify-space-between">
            <store-link
              v-if="addressHash"
              :to="{ name: ROUTE_NAME_ADDRESS_PAGE, params: { id: addressHash }}"
              class="shorten"
            >
              {{ addressHash }}
            </store-link>
            <div class="text-no-wrap ms-2">
              {{ convertAmount(amount) }} {{ COIN_UNIT }}
            </div>
          </div>
        </v-col>
      </v-row>
      <v-row>
        <v-col v-if="isInput || inputIndex >= 0">
          <div
            class="d-flex justify-space-between align-center"
          >
            <div class="text-caption d-flex align-center">
              <store-link :to="{ name: ROUTE_NAME_TRANSACTION_PAGE, params: { id: txHash }}">
                <span>{{ isInput ? 'created' : 'spent' }}</span>
              </store-link>
              on {{ timestamp ? new Date(timestamp).toLocaleString() : '' }}
            </div>
            <privacy-chip
              v-if="privacyType"
              :privacy-type="privacyType"
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
              :readonly="true"
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
              :readonly="true"
              :model-value="scriptToAscii(keyAsm)"
            />
            <v-text-field
              v-if="sigAsm"
              hide-details
              density="compact"
              label="Signature script"
              variant="outlined"
              :readonly="true"
              :model-value="sigAsm"
            />
          </v-col>
        </v-row>
      </v-expand-transition>
    </v-card-text>
    <v-btn
      v-if="keyAsm || sigAsm"
      variant="text"
      :block="true"
      size="x-small"
      @click="expanded = !expanded"
    >
      <v-icon>{{ expanded ? mdiChevronUp : mdiChevronDown }}</v-icon>
    </v-btn>
  </v-card>
</template>

<script setup>
import {mdiChevronUp, mdiChevronDown, mdiFormatColorText} from '@mdi/js';
import {convertAmount} from '@/utilities';
import {COIN_UNIT, ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import PrivacyChip from '@/components/common/PrivacyChip.vue';
import {ref} from 'vue';
import StoreLink from '@/components/common/StoreLink.vue';

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
	privacyType: {type: Number, required: false, default: -1},
	highlight: {type: Boolean, required: false, default: false},
});

const expanded = ref(false);
const showAscii = ref(false);

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
