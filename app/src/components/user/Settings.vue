<template>
  <div class="fill-height" style="padding: 12px 10px 0 10px">
    <v-row class="fill-height">
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
};
</script>

<style scoped>

</style>
