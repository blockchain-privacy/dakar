<template>
  <v-card outlined class="my-2"
          :ripple="false">
    <v-card-text>
      <v-row>
        <v-col cols="auto" class="shorten">
          <router-link :to="{ name: addressRoute, params: { id: addressHash }}" v-if="addressHash">
            {{ addressHash }}
          </router-link>
        </v-col>
        <v-col cols="auto" class="ml-auto">
          {{ convertAmount(amount) }} {{ COIN_UNIT }}
        </v-col>
      </v-row>
      <v-row class="my-0">
        <v-col class="text-caption py-1" v-if="isInput || inputIndex >= 0">
          <div class="d-flex justify-space-between">
            <div>
              <router-link :to="{ name: txRoute, params: { id: txHash }}">
                <span>{{ isInput ? 'created' : 'spent' }}</span>
              </router-link>
              on {{ timestamp ? new Date(timestamp).toLocaleString() : '' }}
            </div>
            <v-chip outlined label color="purple" small v-if="privacyType">
              {{ getPrivacyTypeLabel(privacyType) }}
            </v-chip>
          </div>
        </v-col>
        <v-col class="text-caption py-1" v-else-if="!isInput">
          not spent
        </v-col>
      </v-row>
      <v-expand-transition>
        <v-row v-if="expanded">
          <v-col>
            <v-text-field v-if="keyAsm" hide-details dense label="Key script" class="mb-3"
                          outlined readonly :value="keyAsm"
                          :append-outer-icon="icons.mdiFormatColorText"
                          @click:append-outer="showAscii = !showAscii">
              <template v-slot:append-outer>
                <v-btn
                    id="btn_toggle_ascii"
                    style="top:-7px"
                    v-if="keyAsm" icon
                    @click="showAscii = !showAscii">
                  <v-icon>{{ icons.mdiFormatColorText }}</v-icon>
                </v-btn>
                <v-tooltip bottom activator="#btn_toggle_ascii">
                  <span>Show script encoded in ASCII</span>
                </v-tooltip>
              </template>
            </v-text-field>
            <v-text-field v-if="keyAsm && showAscii &&  scriptToAscii(keyAsm)" hide-details
                          dense label="Key script" class="mb-3"
                          outlined readonly :value="scriptToAscii(keyAsm)"/>
            <v-text-field v-if="sigAsm" hide-details dense label="Signature script"
                          outlined readonly :value="sigAsm"/>
          </v-col>
        </v-row>
      </v-expand-transition>
    </v-card-text>
    <v-btn text plain block x-small @click="expanded = !expanded" v-if="keyAsm || sigAsm">
      <v-icon>{{ expanded ? icons.mdiChevronUp : icons.mdiChevronDown }}</v-icon>
    </v-btn>
  </v-card>
</template>

<script>
import { mdiChevronUp, mdiChevronDown, mdiFormatColorText } from '@mdi/js';
import { convertAmount, getPrivacyTypeLabel } from '../../utilities';
import { COIN_UNIT, ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_TRANSACTION_PAGE } from '../../constants';

const isHex = (str) => /^[A-F0-9]+$/i.test(str);

function hex2Ascii(hex) {
  const hexString = hex.toString();// force conversion
  let str = '';
  for (let i = 0; i < hexString.length; i += 2) {
    str += String.fromCharCode(parseInt(hexString.substr(i, 2), 16));
  }
  return str;
}

export default {
  name: 'Output',
  props: {
    isInput: { type: Boolean, required: true },
    addressHash: { type: String, required: true },
    amount: { type: Number, required: true },
    keyAsm: { type: String, required: false },
    sigAsm: { type: String, required: false },
    inputIndex: { type: Number, required: false },
    outputIndex: { type: Number, required: false },
    txHash: { type: String, required: false },
    timestamp: { type: String, required: false },
    privacyType: { type: Number, required: false },
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
    getPrivacyTypeLabel,
    scriptToAscii(script) {
      const splits = script.split(' ');

      const hex = splits.find((d) => isHex(d));

      if (hex === undefined) return '';

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
