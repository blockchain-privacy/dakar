<!-- source: https://github.com/vuetifyjs/vuetify/issues/1877#issuecomment-593273676  -->
<template>
  <v-menu
      v-model="inputVal"
      :position-x="positionX"
      :position-y="positionY"
      :absolute="absolute"
      :offset-x='isOffsetX'
      :offset-y='isOffsetY'
      :open-on-hover='isOpenOnHover'
      :transition='transition'>
    <template v-slot:activator="{ on }">
      <v-btn v-if='icon' :color='color' v-on="on">
        <v-icon>{{ icon }}</v-icon>
      </v-btn>
      <v-list-item v-else-if='isSubMenu' class='d-flex justify-space-between' v-on="on">
        {{ name }}
        <v-icon>{{ icons.mdiChevronRight }}</v-icon>
      </v-list-item>
      <!--      <v-btn v-else :color='color' v-on="on" text tile>{{ name }}</v-btn>-->
    </template>
    <v-list>
      <template v-for="(item, index) in menuItems">
        <v-divider v-if='item.isDivider' :key='index'/>
        <nested-menu v-else-if='item.menu' :key='index' :name='item.title' :menu-items='item.menu'
                     @nested-menu-click='emitClickEvent'
                     :is-open-on-hover=false :is-offset-x=true :is-offset-y=false :is-sub-menu=true
        />
        <v-list-item v-else :key='index' @click='emitClickEvent(item)'
                     :disabled="item.disabled && !item.disabled()">
          <v-list-item-icon v-if="item.icon">
            <v-icon>{{ item.icon }}</v-icon>
          </v-list-item-icon>
          <v-list-item-title>{{ item.title }}</v-list-item-title>
        </v-list-item>
      </template>
    </v-list>
  </v-menu>
</template>

<script>
import {
  mdiChevronRight,
} from '@mdi/js';

export default {
  name: 'NestedMenu',
  props: {
    value: Boolean,
    name: String,
    icon: String,
    menuItems: Array,
    absolute: { type: Boolean, default: false },
    color: { type: String, default: 'secondary' },
    positionX: { type: Number, default: 0 },
    positionY: { type: Number, default: 0 },
    isOffsetX: { type: Boolean, default: false },
    isOffsetY: { type: Boolean, default: true },
    isOpenOnHover: { type: Boolean, default: false },
    isSubMenu: { type: Boolean, default: false },
    transition: { type: String, default: 'scale-transition' },
  },
  data() {
    return {
      icons: {
        mdiChevronRight,
      },
    };
  },
  methods: {
    emitClickEvent(item) {
      // this.closeAllMenus() // Theoretically, create a method that does this as a workaround
      this.$emit('nested-menu-click', item);
    },
  },
  computed: {
    inputVal: {
      get() {
        return this.value;
      },
      set(val) {
        this.$emit('input', val);
      },
    },
  },
};
</script>

<style scoped>

</style>
