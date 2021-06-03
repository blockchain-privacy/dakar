<template>
  <v-menu offset-y v-model="inputVal">
    <template v-slot:activator="{ on, attrs }">
      <slot v-bind="attrs" v-on="on"></slot>
    </template>
    <v-card class="pa-3">
      <v-row no-gutters v-if="showTools">
        <v-col>
          <LinkCard
              title="Shortest Path"
              :icon="icons.mdiChartTimelineVariant"
              :to="{ name: routes.shortestPathPage }"/>
        </v-col>
        <v-col>
          <LinkCard
              title="Heuristics"
              :icon="icons.mdiGraph"
              :to="{ name: routes.heuristicsPage }"/>
        </v-col>
        <v-col>
          <LinkCard
              title="Connection Lookup"
              :icon="icons.mdiTextBoxSearch"
              :to="{ name: routes.connectionLookupPage }"/>
        </v-col>
      </v-row>
      <v-divider class="my-2"/>
      <v-row no-gutters >
        <v-col>
          <LinkCard
              title="Server Status"
              :icon="icons.mdiServer"
              :to="{ name: routes.serverStatusPage }"/>
        </v-col>
        <v-col v-if="showUserAdmin">
          <LinkCard
              title="User Admin"
              :icon="icons.mdiAccountSupervisor"
              :to="{ name: routes.userAdminPage }"/>
        </v-col>
      </v-row>
    </v-card>
  </v-menu>
</template>

<script>
import {
  mdiAccount, mdiGraph, mdiChartTimelineVariant, mdiTextBoxSearch, mdiAccountSupervisor, mdiServer,
} from '@mdi/js';
import {
  ROUTE_NAME_SHORTEST_PATH_PAGE, ROUTE_NAME_USER_ADMIN_PAGE, ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
  ROUTE_NAME_USER_HEURISTIC_PAGE, ROUTE_NAME_STATUS_PAGE,
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
      },
      routes: {
        userAdminPage: ROUTE_NAME_USER_ADMIN_PAGE,
        shortestPathPage: ROUTE_NAME_SHORTEST_PATH_PAGE,
        heuristicsPage: ROUTE_NAME_USER_HEURISTIC_PAGE,
        connectionLookupPage: ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
        serverStatusPage: ROUTE_NAME_STATUS_PAGE,
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
