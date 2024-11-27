<template>
  <v-menu v-if="modes.length > 0">
    <template #activator="{ props }">
      <v-btn
        v-bind="props"
        id="blockchain-mode"
        icon
      >
        <v-icon>{{ blockchainModeIcon }}</v-icon>
      </v-btn>
    </template>
    <v-list
      nav
      density="compact"
    >
      <v-list-item
        v-for="m in modes"
        :key="m"
        :active="settings.blockchainMode === m"
        :to="{name: ROUTE_NAME_ENTRY_PAGE, params: {blockchainMode: m}}"
      >
        <template #prepend>
          <v-icon :icon="getBlockchainModeIcon(m)" />
        </template>
        <v-list-item-title>{{ getBlockchainModeName(m) }}</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script setup>
import {BLOCKCHAIN_BTC, BLOCKCHAIN_DASH, ROUTE_NAME_ENTRY_PAGE} from '@/constants/index.js';
import {computed} from 'vue';
import {bitcoinLogo, dashLogo} from '@/customIcons/index.js';
import {mdiCog} from '@mdi/js';
import {useLocalStore} from '@/pinia/local.js';
const localStore = useLocalStore();
defineProps({modes: {type: Array, required: true}});

const settings = computed({
	get() {
		return localStore.getSettings;
	},
	set(value) {
		localStore.setSettings(value);
	},
});

const blockchainModeIcon = computed(() => getBlockchainModeIcon(settings.value.blockchainMode));

// Functions
function getBlockchainModeIcon(mode) {
	switch (mode) {
		case BLOCKCHAIN_DASH: return dashLogo;
		case BLOCKCHAIN_BTC: return bitcoinLogo;
		default: return mdiCog;
	}
}

function getBlockchainModeName(mode) {
	switch (mode) {
		case BLOCKCHAIN_DASH: return 'Dash';
		case BLOCKCHAIN_BTC: return 'Bitcoin';
		default: return 'Unknown';
	}
}

</script>

<style scoped>

</style>
