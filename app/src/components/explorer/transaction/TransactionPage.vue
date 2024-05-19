<template>
  <v-container fluid>
    <v-row
      v-if="tx"
      align="center"
      justify="center"
    >
      <!-- duplicate transaction hashes can exist -> loop through all results
      (e.g. d5d27987d2a3dfc724e359870c6644b40e497bdc0589a033220fe15429d88599 in Bitcoin) -->
      <v-col
        v-for="t in tx"
        :key="t.txhash+t.bid"
        cols="12"
        sm="12"
        md="12"
        lg="10"
        xl="8"
      >
        <transaction
          :tx="t"
          :show-heuristic-editor-link="isPrivilegedOrHigher"
          :show-fingerprint-link="isPrivilegedOrHigher"
          show-details
        />
      </v-col>
    </v-row>
    <v-skeleton-loader
      v-else
      class="mx-auto"
      type="list-item-three-line, list-item-three-line, list-item-three-line"
    />
  </v-container>
</template>

<script setup>
import Transaction from './Transaction.vue';
import {PAGE_TITLE} from '@/constants';
import {isAdminIdentity, isPrivilegedIdentity} from '@/utilities';
import {
	computed, onMounted, onUpdated, watch,
} from 'vue';
import {storeToRefs} from 'pinia';
import {useExplorerStore} from '@/pinia/explorer';
import {useLocalStore} from '@/pinia/local';

const {transaction: tx} = storeToRefs(useExplorerStore());
const {session} = storeToRefs(useLocalStore());

// Computed
const isPrivilegedOrHigher = computed(() => isPrivilegedIdentity(session.value) || isAdminIdentity(session.value));

// Watchers
watch(tx, () => {
	setPageTitle();
});

// Hooks
onMounted(() => {
	setPageTitle();
});
onUpdated(() => {
	setPageTitle();
});

// Functions
function setPageTitle() {
	let h = ' ';
	if (tx.value && tx.value[0].txhash) {
		h = ` ${tx.value[0].txhash} `;
	}

	document.title = `Transaction${h}- ${PAGE_TITLE}`;
}
</script>
