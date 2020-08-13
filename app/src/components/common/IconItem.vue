<template>
  <v-list-item>
    <v-list-item-avatar>
      <v-icon v-if="!isColor">{{ icon }}</v-icon>
      <v-icon v-if="isColor" v-bind:class="{ 'green--text': !isRed, 'red--text': isRed }">
        {{ icon }}
      </v-icon>
    </v-list-item-avatar>
    <v-list-item-content>
      <v-list-item-title>
        {{ title }}
        <v-tooltip right v-if="tooltip">
          <template v-slot:activator="{ on }">
            <v-hover v-slot:default="{ hover }" open-delay="0">
              <v-icon small v-on="on">
                {{ hover ? "mdi-help-circle" : "mdi-help-circle-outline" }}
              </v-icon>
            </v-hover>
          </template>
          <span>{{ tooltip }}</span>
        </v-tooltip>
      </v-list-item-title>
      <v-list-item-subtitle>
        <slot/>
      </v-list-item-subtitle>
    </v-list-item-content>
  </v-list-item>
</template>

<script>
export default {
  name: "IconItem",
  props: {
    title: {type: String, required: true},
    icon: {type: String, required: true},
    tooltip: {type: String, default: ''},
    isColor: {type: Boolean, default: false},
    isRed: {type: Boolean, default: false},
  }
}
</script>

<style scoped>

</style>