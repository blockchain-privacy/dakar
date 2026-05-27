<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<template>
  <v-row class="justify-center">
    <v-col
      cols="12"
      md="6"
    >
      <icon-title
        title="Updates"
        :icon="mdiBullhorn"
      />
      <v-card
        v-for="item in updates"
        :key="item.title"
        class="ma-2"
      >
        <v-card-text>
          <div class="d-flex align-center mb-3">
            <div class="text-title-large">
              {{ item.title }}
            </div>
            <div class="ms-auto text-body-large">
              {{ item.date.toLocaleDateString() }}
            </div>
          </div>
          <div>
            {{ item.subtitle }}
          </div>
        </v-card-text>
      </v-card>
    </v-col>
  </v-row>
</template>

<script setup>
import {onMounted} from 'vue';
import {mdiBullhorn} from '@mdi/js';
import {useLocalStore} from '@/pinia/local';
import {PAGE_TITLE, UPDATE_ID} from '@/constants/index.js';
import IconTitle from '@/components/common/IconTitle.vue';

const localStore = useLocalStore();

const updates = [
	{
		title: 'Vuetify 4',
		subtitle: 'Dakar uses the Vuetify UI framework. This update migrates the web application from Vuetify 3 to Vuetify 4. Users may notice slight typographic differences.',
		date: new Date(2026, 5 - 1, 27),
	},
	{
		title: 'Cluster Attributions',
		subtitle: 'The attribution feature has been removed. Sources for public attributions are rare, so this feature was hardly used. In workspaces, attributions can still be manually assigned.',
		date: new Date(2026, 5 - 1, 27),
	},
	{
		title: 'Address Exclusions',
		subtitle: 'The address exclusion feature has been removed. Due to the nature of CoinJoins, a very high number of exclusions is required for it to be useful. In practice this is unachievable.',
		date: new Date(2026, 5 - 1, 27),
	},
	{
		title: 'Model Context Protocol',
		subtitle: 'Dakar now includes experimental support for the Model Context Protocol (MCP).',
		date: new Date(2026, 5 - 1, 27),
	},
	{
		title: 'OAuth',
		subtitle: 'OAuth allows to easily use authenticated Dakar backend APIs. Contact us to get OAuth client details.',
		date: new Date(2026, 5 - 1, 27),
	},
	{
		title: 'Update Page',
		subtitle: 'Added this page. It is accessible via the profile menu.',
		date: new Date(2026, 5 - 1, 27),
	},
];

onMounted(() => {
	document.title = `Updates - ${PAGE_TITLE}`;
	localStore.setLastUpdateID(UPDATE_ID);
});

</script>

<style scoped>

</style>
