<template>
  <v-row>
    <v-col>
      <v-select
        v-model="sort.selected"
        :disabled="loading"
        :loading="loading?'primary':false"
        style="max-width: 300px; min-width: 200px;"
        :items="sort.items"
        item-value="id"
        item-title="text"
        label="Sort"
        @update:model-value="handleSortAndFilter"
      >
        <template #item="{ props, item }">
          <v-divider v-if="item?.raw?.divider !== undefined" />
          <v-list-item
            v-else
            v-bind="props"
            :value="item.raw.id"
            :title="item?.raw?.text"
          />
        </template>
      </v-select>
    </v-col>
    <v-col>
      <v-select
        v-model="filter.selected"
        :disabled="loading"
        :loading="loading?'primary':false"
        style="max-width: 300px; min-width: 200px;"
        :items="filter.items"
        item-value="id"
        item-title="text"
        label="Filter"
        :multiple="true"
        @update:model-value="handleSortAndFilter"
      >
        <template #selection="{ item }">
          <v-chip size="small">
            <span>{{ item.raw.chip }}</span>
          </v-chip>
        </template>
      </v-select>
    </v-col>
    <v-col v-if="isSortingByInput">
      <v-alert
        type="info"
        variant="text"
      >
        Only spent outputs are shown
      </v-alert>
    </v-col>
  </v-row>
</template>

<script>

export default {
	name: 'SortAndFilter',
	props: {
		modelValue: {type: Object, required: true},
		loading: {type: Boolean, required: false, default: false},
		outputCount: {type: Number, required: true},
		inputCount: {type: Number, required: true},
		coinbaseCount: {type: Number, required: true},
		dataAvailable: {type: Boolean, required: true},
	},
	emits: ['update:modelValue', 'change'],
	data() {
		return {
			isSortingByInput: false,
			sort: {
				selected: 0,
				items: [
					{id: 0, text: 'Ascending by output date', disabled: false},
					{id: 1, text: 'Descending by output date', disabled: false},
					{divider: true},
					{id: 2, text: 'Ascending by input date', disabled: false},
					{id: 3, text: 'Descending by input date', disabled: false},
					{divider: true},
					{id: 4, text: 'Ascending by amount', disabled: false},
					{id: 5, text: 'Descending by amount', disabled: false},
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
	mounted() {
		this.updateSortState();
		this.updateFilterState();
	},
	methods: {
		handleSortAndFilter() {
			this.updateSortState();
			this.updateFilterState();

			this.isSortingByInput = this.sort.selected === 2
          || this.sort.selected === 3;

			this.$emit('update:modelValue', {order: this.sort.selected, filter: this.filter.selected});

			this.$emit('change');
		},
		updateSortState() {
			let isUnspentFilterSelected = false;

			if (this.dataAvailable && this.inputCount === 0) {
				isUnspentFilterSelected = true;
			} else {
				this.filter.selected.some(d => {
					if (d === 1) {
						isUnspentFilterSelected = true;
						// Break
						return true;
					}

					return false;
				});
			}

			this.sort.items.forEach(d => {
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

			this.filter.items.forEach(d => {
				if (d.id === 0) {
					d.disabled = disableCoinbaseFilter;
				} else if (d.id === 1) {
					d.disabled = disableUnspentFilter;
				}
			});
		},
	},
};
</script>

<style scoped>

</style>
