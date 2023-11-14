<template>
  <div>
    <v-card
      class="mx-auto mb-0"
      max-width="1200"
      variant="text"
    >
      <IconTitle
        title="Attributions"
        :icon="icon.mdiTag"
      />
      <v-tabs
        v-model="tab"
        align-tabs="start"
      >
        <v-tab>
          Private Attributions
        </v-tab>
        <v-tab>
          Search Public Attributions
        </v-tab>
      </v-tabs>
    </v-card>
    <v-window
      v-model="tab"
      class="mx-auto"
      style="max-width: 1200px"
      :touch="false"
    >
      <v-window-item>
        <attribution-overview />
      </v-window-item>
      <v-window-item>
        <attribution-search />
      </v-window-item>
    </v-window>
  </div>
</template>

<script>
import {mdiTag} from '@mdi/js';
import AttributionOverview from './AttributionOverview.vue';
import AttributionSearch from './AttributionSearch.vue';
import IconTitle from '@/components/common/IconTitle.vue';

export default {
	name: 'AttributionsPage',
	components: {IconTitle, AttributionSearch, AttributionOverview},
	data() {
		return {
			icon: {
				mdiTag,
			},
			tab: null,
		};
	},
};
</script>

<style scoped>

</style>
