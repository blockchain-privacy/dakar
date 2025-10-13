<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div>
    <v-navigation-drawer absolute>
      <v-list-item
        :prepend-icon="mdiToolbox"
        title="Tools"
      />
      <v-divider />
      <v-list nav>
        <v-list-item
          :to="{ name: ROUTE_NAME_WORKSPACES_PAGE}"
          prepend-icon="$graphIcon"
          title="Workspaces"
        />
        <v-list-item
          :to="{ name: ROUTE_NAME_ATTRIBUTIONS}"
          :prepend-icon="mdiTag"
          title="Attributions"
        />
        <v-list-item
          :to="{ name: ROUTE_NAME_CLUSTER_OVERVIEW}"
          :prepend-icon="mdiMerge"
          title="Custom Clusters"
        />
        <v-list-item
          :to="{ name: ROUTE_NAME_ADDRESS_EXCLUSIONS}"
          :prepend-icon="mdiPlaylistRemove"
          title="Address Exclusions"
        />
      </v-list>
    </v-navigation-drawer>
    <div class="mx-2 mx-lg-16 pb-10">
      <router-view v-slot="{ Component }">
        <fade-transition>
          <component :is="Component" />
        </fade-transition>
      </router-view>
    </div>
  </div>
</template>

<script setup>
import {
	mdiMerge, mdiPlaylistRemove, mdiTag, mdiToolbox,
} from '@mdi/js';
import {
	ROUTE_NAME_ADDRESS_EXCLUSIONS,
	ROUTE_NAME_ATTRIBUTIONS,
	ROUTE_NAME_CLUSTER_OVERVIEW,
	ROUTE_NAME_WORKSPACES_PAGE,
} from '@/constants';
import FadeTransition from '../common/FadeTransition.vue';

</script>

<style scoped>

</style>
