<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div>
    <v-card
      class="mx-auto mb-0 mt-2"
      max-width="1200"
      variant="text"
    >
      <icon-title
        :title="`${title} Attributions`"
        :icon="mdiTag"
      />
      <v-tabs
        v-model="tab"
        align-tabs="start"
      >
        <v-tab>
          Private Attributions
        </v-tab>
        <v-tab>
          Search
        </v-tab>
      </v-tabs>
    </v-card>
    <v-window
      v-model="tab"
      class="mx-auto"
      style="max-width: 1200px"
      :touch="false"
    >
      <v-window-item>
        <attribution-overview
          :blockchain-mode="blockchainMode"
          :title="title"
        />
      </v-window-item>
      <v-window-item>
        <attribution-search
          :blockchain-mode="blockchainMode"
          :title="title"
        />
      </v-window-item>
    </v-window>
  </div>
</template>

<script setup>
import {mdiTag} from '@mdi/js';
import AttributionOverview from './AttributionOverview.vue';
import AttributionSearch from './AttributionSearch.vue';
import IconTitle from '@/components/common/IconTitle.vue';
import {ref} from 'vue';

defineProps({
	title: {type: String, required: true},
	blockchainMode: {type: String, required: true},
});

const tab = ref(null);

</script>

<style scoped>

</style>
