<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-menu activator="parent">
    <v-card
      class="pa-3"
      min-width="250px"
      max-width="350px"
    >
      <div>
        <v-row density="compact">
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
              title="Custom Clusters"
              :icon="mdiMerge"
              :color="iconColor.default"
              :to="{ name: ROUTE_NAME_CLUSTER_OVERVIEW}"
            />
          </v-col>
        </v-row>
        <v-row density="compact">
          <v-col>
            <link-card
              title="Server Status"
              :icon="mdiServer"
              :color="iconColor.default"
              :to="{ name: ROUTE_NAME_STATUS_PAGE}"
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
        </v-row>
      </div>
      <template v-if="showUserAdmin">
        <v-divider class="my-2" />
        <v-row density="compact">
          <v-col>
            <link-card
              title="User Admin"
              :icon="mdiAccountSupervisor"
              :color="iconColor.admin"
              :to="{ name: ROUTE_NAME_USER_ADMIN_PAGE }"
            />
          </v-col>
        </v-row>
      </template>
    </v-card>
  </v-menu>
</template>

<script setup>
import {
	mdiAccountSupervisor,
	mdiServer,
	mdiMerge,
	mdiBookOpen,
} from '@mdi/js';
import {computed} from 'vue';
import {storeToRefs} from 'pinia';
import LinkCard from '../common/LinkCard.vue';
import {
	ROUTE_NAME_USER_ADMIN_PAGE,
	ROUTE_NAME_STATUS_PAGE,
	ROUTE_NAME_CLUSTER_OVERVIEW,
	ROUTE_NAME_WIKI_ROOT,
	ROUTE_NAME_WORKSPACES_PAGE,
} from '@/constants';
import {isAnyAdminIdentity} from '@/utilities';
import {useLocalStore} from '@/pinia/local';

const {session} = storeToRefs(useLocalStore());

const iconColor = {default: 'primary', admin: 'red darken-3'};

// Computed
const showUserAdmin = computed(() => isAnyAdminIdentity(session.value));

</script>

<style scoped>

</style>
