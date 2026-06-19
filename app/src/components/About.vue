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
        <v-card variant="text">
          <v-card-text class="text-body-large">
            <div class="text-display-small">
              About
            </div>
            <p>
              Dakar is a blockchain forensics application, focusing on the analysis of CoinJoin transactions.
              It consists of a backend implemented in Go and web app written in Vue 3 with the Vuetify UI framework.
              <br>
              The backend ingests blockchain data via RPC connections to blockchain clients and performs transaction
              classification and address clustering on it.
              <br>
              The web app allows exploring the ingested and transformed data. Additionally, graph based editor enables
              viewing relationships between transactions and address clusters. Several CoinJoin analysis modules are
              available: heuristics, transaction similarity measure, mixing activity overview and more.
              <br>
              Dakar supports CoinJoins on the Bitcoin and Dash blockchain. Dakar's source code is freely available under the AGPLv3 license.
            </p>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import {onMounted} from 'vue';
import {PAGE_TITLE} from '@/constants';

onMounted(async () => {
	document.title = `About - ${PAGE_TITLE}`;
});

</script>

<style scoped>

</style>
