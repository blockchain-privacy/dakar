<template>
  <v-list>
    <v-list-item
      v-for="tx in getLimitedItems"
      :key="tx.txhash"
      :to="{ name: routes.ROUTE_NAME_TRANSACTION_PAGE,
             params: { id: tx.txhash }}"
    >
      <v-list-item-title>
        {{ tx.txhash }}
        <div v-if="tx.destinationCount">
          Destinations: {{ tx.destinationCount }}
        </div>
      </v-list-item-title>
    </v-list-item>
    <v-expand-transition>
      <div v-if="showAllOutputs">
        <v-list-item
          v-for="tx in getResidualItems"
          :key="tx.txhash"
          :to="{ name: routes.ROUTE_NAME_TRANSACTION_PAGE,
                 params: { id: tx.txhash }}"
        >
          <v-list-item-title>
            {{ tx.txhash }}
            <div v-if="tx.destinationCount">
              Destinations: {{ tx.destinationCount }}
            </div>
          </v-list-item-title>
        </v-list-item>
      </div>
    </v-expand-transition>
  </v-list>
  <v-btn
    v-if="areItemsLimited"
    variant="text"
    :rounded="false"
    :block="true"
    size="small"
    @click="showAllOutputs = !showAllOutputs"
  >
    {{ items.length - maxItems }} additional
    {{ plural('transaction', items.length - maxItems) }}
    <v-icon>{{ showAllOutputs ? icons.mdiChevronUp : icons.mdiChevronDown }}</v-icon>
  </v-btn>
</template>

<script>
import {ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import {mdiChevronDown, mdiChevronUp} from '@mdi/js';
import {plural} from '@/utilities';
import loginPage from '../user/LoginPage.vue';

export default {
	name: 'ResultItem',
	props: {
		items: {type: Array, required: true},
		maxItems: {type: Number, required: true},
	},
	data() {
		return {
			showAllOutputs: false,
			routes: {
				ROUTE_NAME_TRANSACTION_PAGE,
			},
			icons: {
				mdiChevronUp, mdiChevronDown,
			},
		};
	},
	computed: {
		loginPage() {
			return loginPage;
		},
		areItemsLimited() {
			return this.items.length > this.maxItems;
		},
		getLimitedItems() {
			if (!this.items) {
				return [];
			}

			return this.items.slice(0, this.maxItems);
		},
		getResidualItems() {
			if (!this.items) {
				return [];
			}

			if (this.items.length <= this.maxItems) {
				return [];
			}

			if (this.showAllOutputs) {
				return this.items.slice(this.maxItems);
			}

			return [];
		},
	},
	methods: {plural},
};
</script>

<style scoped>

</style>
