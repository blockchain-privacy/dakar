<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

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
        lg="10"
        xl="8"
      >
        <v-card>
          <v-toolbar flat>
            <v-toolbar-title>{{ pageTitle }}</v-toolbar-title>
          </v-toolbar>
          <v-card-text>
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
import {PAGE_TITLE} from '@/constants';
import {onMounted, ref} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const props = defineProps({
	pageTitle: {type: String, required: true},
	url: {type: String, required: true},
});

const loadedHTML = ref('');
const route = useRoute();
const msgStore = useMsgStore();

function setErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: true, category: route.name,
	});
}

onMounted(async () => {
	document.title = `${props.pageTitle} - ${PAGE_TITLE}`;

	try {
		const response = await fetch(props.url);
		loadedHTML.value = await response.text();
	} catch (_) {
		setErrorMessage('Unable to load data, try again later');
	}
});

</script>

<style scoped>

</style>
