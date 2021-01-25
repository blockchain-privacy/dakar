<template>
  <v-card
      class="mx-auto elevation-12"
      max-width="700">
    <v-toolbar color="primary" dark flat>
      <v-toolbar-title>
        <v-icon>{{ icon.mdiAccountDetails }}</v-icon>
        Profile
      </v-toolbar-title>
    </v-toolbar>
    <ProfileItem v-for="(item, index) in listItems"
                 :key="index"
                 :title="item.title"
                 :icon="item.icon"
                 :item-value="item.val"
                 :action-function="item.actionFunction"/>
  </v-card>
</template>

<script>
import {
  mdiLock, mdiEmail, mdiCalendar, mdiCalendarEdit, mdiAccountDetails,
} from '@mdi/js';
import { PAGE_TITLE } from '../../constants';
import ProfileItem from './ProfileItem.vue';
import { doGet } from '../../utilities';

export default {
  name: 'Profile',
  components: { ProfileItem },
  data() {
    return {
      icon: {
        mdiLock, mdiEmail, mdiCalendar, mdiCalendarEdit, mdiAccountDetails,
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
    listItems() {
      return [
        {
          title: 'Email:',
          val: this.userData.email,
          icon: this.icon.mdiEmail,
          actionFunction: this.dummyFunc,
        },
        {
          title: 'Change Password:',
          val: '********',
          icon: this.icon.mdiLock,
          actionFunction: this.dummyFunc,
        },
        {
          title: 'Account last modified:',
          val: this.modifiedDate,
          icon: this.icon.mdiCalendarEdit,
        },
        {
          title: 'Account created:',
          val: this.createdDate,
          icon: this.icon.mdiCalendar,
        },
      ];
    },
  },
  methods: {
    dummyFunc() {
      doGet();
      console.log('test');
    },
  },
  mounted() {
    document.title = `Profile - ${PAGE_TITLE}`;
  },
};
</script>

<style scoped>

</style>
