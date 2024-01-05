<template>
  <template v-if="$vuetify.display.lgAndUp">
    <v-btn
      v-for="(item, i) in items"
      :key="i"
      variant="text"
      :to="item.to?item.to:undefined"
      class="my-2"
      @click="item.action?item.action():undefined"
    >
      <v-icon
        v-if="item.icon"
        :icon="item.icon"
        class="me-1"
      />
      {{ item.title }}
    </v-btn>
  </template>
  <v-menu
    v-if="$vuetify.display.mdAndDown"
    location="bottom"
  >
    <template #activator="{ props }">
      <v-btn
        :icon="true"
        variant="text"
        v-bind="props"
        style="outline: 0"
      >
        <v-icon>{{ mdiDotsVertical }}</v-icon>
      </v-btn>
    </template>
    <v-list>
      <template
        v-for="(item, i) in items"
        :key="i"
      >
        <v-list-item
          :to="item.to?item.to:undefined"
          @click="item.action?item.action():undefined"
        >
          <template
            v-if="item.icon"
            #prepend
          >
            <v-icon :icon="item.icon" />
          </template>
          <v-list-item-title>
            {{ item.title }}
          </v-list-item-title>
        </v-list-item>
      </template>
    </v-list>
  </v-menu>
</template>

<script setup>
import {mdiDotsVertical} from '@mdi/js';

defineProps({
	items: {type: Array, required: true},
});

</script>

<style scoped>

</style>
