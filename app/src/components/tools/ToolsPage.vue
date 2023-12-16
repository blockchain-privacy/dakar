<template>
  <v-layout class="fill-height">
    <v-navigation-drawer>
      <v-list-item
        :prepend-icon="mdiToolbox"
        title="Tools"
      />
      <v-divider />
      <v-list :nav="true">
        <v-list-item
          :to="{ name: ROUTE_NAME_USER_HEURISTIC_PAGE}"
          :prepend-icon="mdiGraph"
          title="Heuristics"
        />
        <v-list-item
          :to="{ name: ROUTE_NAME_WORKSPACES_PAGE}"
          :prepend-icon="mdiGraph"
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
        <v-list-item
          :to="{ name: ROUTE_NAME_SHORTEST_PATH_PAGE}"
          :prepend-icon="mdiChartTimelineVariant"
          title="Shortest Path"
        />
        <v-list-item
          :to="{ name: ROUTE_NAME_CONNECTION_LOOKUP_PAGE}"
          :prepend-icon="mdiTextBoxSearch"
          title="Connection Lookup"
        />
      </v-list>
    </v-navigation-drawer>
    <v-main>
      <div
        class="mx-2 mx-lg-16"
        style="padding: 12px 10px 50px 10px"
      >
        <router-view v-slot="{ Component }">
          <fade-transition>
            <component
              :is="Component"
            />
          </fade-transition>
        </router-view>
      </div>
    </v-main>
  </v-layout>
</template>

<script setup>
import {
	mdiGraph, mdiChartTimelineVariant, mdiToolbox, mdiTextBoxSearch, mdiMerge,
	mdiTag, mdiPlaylistRemove,
} from '@mdi/js';
import {
	ROUTE_NAME_USER_HEURISTIC_PAGE,
	ROUTE_NAME_SHORTEST_PATH_PAGE,
	ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
	ROUTE_NAME_CLUSTER_OVERVIEW,
	ROUTE_NAME_ATTRIBUTIONS,
	ROUTE_NAME_ADDRESS_EXCLUSIONS,
	ROUTE_NAME_WORKSPACES_PAGE,
} from '@/constants';
import FadeTransition from '../common/FadeTransition.vue';

</script>

<style scoped>

</style>
