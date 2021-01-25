<template>
  <v-card
      class="mx-auto"
      max-width="600"
      tile>
    <v-list-item>
      <v-list-item-icon class="button-sized-icon">
        <v-icon>
          {{ icon.mdiEmail }}
        </v-icon>
      </v-list-item-icon>
      <v-list-item-content>
        <v-row>
          <v-col>
            <v-list-item-title>E-Mail:</v-list-item-title>
          </v-col>
          <v-col>
            <v-list-item-title>{{ this.userData.email }}</v-list-item-title>
          </v-col>
        </v-row>
      </v-list-item-content>
      <v-list-item-icon>
        <v-btn icon>
          <v-icon>
            {{ icon.mdiPencil }}
          </v-icon>
        </v-btn>
      </v-list-item-icon>
    </v-list-item>
    <v-list-item>
      <v-list-item-icon class="button-sized-icon">
        <v-icon>
          {{ icon.mdiLock }}
        </v-icon>
      </v-list-item-icon>
      <v-list-item-content>
        <v-row>
          <v-col>
            <v-list-item-title>Change Password:</v-list-item-title>
          </v-col>
          <v-col>
            <v-list-item-title>***********</v-list-item-title>
          </v-col>
        </v-row>
      </v-list-item-content>
      <v-list-item-icon>
        <v-btn icon>
          <v-icon>
            {{ icon.mdiPencil }}
          </v-icon>
        </v-btn>
      </v-list-item-icon>
    </v-list-item>
    <v-list-item>
      <v-list-item-icon class="button-sized-icon">
        <v-icon>
          {{ icon.mdiCalendarEdit }}
        </v-icon>
      </v-list-item-icon>
      <v-list-item-content>
        <v-row>
          <v-col>
            <v-list-item-title>Account last modified:</v-list-item-title>
          </v-col>
          <v-col>
            <v-list-item-title>{{ this.modifiedDate }}</v-list-item-title>
          </v-col>
        </v-row>
      </v-list-item-content>
      <v-list-item-icon class="button-sized-icon"/>
    </v-list-item>
    <v-list-item>
      <v-list-item-icon class="button-sized-icon">
        <v-icon>
          {{ icon.mdiCalendar }}
        </v-icon>
      </v-list-item-icon>
      <v-list-item-content>
        <v-row>
          <v-col>
            <v-list-item-title>Account created:</v-list-item-title>
          </v-col>
          <v-col>
            <v-list-item-title>{{ this.createdDate }}</v-list-item-title>
          </v-col>
        </v-row>
      </v-list-item-content>
      <v-list-item-icon class="button-sized-icon"/>
    </v-list-item>
  </v-card>
</template>

<script>
import {
  mdiLock, mdiEmail, mdiPencil, mdiCalendar, mdiCalendarEdit,
} from '@mdi/js';

export default {
  name: 'Profile',
  data() {
    return {
      icon: {
        mdiLock, mdiEmail, mdiPencil, mdiCalendar, mdiCalendarEdit,
      },
    };
  },
  computed: {
    userData: {
      get() {
        return this.$store.getters.getActiveUser;
      },
      set(value) {
        this.$store.dispatch('setActiveUser', value);
      },
    },
    modifiedDate() {
      return new Date(this.userData.modified).toLocaleString();
    },
    createdDate() {
      return new Date(this.userData.created).toLocaleString();
    },
  },
};
</script>

<style scoped>
/* Used for centering */
.button-sized-icon {
  height: 36px;
  width: 36px;
}
</style>
