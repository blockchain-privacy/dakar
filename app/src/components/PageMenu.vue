<template>
  <v-menu offset-y v-model="inputVal" style="z-index: 99">
    <template v-slot:activator="{ on, attrs }">
      <slot v-bind="attrs" v-on="on"></slot>
    </template>
    <v-card class="pa-3" min-width="250px">
      <v-row no-gutters v-if="showTools">
        <v-col>
          <LinkCard
              title="Shortest Path"
              :icon="icons.mdiChartTimelineVariant"
              :color="iconColor.default"
              :to="{ name: routes.shortestPathPage }"/>
        </v-col>
        <v-col>
          <LinkCard
              title="Heuristics"
              :icon="icons.mdiGraph"
              :color="iconColor.default"
              :to="{ name: routes.heuristicsPage }"/>
        </v-col>
        <v-col>
          <LinkCard
              title="Connection Lookup"
              :icon="icons.mdiTextBoxSearch"
              :color="iconColor.default"
              :to="{ name: routes.connectionLookupPage }"/>
        </v-col>
        <v-col>
          <LinkCard
              title="Cluster Lookup"
              :icon="icons.mdiMerge"
              :color="iconColor.default"
              :to="{ name: routes.clusterLookupPage }"/>
        </v-col>
      </v-row>
      <v-divider class="my-2"/>
      <v-row no-gutters >
        <v-col>
          <LinkCard
              title="Server Status"
              :icon="icons.mdiServer"
              :color="iconColor.default"
              :to="{ name: routes.serverStatusPage }"/>
        </v-col>
        <v-col v-if="showUserAdmin">
          <LinkCard
              title="User Admin"
              :icon="icons.mdiAccountSupervisor"
              :color="iconColor.admin"
              :to="{ name: routes.userAdminPage }"/>
        </v-col>
      </v-row>
    </v-card>
  </v-menu>
</template>

<script>
import {
  mdiAccount, mdiGraph, mdiChartTimelineVariant, mdiTextBoxSearch, mdiAccountSupervisor, mdiServer,
  mdiMerge,
} from '@mdi/js';
import {
  ROUTE_NAME_SHORTEST_PATH_PAGE, ROUTE_NAME_USER_ADMIN_PAGE, ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
  ROUTE_NAME_USER_HEURISTIC_PAGE, ROUTE_NAME_STATUS_PAGE, ROUTE_NAME_CLUSTER_LOOKUP_PAGE,
} from '../constants';
import LinkCard from './common/LinkCard.vue';
import { isAdminUser, isPrivilegedUser } from '../utilities';

export default {
  name: 'PageMenu',
  components: { LinkCard },
  props: {
    value: { type: Boolean, required: true },
  },
  data() {
    return {
      icons: {
        mdiAccount,
        mdiGraph,
        mdiChartTimelineVariant,
        mdiTextBoxSearch,
        mdiAccountSupervisor,
        mdiServer,
        mdiMerge,
      },
      iconColor: {
        default: 'primary',
        admin: 'red darken-3',
      },
      routes: {
        userAdminPage: ROUTE_NAME_USER_ADMIN_PAGE,
        shortestPathPage: ROUTE_NAME_SHORTEST_PATH_PAGE,
        heuristicsPage: ROUTE_NAME_USER_HEURISTIC_PAGE,
        connectionLookupPage: ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
        serverStatusPage: ROUTE_NAME_STATUS_PAGE,
        clusterLookupPage: ROUTE_NAME_CLUSTER_LOOKUP_PAGE,
      },
    };
  },
  computed: {
    inputVal: {
      get() {
        return this.value;
      },
      set(val) {
        this.$emit('input', val);
      },
    },
    userData: {
      get() {
        return this.$store.getters.getActiveUser;
      },
      set(value) {
        this.$store.dispatch('setActiveUser', value);
      },
    },
    showUserAdmin() {
      return isAdminUser(this.userData);
    },
    showTools() {
      return isPrivilegedUser(this.userData) || isAdminUser(this.userData);
    },
  },
};
</script>

<style scoped>

</style>
