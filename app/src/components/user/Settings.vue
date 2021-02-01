<template>
  <v-container fluid>
    <v-row style="height: 80vh">
      <v-col cols="2" class="hidden-sm-and-down">
        <v-navigation-drawer permanent>
          <v-list-item>
            <v-list-item-icon>
              <v-icon>{{ icon.mdiCog }}</v-icon>
            </v-list-item-icon>
            <v-list-item-content>
              <v-list-item-title class="title">
                Settings
              </v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-divider></v-divider>
          <v-list dense nav>
            <v-list-item :to="{ name: profileRoute}">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiAccountDetails }}</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>
                  Profile
                </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item :to="{ name: miscRoute}">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiTune }}</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>
                  Misc
                </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-navigation-drawer>
      </v-col>
      <v-col>
        <router-view></router-view>
      </v-col>
    </v-row>

    <v-bottom-navigation class="hidden-md-and-up" fixed color="primary">
      <v-btn :to="{ name: profileRoute}">
        <span>Profile</span>
        <v-icon>{{ icon.mdiAccountDetails }}</v-icon>
      </v-btn>
      <v-btn :to="{ name: miscRoute}">
        <span>Misc</span>
        <v-icon>{{ icon.mdiTune }}</v-icon>
      </v-btn>
    </v-bottom-navigation>

  </v-container>
</template>

<script>
import {
  mdiAccountDetails, mdiTune, mdiCog,
} from '@mdi/js';
import { ROUTE_NAME_USER_PROFILE_PAGE, ROUTE_NAME_USER_MISC_PAGE } from '../../constants';

export default {
  name: 'Settings',
  data() {
    return {
      profileRoute: ROUTE_NAME_USER_PROFILE_PAGE,
      miscRoute: ROUTE_NAME_USER_MISC_PAGE,
      icon: {
        mdiAccountDetails, mdiTune, mdiCog,
      },
    };
  },
};
</script>

<style scoped>

</style>
