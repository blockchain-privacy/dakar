<template>
  <v-menu activator="parent">
    <v-card
      class="pa-3"
      min-width="250px"
      max-width="350px"
    >
      <div v-if="showTools">
        <v-row :no-gutters="true">
          <v-col>
            <link-card
              title="Workspaces"
              icon="$graphIcon"
              :color="iconColor.default"
              :to="{ name: ROUTE_NAME_WORKSPACES_PAGE }"
            />
          </v-col>
          <v-col>
            <link-card
              title="Attributions"
              :icon="mdiTag"
              :color="iconColor.default"
              :to="{ name: ROUTE_NAME_ATTRIBUTIONS }"
            />
          </v-col>
          <v-col>
            <link-card
              title="Custom Clusters"
              :icon="mdiMerge"
              :color="iconColor.default"
              :to="{ name: ROUTE_NAME_CLUSTER_OVERVIEW }"
            />
          </v-col>
        </v-row>
        <v-row :no-gutters="true">
          <v-col>
            <link-card
              title="Address Exclusions"
              :icon="mdiPlaylistRemove"
              :color="iconColor.default"
              :to="{ name: ROUTE_NAME_ADDRESS_EXCLUSIONS }"
            />
          </v-col>
          <v-col>
            <link-card
              title="Shortest Path"
              :icon="mdiChartTimelineVariant"
              :color="iconColor.default"
              :to="{ name: ROUTE_NAME_SHORTEST_PATH_PAGE }"
            />
          </v-col>
          <v-col>
            <link-card
              title="Connection Lookup"
              :icon="mdiTextBoxSearch"
              :color="iconColor.default"
              :to="{ name: ROUTE_NAME_CONNECTION_LOOKUP_PAGE }"
            />
          </v-col>
        </v-row>
      </div>
      <v-divider class="my-2" />
      <v-row :no-gutters="true">
        <v-col>
          <link-card
            title="Server Status"
            :icon="mdiServer"
            :color="iconColor.default"
            :to="{ name: ROUTE_NAME_STATUS_PAGE }"
          />
        </v-col>
        <v-col>
          <link-card
            title="Wiki"
            :icon="mdiBookOpen"
            :color="iconColor.default"
            :to="{ name: ROUTE_NAME_WIKI_ROOT }"
          />
        </v-col>
        <v-col v-if="showUserAdmin">
          <link-card
            title="User Admin"
            :icon="mdiAccountSupervisor"
            :color="iconColor.admin"
            :to="{ name: ROUTE_NAME_USER_ADMIN_PAGE }"
          />
        </v-col>
      </v-row>
    </v-card>
  </v-menu>
</template>

<script setup>
import {
	mdiChartTimelineVariant, mdiTextBoxSearch, mdiAccountSupervisor, mdiServer,
	mdiMerge, mdiTag, mdiPlaylistRemove, mdiBookOpen,
} from '@mdi/js';
import {
	ROUTE_NAME_SHORTEST_PATH_PAGE, ROUTE_NAME_USER_ADMIN_PAGE, ROUTE_NAME_CONNECTION_LOOKUP_PAGE,
	ROUTE_NAME_STATUS_PAGE, ROUTE_NAME_CLUSTER_OVERVIEW,
	ROUTE_NAME_ATTRIBUTIONS, ROUTE_NAME_ADDRESS_EXCLUSIONS, ROUTE_NAME_WIKI_ROOT, ROUTE_NAME_WORKSPACES_PAGE,
} from '@/constants';
import LinkCard from '../common/LinkCard.vue';
import {isAdminIdentity, isPrivilegedIdentity} from '@/utilities';
import {computed} from 'vue';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local';

const {session} = storeToRefs(useLocalStore());

const iconColor = {default: 'primary', admin: 'red darken-3'};

// Computed
const showUserAdmin = computed(() => isAdminIdentity(session.value));
const showTools = computed(() => isPrivilegedIdentity(session.value) || showUserAdmin);

</script>

<style scoped>

</style>
