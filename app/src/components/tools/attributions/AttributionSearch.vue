<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="my-2 mx-1">
    <div class="d-flex justify-center">
      <v-form
        ref="attributionSearchForm"
        validate-on="submit"
        style="max-width: 700px; width: 100%;"
        @submit.prevent="handleQuery"
      >
        <v-text-field
          v-model="query"
          label="Search for attributions"
          :append-inner-icon="mdiMagnify"
          variant="solo"
          :rules="rule"
          :loading="loading"
          @click:append-inner="handleQuery"
        />
        <alert :text="errorMsg" />
      </v-form>
    </div>
    <template v-if="!loading && attributions.length > 0">
      <v-row
        v-for="(attribution, i) in attributions"
        :key="i"
      >
        <v-col>
          <div class="d-flex justify-center">
            <attribution-details
              :attribution="attribution"
              :blockchain-mode="blockchainMode"
              :title="title"
              @deleted="handleAttributionDeletion"
            />
          </div>
        </v-col>
      </v-row>
    </template>
  </div>
</template>

<script setup>
import {mdiMagnify} from '@mdi/js';
import {ref, useTemplateRef} from 'vue';
import AttributionDetails from './AttributionDetails.vue';
import {getDakarClient} from '@/utilities';
import Alert from '@/components/common/Alert.vue';

const props = defineProps({
	title: {type: String, required: true},
	blockchainMode: {type: String, required: true},
});

const dakar = getDakarClient(props.blockchainMode);

const loading = ref(false);
const query = ref('');
const attributions = ref([]);
const errorMsg = ref('');

const rule = [v => (Boolean(v) && v.trim().length >= 3) || 'query must be at least 3 characters long'];

const attributionForm = useTemplateRef('attributionSearchForm');

// Functions
async function handleQuery() {
	const {valid} = await attributionForm.value.validate();
	if (!valid) {
		return;
	}

	const q = query.value.trim();

	await loadSearchData(q);
}

async function loadSearchData(q) {
	loading.value = true;
	attributions.value = [];
	errorMsg.value = '';

	try {
		const response = await dakar.attribution.attributionsSearchQueryGet({q});

		if (response.attributions) {
			// Parse date
			response.attributions = response.attributions.map(d => {
				d.ts = new Date(d.ts);
				return d;
			});

			// Sort attributions by time stamp
			attributions.value = response.attributions.toSorted((a, b) => b.ts - a.ts);
		}
	} catch (error) {
		errorMsg.value = error.message;
	}

	loading.value = false;
}

function handleAttributionDeletion(attributionUid) {
	attributions.value = attributions.value.filter(d => d.uid !== attributionUid);
}

</script>

<style scoped>

</style>
