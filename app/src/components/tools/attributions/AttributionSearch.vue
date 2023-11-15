<template>
  <div class="my-2 mx-1">
    <v-card variant="text">
      <v-card-text>
        <v-text-field
          v-model="query"
          label="Search for attributions"
          :append-icon="icon.mdiMagnify"
          @click:append="handleQuery"
          @keydown.enter="handleQuery"
        />
        <v-progress-linear v-if="loading" />
      </v-card-text>
    </v-card>
    <v-row
      v-if="!loading && attributions.length > 0"
      class="mt-3 mx-auto mb-2"
    >
      <div
        class="d-flex flex-wrap align-baseline"
        style="gap: 20px 20px"
      >
        <attribution-details
          v-for="(attribution, i) in attributions"
          :key="i"
          :attribution="attribution"
          @deleted="handleAttributionDeletion"
        />
      </div>
    </v-row>
  </div>
</template>

<script>
import {mdiMagnify} from '@mdi/js';
import AttributionDetails from './AttributionDetails.vue';
import {handleError} from '@/utilities';

function isValidQuery(query) {
	return query.trim().length > 0;
}

export default {
	name: 'AttributionSearch',
	components: {AttributionDetails},
	data() {
		return {
			icon: {
				mdiMagnify,
			},
			loading: false,
			query: '',
			attributions: [],
		};
	},
	methods: {
		setWarningMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'warning', temporary: true, category: this.$route.name});
		},
		async handleQuery() {
			const q = this.query;
			if (!isValidQuery(q)) {
				this.setWarningMessage('search query is not valid');
				return;
			}

			await this.loadSearchData(q);
		},
		async loadSearchData(query) {
			this.loading = true;
			this.attributions = [];

			try {
				const response = await this.dakar.attribution.searchAttributionsPost({attribution: {q: query}});

				if (response.attributions) {
					// Parse date
					response.attributions = response.attributions.map(d => {
						d.ts = new Date(d.ts);
						return d;
					});

					// Sort attributions by time stamp
					this.attributions = response.attributions.sort((a, b) => b.ts - a.ts);
				}
			} catch (e) {
				handleError(this, e);
			}

			this.loading = false;
		},
		handleAttributionDeletion(attributionUid) {
			this.attributions = this.attributions.filter(d => d.uid !== attributionUid);
		},
	},
};
</script>

<style scoped>

</style>
