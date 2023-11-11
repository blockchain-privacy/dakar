<template>
  <v-layout class="fill-height">
    <v-navigation-drawer>
      <v-list-item
        :prepend-icon="icon.mdiToolbox"
        title="Tools"
      />
      <v-divider />
      <v-list :nav="true">
        <v-list-item
          :to="{ name: heuristicsPage}"
          :prepend-icon="icon.mdiGraph"
          title="Heuristics"
        />
        <v-list-item
          :to="{ name: attributionsPage}"
          :prepend-icon="icon.mdiTag"
          title="Attributions"
        />
        <v-list-item
          :to="{ name: clusterOverviewPage}"
          :prepend-icon="icon.mdiMerge"
          title="Custom Clusters"
        />
        <v-list-item
          :to="{ name: addressExclusionPage}"
          :prepend-icon="icon.mdiPlaylistRemove"
          title="Address Exclusions"
        />
        <v-list-item
          :to="{ name: shortestPathPage}"
          :prepend-icon="icon.mdiChartTimelineVariant"
          title="Shortest Path"
        />
        <v-list-item
          :to="{ name: connectionLookupPage}"
          :prepend-icon="icon.mdiTextBoxSearch"
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
          <FadeTransition>
            <component
              :is="Component"
            />
          </FadeTransition>
        </router-view>
      </div>
    </v-main>
  </v-layout>
</template>

<script>
import {
	mdiGraph, mdiChartTimelineVariant, mdiToolbox, mdiTextBoxSearch, mdiMerge,
	mdiTag, mdiPlaylistRemove,
} from '@mdi/js';
import {
	ROUTE_NAME_USER_HEURISTIC_PAGE,
	ROUTE_NAME_SHORTEST_PATH_PAGE,
	ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
	ROUTE_NAME_CLUSTER_OVERVIEW,
	ROUTE_NAME_ATTRIBUTIONS, ROUTE_NAME_ADDRESS_EXCLUSIONS,
} from '@/constants';
import FadeTransition from '../common/FadeTransition.vue';

export default {
	name: 'ToolsPage',
	components: {FadeTransition},
	data() {
		return {
			heuristicsPage: ROUTE_NAME_USER_HEURISTIC_PAGE,
			shortestPathPage: ROUTE_NAME_SHORTEST_PATH_PAGE,
			connectionLookupPage: ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
			clusterOverviewPage: ROUTE_NAME_CLUSTER_OVERVIEW,
			attributionsPage: ROUTE_NAME_ATTRIBUTIONS,
			addressExclusionPage: ROUTE_NAME_ADDRESS_EXCLUSIONS,
			icon: {
				mdiGraph,
				mdiChartTimelineVariant,
				mdiToolbox,
				mdiTextBoxSearch,
				mdiMerge,
				mdiTag,
				mdiPlaylistRemove,
			},
		};
	},
};
</script>

<style scoped>

</style>
