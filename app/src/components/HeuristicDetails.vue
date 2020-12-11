<template>
  <v-bottom-sheet scrollable v-model="inputVal">
    <v-card style="max-height: 400px">
      <v-subheader>Heuristic Properties</v-subheader>
      <v-card-text style="height: 80%">
        <div class="d-flex flex-wrap" style="align-items: flex-start;">
          <div>
            <p>Type: {{ heuristicData.heuristicType }}</p>
            <p v-if="heuristicData.heuristicParameter">
              Parameter: {{ heuristicData.heuristicParameter }}
            </p>
            <p v-if="heuristicData.resultCount">
              Number of results: {{ heuristicData.resultCount }}
            </p>
          </div>
          <v-card
              class="mx-auto my-12"
              v-for="[key, v] of addressMap"
              :key="key"
              max-width="300">
            <v-card-title>
              {{ key }}
            </v-card-title>
            <v-card-subtitle>
              Number of origins: {{ v.length }}
            </v-card-subtitle>
          </v-card>
        </div>
      </v-card-text>
    </v-card>
  </v-bottom-sheet>
</template>

<script>
export default {
  name: 'HeuristicDetails',
  props: {
    value: Boolean,
    heuristicData: Object,
    addressMap: Map,
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
