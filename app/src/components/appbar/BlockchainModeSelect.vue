<template>
  <v-menu v-if="modes.length > 0">
    <template #activator="{ props }">
      <v-btn
        v-bind="props"
        id="blockchain-mode"
        :class="$attrs.class"
        :color="blockchainModeIconColor"
        icon
      >
        <v-icon
          :icon="blockchainModeIcon"
          size="large"
        />
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
          <v-icon
            :icon="getBlockchainModeIcon(m)"
            :color="getBlockchainModeColor(m)"
          />
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

const modes = [BLOCKCHAIN_DASH, BLOCKCHAIN_BTC];

const settings = computed({
	get() {
		return localStore.getSettings;
	},
	set(value) {
		localStore.setSettings(value);
	},
});

const blockchainModeIcon = computed(() => getBlockchainModeIcon(settings.value.blockchainMode));
const blockchainModeIconColor = computed(() => getBlockchainModeColor(settings.value.blockchainMode));

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

function getBlockchainModeColor(mode) {
	switch (mode) {
		case BLOCKCHAIN_DASH: return '#008CE4';
		case BLOCKCHAIN_BTC: return '#FF9315';
		default: return 'black';
	}
}

</script>

<style scoped>

</style>
