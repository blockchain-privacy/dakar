<template>
  <v-card
      class="mx-auto elevation-12"
      max-width="700">
    <v-toolbar color="primary" dark flat>
      <v-toolbar-title>
        <v-icon>{{ icon.mdiTune }}</v-icon>
        Misc
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
  mdiInvertColors, mdiTune,
} from '@mdi/js';
import ProfileItem from './ProfileItem.vue';

export default {
  name: 'Misc',
  components: { ProfileItem },
  data() {
    return {
      icon: {
        mdiInvertColors, mdiTune,
      },
    };
  },
  computed: {
    listItems() {
      return [
        {
          title: 'Dark mode',
          val: 'enabled',
          icon: this.icon.mdiInvertColors,
        },
      ];
    },
  },
};
</script>

<style scoped>

</style>
