<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div>
    <v-card
      v-if="!showEmptyText"
      variant="text"
    >
      <v-card-text class="d-flex align-center">
        <p>
          The following
          <wiki-tooltip description-url="addressCluster.md">
            clusters
          </wiki-tooltip>
          are attached to this address. New clusters can be created at the
          <router-link
            :to="{ name: ROUTE_NAME_CLUSTER_OVERVIEW}"
            class="d-inline-block"
          >
            custom clusters
          </router-link>
          page.
        </p>
        <v-fade-transition>
          <v-btn
            v-if="clusters.length > 0"
            icon
            variant="text"
            class="ms-auto"
            :loading="isClusterReportLoading"
            @click="downloadClusterReport"
          >
            <v-icon>{{ mdiFileDownloadOutline }}</v-icon>
          </v-btn>
        </v-fade-transition>
      </v-card-text>
    </v-card>
    <v-progress-linear
      v-if="isLoading"
      class="mt-10"
      indeterminate
    />
    <v-card
      v-if="showEmptyText"
      flat
      class="my-3"
      variant="text"
    >
      <v-card-text
        class="text-h6"
        style="text-align: center"
      >
        No clusters found
      </v-card-text>
    </v-card>
    <div v-if="clusters.length > 0">
      <v-card
        v-for="(c, i) in clusters"
        :key="i"
        class="mx-3 my-3"
      >
        <v-card-title class="d-flex align-center">
          {{ getClusterTypeLabel(c.type) }}
          <v-chip
            v-if="!$vuetify.display.xs"
            rounded
            class="me-2 ms-auto"
          >
            {{ c.addressCount }}
            {{ (c.addressCount === 1) ? 'Address' : 'Addresses' }}
          </v-chip>
          <v-chip
            v-if="$vuetify.display.xs"
            rounded
            class="me-2 ms-auto"
          >
            {{ c.addressCount }}
          </v-chip>
          <v-btn
            v-if="c.type === 'custom'"
            icon
            variant="text"
            @click="deleteCluster(c.uid, c.addressCount)"
          >
            <v-icon>{{ mdiDelete }}</v-icon>
          </v-btn>
        </v-card-title>
        <v-card-text v-if="c.attributions">
          <attribution-tag
            v-for="(a, y) in c.attributions"
            :key="y"
            class="mr-2"
            :attribution="a"
          />
        </v-card-text>
        <v-card-text v-if="c.txhash">
          <p class="text-subtitle-1">
            Last updated by
          </p>
          <cluster-details
            :tx-hash="c.txhash"
            :block-hash="c.bhash"
            :block-id="c.bid"
            :timestamp="c.ts"
          />
        </v-card-text>
        <v-expansion-panels
          v-if="c.addresses && c.addresses.length > 0"
        >
          <v-expansion-panel elevation="0">
            <v-expansion-panel-title>
              Address Sample ({{ c.addresses.length }})
            </v-expansion-panel-title>
            <v-expansion-panel-text>
              <v-data-table
                dense
                :headers="tableHeaders"
                :sort-by="['unspent_output_count']"
                :items="c.addresses"
                item-key="addresshash"
              >
                <template #item.addresshash="{ item }">
                  <workspace-link
                    disable-select
                    :to="{ name: ROUTE_NAME_ADDRESS_PAGE, params: { id: item.addresshash, blockchainMode: route.params.blockchainMode }}"
                  >
                    {{ item.addresshash }}
                  </workspace-link>
                </template>
                <template #item.unspent_output_count="{ item }">
                  {{ item.output_count - item.spent_output_count }}
                </template>
              </v-data-table>
            </v-expansion-panel-text>
          </v-expansion-panel>
        </v-expansion-panels>
      </v-card>
    </div>
    <delete-cluster-dialog
      v-model="deleteClusterDialogModel.show"
      :cluster-uid="deleteClusterDialogModel.uid"
      :num-addresses="deleteClusterDialogModel.size"
      :blockchain-mode="route.params.blockchainMode"
      :title="BLOCKCHAIN_ATTRIBUTES[route.params.blockchainMode].title"
      @deleted="doLookup(true)"
    />
  </div>
</template>

<script setup>
import {mdiDelete, mdiFileDownloadOutline} from '@mdi/js';
import {BLOCKCHAIN_ATTRIBUTES, ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_CLUSTER_OVERVIEW} from '@/constants';
import {
	getClusterTypeLabel, getCurrentDate, getDakarClient, handleError,
} from '@/utilities';
import ClusterDetails from './ClusterDetails.vue';
import DeleteClusterDialog from '../../tools/clusters/DeleteClusterDialog.vue';
import AttributionTag from '../../tools/attributions/AttributionTag.vue';
import WikiTooltip from '../../wiki/WikiTooltip.vue';
import {onUpdated, ref} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import WorkspaceLink from '@/components/common/WorkspaceLink.vue';

const route = useRoute();
const context = {addMessage: useMsgStore().addMessage, $route: route};
const dakar = getDakarClient(route.params.blockchainMode);

const props = defineProps({addressHash: {type: String, required: true}});

// V-model
const isLoading = ref(false);
const clusters = ref([]);
const isClusterReportLoading = ref(false);
const showEmptyText = ref(false);
const tableHeaders = [
	{
		title: 'Address Hash',
		align: 'start',
		sortable: false,
		key: 'addresshash',
	},
	{title: 'Output Count', key: 'output_count'},
	{title: 'Unspent Output Count', key: 'unspent_output_count'},
];
const deleteClusterDialogModel = ref({
	show: false,
	uid: '',
	size: -1,
});
let oldAddressHash = null;

// Hooks
onUpdated(() => {
	doLookup();
});

// Functions
async function doLookup(force = false) {
	if (!force && (!props.addressHash || props.addressHash === oldAddressHash)) {
		return;
	}

	oldAddressHash = props.addressHash;

	isLoading.value = true;
	showEmptyText.value = false;
	clusters.value = [];

	try {
		const response = await dakar.cluster.clustersHashGet({hash: props.addressHash.trim()});

		if (response.clusters?.length > 0) {
			// Add all clusters to array if they are not fmi
			clusters.value = response.clusters;
		} else {
			showEmptyText.value = true;
		}
	} catch (e) {
		handleError(context, e);
	}

	isLoading.value = false;
}

async function downloadClusterReport() {
	if (!props.addressHash) {
		return;
	}

	isClusterReportLoading.value = true;
	const fileName = props.addressHash.trim();

	try {
		const response = await dakar.cluster.clustersReportHashGet({hash: props.addressHash.trim()});

		// Looks hacky, but it is the only way with good UX
		const a = document.createElement('a');
		a.href = URL.createObjectURL(response);

		a.setAttribute(
			'download',
			`cluster_report_${getCurrentDate()}_${fileName}.csv`,
		);
		a.click();
		a.remove();
	} catch (e) {
		handleError(context, e);
	}

	isClusterReportLoading.value = false;
}

function deleteCluster(clusterUid, clusterSize) {
	if (!clusterUid || clusterSize <= 0) {
		return;
	}

	deleteClusterDialogModel.value.uid = clusterUid;
	deleteClusterDialogModel.value.size = clusterSize;
	deleteClusterDialogModel.value.show = true;
}

// Initial lookup
doLookup();

</script>

<style scoped>

</style>
