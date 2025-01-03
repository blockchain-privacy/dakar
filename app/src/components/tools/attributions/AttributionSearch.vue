<template>
  <div class="my-2 mx-1">
    <div class="d-flex justify-center">
      <v-text-field
        v-model="query"
        label="Search for attributions"
        :append-inner-icon="mdiMagnify"
        variant="solo"
        max-width="700px"
        @click:append-inner="handleQuery"
        @keydown.enter="handleQuery"
      />
    </div>

    <v-progress-linear v-if="loading" />
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

<script setup>
import {mdiMagnify} from '@mdi/js';
import AttributionDetails from './AttributionDetails.vue';
import {getDakarClient, handleError} from '@/utilities';
import {ref} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import {storeToRefs} from 'pinia';
import {useLocalStore} from '@/pinia/local.js';

const {getSettings} = storeToRefs(useLocalStore());
const route = useRoute();
const msgStore = useMsgStore();
const context = {addMessage: msgStore.addMessage, $route: route};
const dakar = getDakarClient(getSettings.value.blockchainMode);

const loading = ref(false);
const query = ref('');
const attributions = ref([]);

// Functions
function isValidQuery(query) {
	return query.trim().length > 0;
}

function setWarningMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'warning', temporary: true, category: route.name,
	});
}

async function handleQuery() {
	const q = query.value;
	if (!isValidQuery(q)) {
		setWarningMessage('search query is not valid');
		return;
	}

	await loadSearchData(q);
}

async function loadSearchData(query) {
	loading.value = true;
	attributions.value = [];

	try {
		const response = await dakar.attribution.attributionsSearchQueryGet({query});

		if (response.attributions) {
			// Parse date
			response.attributions = response.attributions.map(d => {
				d.ts = new Date(d.ts);
				return d;
			});

			// Sort attributions by time stamp
			attributions.value = response.attributions.sort((a, b) => b.ts - a.ts);
		}
	} catch (e) {
		handleError(context, e);
	}

	loading.value = false;
}

function handleAttributionDeletion(attributionUid) {
	attributions.value = attributions.value.filter(d => d.uid !== attributionUid);
}

</script>

<style scoped>

</style>
