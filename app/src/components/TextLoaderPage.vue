<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-container fluid>
    <v-row class="align-center justify-center">
      <v-col
        md="12"
        xl="8"
      >
        <v-card>
          <v-card-title flat>
            {{ pageTitle }}
          </v-card-title>
          <v-card-text>
            <alert :text="errorMsg" />
            <!-- html is loaded from safe source -->
            <!-- eslint-disable vue/no-v-html -->
            <div
              v-if="loadedHTML"
              v-html="loadedHTML"
            />
            <v-skeleton-loader
              v-else
              type="article@3"
            />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import {onMounted, ref} from 'vue';
import {PAGE_TITLE} from '@/constants';
import Alert from '@/components/common/Alert.vue';

const props = defineProps({
	pageTitle: {type: String, required: true},
	url: {type: String, required: true},
});

const loadedHTML = ref('');
const errorMsg = ref('');

onMounted(async () => {
	document.title = `${props.pageTitle} - ${PAGE_TITLE}`;
	errorMsg.value = '';
	try {
		const response = await fetch(props.url);
		loadedHTML.value = await response.text();
	} catch {
		errorMsg.value = 'Unable to load data, try again later';
	}
});

</script>

<style scoped>

</style>
