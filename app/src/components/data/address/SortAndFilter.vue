<template>
  <v-row>
    <v-col>
      <v-select
          :disabled="loading"
          :loading="loading?'primary':false"
          style="max-width: 300px; min-width: 200px;"
          v-model="sort.selected"
          :items="sort.items"
          item-value="id"
          item-text="text"
          label="Sort"
          v-on:change="handleSortAndFilter"
      ></v-select>
    </v-col>
    <v-col>
      <v-select
          :disabled="loading"
          :loading="loading?'primary':false"
          style="max-width: 300px; min-width: 200px;"
          v-model="filter.selected"
          :items="filter.items"
          item-value="id"
          item-text="text"
          label="Filter"
          v-on:change="handleSortAndFilter"
          multiple>
        <template v-slot:selection="{ item }">
          <v-chip small>
            <span>{{ item.chip }}</span>
          </v-chip>
        </template>
      </v-select>
    </v-col>
    <v-col v-if="isSortingByInput">
      <v-alert type="info" text>Only spent outputs are shown</v-alert>
    </v-col>
  </v-row>
</template>

<script>

export default {
  name: 'SortAndFilter',
  props: {
    value: { type: Object, required: true },
    loading: { type: Boolean, required: false, default: false },
    outputCount: { type: Number, required: true },
    inputCount: { type: Number, required: true },
    coinbaseCount: { type: Number, required: true },
    dataAvailable: { type: Boolean, required: true },
  },
  data() {
    return {
      isSortingByInput: false,
      sort: {
        selected: 0,
        items: [
          { id: 0, text: 'Ascending by output date', disabled: false },
          { id: 1, text: 'Descending by output date', disabled: false },
          { divider: true },
          { id: 2, text: 'Ascending by input date', disabled: false },
          { id: 3, text: 'Descending by input date', disabled: false },
          { divider: true },
          { id: 4, text: 'Ascending by amount', disabled: false },
          { id: 5, text: 'Descending by amount', disabled: false },
        ],
      },
      filter: {
        selected: [],
        items: [
          {
            id: 0, text: 'Only show coinbase outputs', chip: 'Coinbase outputs', disabled: false,
          },
          {
            id: 1, text: 'Only show unspent outputs', chip: 'Unspent outputs', disabled: false,
          },
        ],
      },
    };
  },
  methods: {
    handleSortAndFilter() {
      this.updateSortState();
      this.updateFilterState();

      this.isSortingByInput = this.sort.selected === 2
          || this.sort.selected === 3;

      this.$emit('input', { order: this.sort.selected, filter: this.filter.selected });

      this.$emit('change');
    },
    updateSortState() {
      let isUnspentFilterSelected = false;

      if (this.dataAvailable && this.inputCount === 0) {
        isUnspentFilterSelected = true;
      } else {
        this.filter.selected.some((d) => {
          if (d === 1) {
            isUnspentFilterSelected = true;
            // break
            return true;
          }
          return false;
        });
      }

      this.sort.items.forEach((d) => {
        if (d.id === 2 || d.id === 3) {
          d.disabled = isUnspentFilterSelected;
        }
      });
    },
    updateFilterState() {
      let disableUnspentFilter = false;
      if (this.dataAvailable && this.outputCount - this.inputCount === 0) {
        disableUnspentFilter = true;
      } else {
        const selection = this.sort.selected;
        if (selection === 2 || selection === 3) {
          disableUnspentFilter = true;
        }
      }

      let disableCoinbaseFilter = false;
      if (this.dataAvailable
          && (this.coinbaseCount === 0 || this.coinbaseCount === this.outputCount)) {
        disableCoinbaseFilter = true;
      }

      this.filter.items.forEach((d) => {
        if (d.id === 0) {
          d.disabled = disableCoinbaseFilter;
        } else if (d.id === 1) {
          d.disabled = disableUnspentFilter;
        }
      });
    },
  },
  mounted() {
    this.updateSortState();
    this.updateFilterState();
  },
};
</script>

<style scoped>

</style>
