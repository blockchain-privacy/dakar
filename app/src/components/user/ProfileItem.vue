<template>
  <v-list-item>
    <v-list-item-icon class="button-sized-icon">
      <v-icon>
        {{ icon }}
      </v-icon>
    </v-list-item-icon>
    <v-list-item-content>
      <v-row>
        <v-col>
          <v-list-item-title>{{ title }}</v-list-item-title>
        </v-col>
        <v-col>
          <v-list-item-title>{{ itemValue }}</v-list-item-title>
        </v-col>
      </v-row>
    </v-list-item-content>
    <v-list-item-icon class="button-sized-icon" v-if="!actionFunction"/>
    <v-switch v-if="actionFunction && isBoolean" @change="actionFunction"
              v-model="switchEnabled"/>
    <v-list-item-icon v-if="actionFunction && !isBoolean">
      <v-btn icon @click="actionFunction">
        <v-icon>
          {{ editIcon }}
        </v-icon>
      </v-btn>
    </v-list-item-icon>
  </v-list-item>
</template>

<script>
import { mdiPencil } from '@mdi/js';

export default {
  name: 'ProfileItem',
  data() {
    return {
      editIcon: mdiPencil,
      switchEnabled: false,
    };
  },
  props: {
    icon: { type: String, required: true },
    title: { type: String, required: true },
    itemValue: { type: String, required: false },
    isBoolean: { type: Boolean, required: false, default: false },
    isBooleanEnabled: { type: Boolean, required: false, default: false },
    actionFunction: { type: Function, required: false, default: null },
  },
  created() {
    this.switchEnabled = this.isBooleanEnabled;
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
