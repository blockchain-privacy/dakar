<template>
  <v-container :fluid="true">
    <v-row
      v-if="data"
      align="center"
      justify="center"
    >
      <!-- duplicate transaction hashes can exist -> loop through all results
      (e.g. d5d27987d2a3dfc724e359870c6644b40e497bdc0589a033220fe15429d88599 in Bitcoin) -->
      <v-col
        v-for="tx in data"
        :key="tx.txhash+tx.bid"
        cols="12"
        sm="12"
        md="12"
        lg="10"
        xl="8"
      >
        <transaction
          :tx="tx"
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
import {computed, onMounted, onUpdated, watch} from 'vue';
import {useStore} from 'vuex';

const store = useStore();

// Computed
const data = computed(() => store.getters.getTransactionData);
const session = computed(() => store.getters.getSession);
const isPrivilegedOrHigher = computed(() => isPrivilegedIdentity(session.value) || isAdminIdentity(session.value));

// Watchers
watch(data, () => {
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
	if (data.value && data.value[0].txhash) {
		h = ` ${data.value[0].txhash} `;
	}

	document.title = `Transaction${h}- ${PAGE_TITLE}`;
}
</script>
