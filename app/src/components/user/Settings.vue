<template>
  <div class="fill-height" style="padding: 12px 10px 0 10px" v-if="this.userData">
    <v-row class="fill-height">
      <v-col cols="2" class="hidden-sm-and-down pa-0">
        <v-navigation-drawer permanent>
          <v-list-item>
            <v-list-item-icon>
              <v-icon>{{ icon.mdiCog }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title class="text-h6">
              Settings
            </v-list-item-title>
          </v-list-item>
          <v-divider></v-divider>
          <v-list dense nav>
            <v-list-item :to="{ name: profileRoute}">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiAccountDetails }}</v-icon>
              </v-list-item-icon>
              <v-list-item-title>
                Profile
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
    <v-bottom-navigation class="hidden-md-and-up" fixed color="primary">
      <v-btn :to="{ name: profileRoute}">
        <span>Profile</span>
        <v-icon>{{ icon.mdiAccountDetails }}</v-icon>
      </v-btn>
    </v-bottom-navigation>
  </div>
</template>

<script>
import {
  mdiAccountDetails, mdiCog,
} from '@mdi/js';
import { ROUTE_NAME_USER_PROFILE_PAGE } from '../../constants';

export default {
  name: 'Settings',
  data() {
    return {
      profileRoute: ROUTE_NAME_USER_PROFILE_PAGE,
      icon: {
        mdiAccountDetails, mdiCog,
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
