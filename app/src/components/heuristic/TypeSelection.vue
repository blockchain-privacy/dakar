<template>
  <v-bottom-sheet scrollable v-model="inputVal" width="100%">
    <v-card>
      <v-card-title>
        <v-icon class="mr-2">{{ icon.mdiShapeSquareRoundedPlus }}</v-icon>
        <div class="mr-auto">Add Heuristic</div>
        <v-switch
            hide-details
            class="mr-2 mt-0 pt-0"
            v-model="isHeuristicSheetFixed"
            label="Fixed"/>
      </v-card-title>
      <v-divider/>
      <v-card-text style="height: 80%">
        <v-tabs v-model="heuristicTabs" v-if="tabItems" class="my-4">
          <v-tab v-for="item in tabItems" :key="item">
            {{ item }}
          </v-tab>
        </v-tabs>
        <v-tabs-items v-model="heuristicTabs" class="mt-3">
          <v-tab-item transition="fade-transition"
                      v-for="item in tabItems"
                      :key="item">
            <div class="d-flex flex-wrap" style="align-items: flex-start;">
              <v-card outlined
                      v-for="(item, index)
                      in heuristicTypes.filter((e) => {
                        if(!e.category && item === 'Other')
                          return true;
                        return e.category === item
                      })"
                      :key="index"
                      class="mx-3 mb-6"
                      max-width="300">
                <v-card-title>
                  {{ item.title }}
                </v-card-title>
                <v-card-subtitle>
                  {{ item.description }}
                </v-card-subtitle>
                <v-card-subtitle>
                  <v-form v-model="item.parameter.valid" v-if="item.parameter !== undefined">
                    <v-text-field
                        v-model="item.parameter.value"
                        :rules="parameterRules.get(item.parameter.type)"
                        :label="item.parameter.description"
                        required>
                    </v-text-field>
                  </v-form>
                  <v-switch label="Use custom clusters"
                      v-model="item.useCustomClusters"/>
                  <v-switch label="Use address exclusion list"
                            v-model="item.useAddressExclusionList"/>
                </v-card-subtitle>
                <v-card-actions class="pt-0">
                  <v-btn class="ml-auto" outlined color="primary"
                         @click="addNewHeuristicAction(item)">Add
                  </v-btn>
                </v-card-actions>
              </v-card>
            </div>
          </v-tab-item>
        </v-tabs-items>
      </v-card-text>
    </v-card>
  </v-bottom-sheet>
</template>

<script>
import { mdiShapeSquareRoundedPlus } from '@mdi/js';

export default {
  name: 'TypeSelection',
  props: {
    value: { type: Boolean, required: true },
    tabItems: { type: Array, required: true },
    descriptors: { type: Array, required: true },
    // events: add-heuristic
  },
  data() {
    return {
      icon: {
        mdiShapeSquareRoundedPlus,
      },
      isHeuristicSheetFixed: false,
      heuristicTabs: null,
      parameterRules: new Map([
        ['int', [(v) => {
          if (!/^\d+$/.test(v)) return false;
          const num = parseInt(v, 10);
          return Number.isInteger(num) && num > 0;
        }]],
        // string rule is not implemented yet
        ['string', null]]),
    };
  },
  computed: {
    heuristicTypes() {
      return this.descriptors.map((descriptor) => {
        descriptor.useCustomClusters = false;
        descriptor.useAddressExclusionList = false;
        return descriptor;
      });
    },
    inputVal: {
      get() {
        return this.value;
      },
      set(val) {
        this.$emit('input', val);
      },
    },
  },
  methods: {
    addNewHeuristicAction(item) {
      if (item.parameter !== undefined && !item.parameter.valid) {
        return;
      }
      if (!this.isHeuristicSheetFixed) this.inputVal = false;
      this.$emit('add-heuristic', item);
    },
  },
};
</script>

<style scoped>

</style>
