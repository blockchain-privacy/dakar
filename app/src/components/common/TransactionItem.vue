<template>
  <v-card>
    <icon-title
      :title="`Transaction ${tx.txhash}`"
      :icon="icons.mdiTransfer"
      :to="{ name: route.tx, params: { id: tx.txhash }}"
    >
      <privacy-chip
        v-if="tx.privacytype"
        class="mx-3"
        :privacy-type="tx.privacytype"
      />
    </icon-title>
    <v-card-text>
      <v-list>
        <v-list-item :to="{ name: route.block, params: { id: tx.bhash }}">
          <v-list-item-title>
            Block Hash
          </v-list-item-title>
          <v-list-item-subtitle class="shorten">
            {{ tx.bhash }}
          </v-list-item-subtitle>
        </v-list-item>
        <v-list-item :to="{ name: route.block, params: { id: tx.bid }}">
          <v-list-item-title>
            Block ID
          </v-list-item-title>
          <v-list-item-subtitle>
            {{ tx.bid }}
          </v-list-item-subtitle>
        </v-list-item>
      </v-list>
    </v-card-text>
  </v-card>
</template>

<script>
import PrivacyChip from '@/components/common/PrivacyChip.vue';
import {ROUTE_NAME_BLOCK_PAGE, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import IconTitle from '@/components/common/IconTitle.vue';
import {mdiTransfer} from '@mdi/js';

export default {
	name: 'TransactionItem',
	components: {IconTitle, PrivacyChip},
	props: {
		tx: {type: Object, required: true},
	},
	data() {
		return {
			route: {
				block: ROUTE_NAME_BLOCK_PAGE,
				tx: ROUTE_NAME_TRANSACTION_PAGE,
			},
			icons: {
				mdiTransfer,
			},
		};
	},
};
</script>
<style scoped>

</style>
