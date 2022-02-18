<template>
  <div class="fill-height" style="padding: 12px 10px 0 10px" v-if="this.userData">
    <v-row class="fill-height">
      <v-col cols="2" class="hidden-md-and-down pa-0">
        <v-navigation-drawer permanent>
          <v-list-item>
            <v-list-item-icon>
              <v-icon>{{ icon.mdiToolbox }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title class="text-h6">
              Tools
            </v-list-item-title>
          </v-list-item>
          <v-divider></v-divider>
          <v-list dense nav>
            <v-list-item :to="{ name: shortestPathPage}">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiChartTimelineVariant }}</v-icon>
              </v-list-item-icon>
              <v-list-item-title>
                Shortest Path
              </v-list-item-title>
            </v-list-item>
            <v-list-item :to="{ name: heuristicsPage}">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiGraph }}</v-icon>
              </v-list-item-icon>
              <v-list-item-title>
                Heuristics
              </v-list-item-title>
            </v-list-item>
            <v-list-item :to="{ name: connectionLookupPage}">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiTextBoxSearch }}</v-icon>
              </v-list-item-icon>
              <v-list-item-title>
                Connection Lookup
              </v-list-item-title>
            </v-list-item>
            <v-list-item :to="{ name: clusterOverviewPage}">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiMerge }}</v-icon>
              </v-list-item-icon>
              <v-list-item-title>
                Cluster Overview
              </v-list-item-title>
            </v-list-item>
            <v-list-item :to="{ name: attributionsPage}">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiTag }}</v-icon>
              </v-list-item-icon>
              <v-list-item-title>
                Attributions
              </v-list-item-title>
            </v-list-item>
          </v-list>
        </v-navigation-drawer>
      </v-col>
      <v-col class="fill-height">
        <transition name="component-fade" mode="out-in">
          <router-view></router-view>
        </transition>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import {
  mdiGraph, mdiChartTimelineVariant, mdiToolbox, mdiTextBoxSearch, mdiMerge,
  mdiTag,
} from '@mdi/js';
import {
  ROUTE_NAME_USER_HEURISTIC_PAGE,
  ROUTE_NAME_SHORTEST_PATH_PAGE,
  ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
  ROUTE_NAME_CLUSTER_OVERVIEW,
  ROUTE_NAME_ATTRIBUTIONS,
} from '../../constants';

export default {
  name: 'Tools',
  data() {
    return {
      heuristicsPage: ROUTE_NAME_USER_HEURISTIC_PAGE,
      shortestPathPage: ROUTE_NAME_SHORTEST_PATH_PAGE,
      connectionLookupPage: ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
      clusterOverviewPage: ROUTE_NAME_CLUSTER_OVERVIEW,
      attributionsPage: ROUTE_NAME_ATTRIBUTIONS,
      icon: {
        mdiGraph,
        mdiChartTimelineVariant,
        mdiToolbox,
        mdiTextBoxSearch,
        mdiMerge,
        mdiTag,
      },
    };
  },
  computed: {
    userData() {
      return this.$store.getters.getActiveUser;
    },
  },
};
</script>

<style scoped>

</style>
