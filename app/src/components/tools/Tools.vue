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
            <v-list-item :to="{ name: clusterLookupPage}">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiMerge }}</v-icon>
              </v-list-item-icon>
              <v-list-item-title>
                Cluster Lookup
              </v-list-item-title>
            </v-list-item>
            <v-list-item :to="{ name: addClusterPage}">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiMerge }}</v-icon>
              </v-list-item-icon>
              <v-list-item-title>
                Add Cluster
              </v-list-item-title>
            </v-list-item>
            <v-list-item :to="{ name: mixingActivityPage}">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiChartBar }}</v-icon>
              </v-list-item-icon>
              <v-list-item-title>
                Mixing Activity
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
    <v-bottom-navigation shift class="hidden-lg-and-up" fixed color="primary">
      <v-btn :to="{ name: shortestPathPage}">
        <span>Shortest Path</span>
        <v-icon>{{ icon.mdiChartTimelineVariant }}</v-icon>
      </v-btn>
      <v-btn :to="{ name: heuristicsPage}">
        <span>Heuristics</span>
        <v-icon>{{ icon.mdiGraph }}</v-icon>
      </v-btn>
      <v-btn :to="{ name: connectionLookupPage}">
        <span>Connection Lookup</span>
        <v-icon>{{ icon.mdiTextBoxSearch }}</v-icon>
      </v-btn>
      <v-btn :to="{ name: clusterLookupPage}">
        <span>Cluster Lookup</span>
        <v-icon>{{ icon.mdiMerge }}</v-icon>
      </v-btn>
      <v-btn :to="{ name: mixingActivityPage}">
        <span>Mixing Activity</span>
        <v-icon>{{ icon.mdiChartBar }}</v-icon>
      </v-btn>
    </v-bottom-navigation>
  </div>
</template>

<script>
import {
  mdiGraph, mdiChartTimelineVariant, mdiToolbox, mdiTextBoxSearch, mdiMerge,
  mdiChartBar,
} from '@mdi/js';
import {
  ROUTE_NAME_USER_HEURISTIC_PAGE,
  ROUTE_NAME_SHORTEST_PATH_PAGE,
  ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
  ROUTE_NAME_CLUSTER_LOOKUP_PAGE,
  ROUTE_NAME_MIXING_ACTIVITY,
  ROUTE_NAME_ADD_CLUSTER_PAGE,
} from '../../constants';

export default {
  name: 'Tools',
  data() {
    return {
      heuristicsPage: ROUTE_NAME_USER_HEURISTIC_PAGE,
      shortestPathPage: ROUTE_NAME_SHORTEST_PATH_PAGE,
      connectionLookupPage: ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
      clusterLookupPage: ROUTE_NAME_CLUSTER_LOOKUP_PAGE,
      addClusterPage: ROUTE_NAME_ADD_CLUSTER_PAGE,
      mixingActivityPage: ROUTE_NAME_MIXING_ACTIVITY,
      icon: {
        mdiGraph,
        mdiChartTimelineVariant,
        mdiToolbox,
        mdiTextBoxSearch,
        mdiMerge,
        mdiChartBar,
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
