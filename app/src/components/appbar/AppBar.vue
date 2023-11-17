<template>
  <v-app-bar
    :absolute="true"
    :flat="minimize"
    :color="minimize?'transparent':null"
  >
    <v-spacer v-if="minimize" />
    <router-link
      v-if="!minimize"
      :to="{name: route.rootPage}"
      class="ms-2"
    >
      <v-img
        style="cursor:pointer"
        alt="Dakar Logo"
        class="shrink mr-2"
        src="../../assets/dakar_dash.svg"
        transition="fade-transition"
        width="32"
      />
    </router-link>
    <router-link
      v-if="!minimize"
      :to="{name: route.rootPage}"
      style="color: inherit; text-decoration: inherit"
    >
      <v-app-bar-title class="ms-2 d-none d-sm-flex">
        {{ applicationName }}
      </v-app-bar-title>
    </router-link>
    <query-input
      v-if="!minimize"
      class="mx-auto px-2"
      style="max-width: 600px"
    />
    <v-btn
      v-if="isPrivilegedOrHigher"
      icon
    >
      <v-icon>{{ icon.mdiDotsGrid }}</v-icon>
      <page-menu />
    </v-btn>
    <!-- todo: check if https://github.com/vuetifyjs/vuetify/issues/17234 is fixed (wrong menu position after window resize) -->
    <v-menu
      v-if="session"
      @update:model-value="triggerMenuResize"
    >
      <template #activator="{ props }">
        <v-btn
          icon
          v-bind="props"
        >
          <v-icon>{{ icon.mdiAccount }}</v-icon>
        </v-btn>
      </template>
      <v-list
        :nav="true"
        density="compact"
      >
        <v-list-item>
          <template #prepend>
            <v-icon :icon="icon.mdiAccountCircle" />
          </template>
          <v-list-item-title> {{ session.identity.traits.email }}</v-list-item-title>
        </v-list-item>
        <v-divider />
        <v-list-item :to="{name: route.userProfilePage}">
          <template #prepend>
            <v-icon :icon="icon.mdiCog" />
          </template>
          <v-list-item-title>Settings</v-list-item-title>
        </v-list-item>
        <v-list-item>
          <template #prepend>
            <v-icon :icon="icon.mdiThemeLightDark" />
          </template>
          <div class="d-flex">
            <v-list-item-title style="display:flex; align-items:center">
              Dark Mode
            </v-list-item-title>
            <dark-mode-switch class="mt-0 ml-2" />
          </div>
        </v-list-item>
        <v-list-item @click="initLogoutFlow">
          <template #prepend>
            <v-icon
              color="red"
              :icon="icon.mdiLogout"
            />
          </template>
          <v-list-item-title>Logout</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
    <v-btn
      v-if="!session"
      variant="flat"
      color="primary"
      :to="{ name: route.userLoginPage }"
    >
      <v-icon>{{ icon.mdiLogin }}</v-icon>
      Login
    </v-btn>
  </v-app-bar>
</template>

<script>
import {
	mdiAccount, mdiAccountCircle, mdiCog, mdiLogin, mdiLogout,
	mdiDotsGrid, mdiThemeLightDark,
} from '@mdi/js';
import PageMenu from './PageMenu.vue';
import QueryInput from './QueryInput.vue';
import DarkModeSwitch from './DarkModeSwitch.vue';
import {
	APPLICATION_NAME, ROUTE_NAME_ENTRY_PAGE, ROUTE_NAME_LOGIN_PAGE,
	ROUTE_NAME_USER_PROFILE_PAGE,
} from '@/constants';
import {isAdminIdentity, isPrivilegedIdentity} from '@/utilities';
import handleGetFlowError from '@/kratos';

export default {
	name: 'AppBar',
	components: {
		DarkModeSwitch,
		PageMenu,
		QueryInput,
	},
	props: {
		minimize: Boolean,
	},
	data() {
		return {
			applicationName: APPLICATION_NAME,
			icon: {
				mdiAccount,
				mdiLogin,
				mdiLogout,
				mdiAccountCircle,
				mdiCog,
				mdiDotsGrid,
				mdiThemeLightDark,
			},
			route: {
				userProfilePage: ROUTE_NAME_USER_PROFILE_PAGE,
				userLoginPage: ROUTE_NAME_LOGIN_PAGE,
				rootPage: ROUTE_NAME_ENTRY_PAGE,
			},
		};
	},
	computed: {
		session: {
			get() {
				return this.$store.getters.getSession;
			},
			set(value) {
				this.$store.dispatch('setSession', value);
			},
		},
		showUserAdmin() {
			return isAdminIdentity(this.session);
		},
		isPrivilegedOrHigher() {
			return isPrivilegedIdentity(this.session) || this.showUserAdmin;
		},
	},
	methods: {
		// Workaround for https://github.com/vuetifyjs/vuetify/issues/17234,
		// trigger resize a few times so the menu size is updated.
		async triggerMenuResize(val) {
			if (!val) {
				return;
			}

			for (let i = 0; i < 20; i++) {
				// eslint-disable-next-line no-await-in-loop
				await new Promise(resolve => {
					setTimeout(resolve, 10);
				});
				window.dispatchEvent(new Event('resize'));
			}
		},
		// GoToPage should receive a page name from ./constants
		goToPage(pageName) {
			// Only change route if not already on page
			if (this.$route.name !== pageName) {
				this.$router.push({name: pageName});
			}
		},
		async initLogoutFlow() {
			try {
				const response = await this.ory.frontend.createBrowserLogoutFlow();
				if (!response.data.logout_token) {
					return;
				}

				const logoutResponse = await this.ory.frontend.updateLogoutFlow({token: response.data.logout_token});

				if (logoutResponse.status === 204) {
					this.session = null;
					this.goToPage(this.route.rootPage);
				}
			} catch (e) {
				await handleGetFlowError(this, e, null);
			}
		},
	},
};
</script>

<style scoped>

</style>
