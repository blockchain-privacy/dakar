<template>
  <v-card
      class="mx-auto"
      max-width="600"
      tile>
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
  mdiLock, mdiEmail, mdiPencil, mdiCalendar, mdiCalendarEdit,
} from '@mdi/js';
import ProfileItem from './ProfileItem.vue';

export default {
  name: 'Profile',
  components: { ProfileItem },
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
          icon: this.icon.mdiCalendarEdit,
        },
      ];
    },
  },
  methods: {
    dummyFunc() {
      console.log('test');
    },
  },
};
</script>

<style scoped>

</style>
