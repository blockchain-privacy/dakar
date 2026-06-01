<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-list min-width="270px">
    <v-list-item
      v-for="(item, index) in items"
      :key="index"
      @click="handleClick(item)"
    >
      <template #prepend>
        <v-icon
          :icon="BLOCKCHAIN_ATTRIBUTES[item.mode].icon"
          :color="BLOCKCHAIN_ATTRIBUTES[item.mode].color"
        />
      </template>
      <template #append>
        <v-chip
          v-if="getResultType(item.type)"
          class="ms-2"
        >
          {{ getResultType(item.type) }}
        </v-chip>
      </template>
      <div class="shorten">
        {{ item.title }}
      </div>
    </v-list-item>
  </v-list>
</template>

<script setup>
import {BLOCKCHAIN_ATTRIBUTES} from '@/constants/index.js';

defineProps({items: {type: Array, required: true}});
const emit = defineEmits(['itemClicked']);

// Functions
function getResultType(type) {
	switch (type) {
		case 'tx': {return 'Transaction';}

		case 'block': {return 'Block';}

		case 'addr': {return 'Address';}

		default: {
			return '';
		}
	}
}

function handleClick(item) {
	emit('itemClicked', item);
}
</script>

<style scoped>

</style>
