<template>
  <div class="my-2 mx-1">
    <v-card variant="text">
      <v-card-text>
        <v-progress-linear
          v-if="isLoading"
          :indeterminate="true"
        />
        <div v-else>
          <v-row>
            <v-col
              v-if="items.length > 0"
              class="d-flex"
            >
              <p class="text-subtitle-1 my-auto mr-auto">
                Attributions help to easier identify addresses belonging to the same
                <wiki-tooltip description-url="addressCluster.md">
                  address cluster
                </wiki-tooltip>.
              </p>
              <v-menu location="bottom">
                <template #activator="{ props }">
                  <v-btn
                    icon
                    variant="text"
                    v-bind="props"
                  >
                    <v-icon>{{ mdiDotsVertical }}</v-icon>
                  </v-btn>
                </template>
                <v-list>
                  <v-list-item @click="addAttributionDialog = true">
                    <template #prepend>
                      <v-icon>{{ mdiTagPlus }}</v-icon>
                    </template>
                    <v-list-item-title>Import Attributions</v-list-item-title>
                  </v-list-item>
                  <v-list-item @click="deleteAllAttributionsDialogModel = true">
                    <template #prepend>
                      <v-icon>{{ mdiDelete }}</v-icon>
                    </template>
                    <v-list-item-title>Delete All Attributions</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>
            </v-col>
            <v-col v-else>
              <div class="d-flex justify-center">
                <v-btn
                  variant="text"
                  @click="addAttributionDialog = true"
                >
                  <v-icon>{{ mdiFileImport }}</v-icon>
                  Import attributions
                </v-btn>
              </div>
            </v-col>
          </v-row>
        </div>
      </v-card-text>
      <import-attribution-dialog
        v-model="addAttributionDialog"
        @added="loadOverviewData"
      />
      <delete-all-attributions-dialog
        v-model="deleteAllAttributionsDialogModel"
        @deleted="loadOverviewData"
      />
    </v-card>
    <v-row
      v-if="items.length > 0"
      class="mt-3 mx-auto mb-2"
    >
      <div
        class="d-flex flex-wrap align-baseline"
        style="gap: 20px 20px"
      >
        <attribution-details
          v-for="(item, i) in items"
          :key="i"
          :attribution="item"
          @deleted="handleAttributionDeletion"
        />
      </div>
    </v-row>
  </div>
</template>

<script setup>
import {
	mdiDelete, mdiDotsVertical,	mdiFileImport, mdiTagPlus,
} from '@mdi/js';
import {PAGE_TITLE} from '@/constants';
import {handleError} from '@/utilities';
import ImportAttributionDialog from './ImportAttributionsDialog.vue';
import DeleteAllAttributionsDialog from './DeleteAllAttributionsDialog.vue';
import AttributionDetails from './AttributionDetails.vue';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';
import {inject, onMounted, ref} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const dakar = inject('dakar');
const route = useRoute();
const context = {addMessage: useMsgStore().addMessage, $route: route};

const isLoading = ref(false);
const addAttributionDialog = ref(false);
const deleteAllAttributionsDialogModel = ref(false);
const items = ref([]);

// Hooks
onMounted(async () => {
	document.title = `Attribution Overview - ${PAGE_TITLE}`;
	await loadOverviewData();
});

// Function
async function loadOverviewData() {
	isLoading.value = true;
	items.value = [];

	try {
		const response = await dakar.attribution.attributionsGet();

		if (response.attributions) {
			// Parse date
			response.attributions = response.attributions.map(d => {
				d.ts = new Date(d.ts);
				return d;
			});

			// Sort attributions by time stamp
			items.value = response.attributions.sort((a, b) => b.ts - a.ts);
		}
	} catch (e) {
		handleError(context, e);
	}

	isLoading.value = false;
}

function handleAttributionDeletion(attributionUid) {
	items.value = items.value.filter(d => d.uid !== attributionUid);
}

</script>

<style scoped>

</style>
