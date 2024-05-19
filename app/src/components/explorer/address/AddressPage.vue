<template>
  <v-container fluid>
    <v-row
      align="center"
      justify="center"
    >
      <v-col
        cols="12"
        sm="12"
        md="12"
        lg="9"
        xl="8"
      >
        <fade-transition>
          <address-view
            v-if="addressData"
            :address-data="addressData"
          />
          <v-skeleton-loader
            v-else
            class="mx-auto"
            type="list-item-three-line, list-item-three-line, list-item-three-line"
          />
        </fade-transition>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import {PAGE_TITLE} from '@/constants';
import {onMounted, watch} from 'vue';
import {useExplorerStore} from '@/pinia/explorer';
import {storeToRefs} from 'pinia';
import AddressView from '@/components/explorer/address/Address.vue';
import FadeTransition from '@/components/common/FadeTransition.vue';
const {address: addressData} = storeToRefs(useExplorerStore());

// Watchers
watch(addressData, () => {
	setInitialState();
});

// Hooks
onMounted(() => {
	setInitialState();
});

// Functions
function setInitialState() {
	let h = '';

	// Detect if address hash has changed
	if (addressData.value && addressData.value.addresshash) {
		h = `${addressData.value.addresshash} `;
	}

	document.title = `Address ${h}- ${PAGE_TITLE}`;
}

</script>

<style scoped>

</style>
