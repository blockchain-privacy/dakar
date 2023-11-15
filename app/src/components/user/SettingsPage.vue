<template>
  <div
    class="fill-height"
    style="padding: 12px 10px 0 10px"
  >
    <v-row class="fill-height">
      <v-col class="fill-height">
        <router-view v-slot="{ Component }">
          <FadeTransition>
            <component :is="Component" />
          </FadeTransition>
        </router-view>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import {
	mdiAccountDetails, mdiCog,
} from '@mdi/js';
import {ROUTE_NAME_USER_PROFILE_PAGE} from '@/constants';
import FadeTransition from '../common/FadeTransition.vue';

export default {
	name: 'SettingsPage',
	components: {FadeTransition},
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
