<template>
  <v-menu activator="parent">
    <v-card
      class="pa-3"
      min-width="250px"
      max-width="350px"
    >
      <div v-if="showTools">
        <v-row no-gutters>
          <v-col>
            <LinkCard
              title="Heuristics"
              :icon="icons.mdiGraph"
              :color="iconColor.default"
              :to="{ name: routes.heuristicsPage }"
            />
          </v-col>
          <v-col>
            <LinkCard
              title="Attributions"
              :icon="icons.mdiTag"
              :color="iconColor.default"
              :to="{ name: routes.attributionsPage }"
            />
          </v-col>
          <v-col>
            <LinkCard
              title="Custom Clusters"
              :icon="icons.mdiMerge"
              :color="iconColor.default"
              :to="{ name: routes.clusterOverviewPage }"
            />
          </v-col>
        </v-row>
        <v-row no-gutters>
          <v-col>
            <LinkCard
              title="Address Exclusions"
              :icon="icons.mdiPlaylistRemove"
              :color="iconColor.default"
              :to="{ name: routes.addressExclusionPage }"
            />
          </v-col>
          <v-col>
            <LinkCard
              title="Shortest Path"
              :icon="icons.mdiChartTimelineVariant"
              :color="iconColor.default"
              :to="{ name: routes.shortestPathPage }"
            />
          </v-col>
          <v-col>
            <LinkCard
              title="Connection Lookup"
              :icon="icons.mdiTextBoxSearch"
              :color="iconColor.default"
              :to="{ name: routes.connectionLookupPage }"
            />
          </v-col>
        </v-row>
      </div>
      <v-divider class="my-2" />
      <v-row no-gutters>
        <v-col>
          <LinkCard
            title="Server Status"
            :icon="icons.mdiServer"
            :color="iconColor.default"
            :to="{ name: routes.serverStatusPage }"
          />
        </v-col>
        <v-col>
          <LinkCard
            title="Wiki"
            :icon="icons.mdiBookOpen"
            :color="iconColor.default"
            :to="{ name: routes.wikiRootPage }"
          />
        </v-col>
        <v-col v-if="showUserAdmin">
          <LinkCard
            title="User Admin"
            :icon="icons.mdiAccountSupervisor"
            :color="iconColor.admin"
            :to="{ name: routes.userAdminPage }"
          />
        </v-col>
      </v-row>
    </v-card>
  </v-menu>
</template>

<script>
import {
	mdiAccount, mdiGraph, mdiChartTimelineVariant, mdiTextBoxSearch, mdiAccountSupervisor, mdiServer,
	mdiMerge, mdiTag, mdiPlaylistRemove, mdiBookOpen,
} from '@mdi/js';
import {
	ROUTE_NAME_SHORTEST_PATH_PAGE, ROUTE_NAME_USER_ADMIN_PAGE, ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
	ROUTE_NAME_USER_HEURISTIC_PAGE, ROUTE_NAME_STATUS_PAGE, ROUTE_NAME_CLUSTER_OVERVIEW,
	ROUTE_NAME_ATTRIBUTIONS, ROUTE_NAME_ADDRESS_EXCLUSIONS, ROUTE_NAME_WIKI_ROOT,
} from '@/constants';
import LinkCard from '../common/LinkCard.vue';
import {isAdminIdentity, isPrivilegedIdentity} from '@/utilities';

export default {
	name: 'PageMenu',
	components: {LinkCard},
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
				mdiTag,
				mdiPlaylistRemove,
				mdiBookOpen,
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
				clusterOverviewPage: ROUTE_NAME_CLUSTER_OVERVIEW,
				attributionsPage: ROUTE_NAME_ATTRIBUTIONS,
				addressExclusionPage: ROUTE_NAME_ADDRESS_EXCLUSIONS,
				wikiRootPage: ROUTE_NAME_WIKI_ROOT,
			},
		};
	},
	computed: {
		session() {
			return this.$store.getters.getSession;
		},
		showUserAdmin() {
			return isAdminIdentity(this.session);
		},
		showTools() {
			return isPrivilegedIdentity(this.session) || this.showUserAdmin;
		},
	},
};
</script>

<style scoped>

</style>
