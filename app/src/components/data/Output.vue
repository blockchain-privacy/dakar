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
                          outlined readonly :value="keyAsm"/>
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
import { mdiChevronUp, mdiChevronDown } from '@mdi/js';
import { convertAmount, getPrivacyTypeLabel } from '../../utilities';
import { COIN_UNIT, ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_TRANSACTION_PAGE } from '../../constants';

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
        mdiChevronUp, mdiChevronDown,
      },
      COIN_UNIT,
      addressRoute: ROUTE_NAME_ADDRESS_PAGE,
      txRoute: ROUTE_NAME_TRANSACTION_PAGE,
      expanded: false,
    };
  },
  methods: {
    convertAmount,
    getPrivacyTypeLabel,
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
