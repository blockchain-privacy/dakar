<template>
  <v-card
    variant="text"
    class="my-2"
    :ripple="false"
  >
    <v-card-text style="min-height: 90px">
      <v-row>
        <v-col>
          <div class="d-flex justify-space-between">
            <router-link
              v-if="addressHash"
              :to="{ name: addressRoute, params: { id: addressHash }}"
              class="shorten"
            >
              {{ addressHash }}
            </router-link>
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
            <div class="text-caption">
              <router-link :to="{ name: txRoute, params: { id: txHash }}">
                <span>{{ isInput ? 'created' : 'spent' }}</span>
              </router-link>
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
                      <v-icon>{{ icons.mdiFormatColorText }}</v-icon>
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
      <v-icon>{{ expanded ? icons.mdiChevronUp : icons.mdiChevronDown }}</v-icon>
    </v-btn>
  </v-card>
</template>

<script>
import {mdiChevronUp, mdiChevronDown, mdiFormatColorText} from '@mdi/js';
import {convertAmount} from '@/utilities';
import {COIN_UNIT, ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import PrivacyChip from '@/components/common/PrivacyChip.vue';

const isHex = str => /^[A-F\d]+$/i.test(str);

function hex2Ascii(hex) {
	const hexString = hex.toString();// Force conversion
	let str = '';
	for (let i = 0; i < hexString.length; i += 2) {
		str += String.fromCharCode(parseInt(hexString.substring(i, i + 2), 16));
	}

	return str;
}

export default {
	name: 'OutputItem',
	components: {PrivacyChip},
	props: {
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
	},
	data() {
		return {
			icons: {
				mdiChevronUp, mdiChevronDown, mdiFormatColorText,
			},
			COIN_UNIT,
			addressRoute: ROUTE_NAME_ADDRESS_PAGE,
			txRoute: ROUTE_NAME_TRANSACTION_PAGE,
			expanded: false,
			showAscii: false,
		};
	},
	methods: {
		convertAmount,
		scriptToAscii(script) {
			const hex = script.split(' ').find(d => isHex(d));

			if (hex === undefined) {
				return '';
			}

			return hex2Ascii(hex);
		},
	},
};
</script>

<style scoped>
.shorten {
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}
</style>
