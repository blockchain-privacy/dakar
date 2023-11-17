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
            :to="{ name: clusterOverview}"
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
            @click="downloadClusterSummary"
          >
            <v-icon>{{ icon.mdiFileDownloadOutline }}</v-icon>
          </v-btn>
        </v-fade-transition>
      </v-card-text>
    </v-card>
    <v-progress-linear
      v-if="isLoading"
      class="mt-10"
      :indeterminate="true"
    />
    <v-card
      v-if="showEmptyText"
      :flat="true"
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
            :rounded="true"
            class="me-2 ms-auto"
          >
            {{ c.addressCount }}
            {{ (c.addressCount === 1) ? 'Address' : 'Addresses' }}
          </v-chip>
          <v-chip
            v-if="$vuetify.display.xs"
            :rounded="true"
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
            <v-icon>{{ icon.mdiDelete }}</v-icon>
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
          <div v-if="c.hmi">
            <p class="text-subtitle-1">
              First included by
            </p>
            <cluster-details
              :tx-hash="c.hmi.txhash"
              :block-hash="c.hmi.bhash"
              :block-id="c.hmi.bid"
              :timestamp="c.hmi.ts"
            />
          </div>
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
                  <router-link :to="{ name: addressRoute, params: { id: item.addresshash }}">
                    {{ item.addresshash }}
                  </router-link>
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
      v-model="deleteClusterDialog.show"
      :cluster-uid="deleteClusterDialog.uid"
      :num-addresses="deleteClusterDialog.size"
      @deleted="doLookup"
    />
  </div>
</template>

<script>
import {mdiDelete, mdiFileDownloadOutline} from '@mdi/js';
import {
	CLUSTER_TYPE_FMI,
	CLUSTER_TYPE_HMI,
	ROUTE_NAME_ADDRESS_PAGE,
	ROUTE_NAME_BLOCK_PAGE,
	ROUTE_NAME_CLUSTER_OVERVIEW,
	ROUTE_NAME_TRANSACTION_PAGE,
} from '@/constants';
import {getClusterTypeLabel, getCurrentDate, handleError} from '@/utilities';
import ClusterDetails from './ClusterDetails.vue';
import DeleteClusterDialog from '../../tools/clusters/DeleteClusterDialog.vue';
import AttributionTag from '../../tools/attributions/AttributionTag.vue';
import WikiTooltip from '../../wiki/WikiTooltip.vue';

export default {
	name: 'ClusterLookup',
	components: {
		AttributionTag, ClusterDetails, DeleteClusterDialog, WikiTooltip,
	},
	props: {
		addressHash: {type: String, required: true},
	},
	data() {
		return {
			icon: {
				mdiFileDownloadOutline, mdiDelete,
			},
			blockRoute: ROUTE_NAME_BLOCK_PAGE,
			txRoute: ROUTE_NAME_TRANSACTION_PAGE,
			addressRoute: ROUTE_NAME_ADDRESS_PAGE,
			clusterOverview: ROUTE_NAME_CLUSTER_OVERVIEW,
			// V-model
			isLoading: false,
			clusters: [],
			isClusterSummaryLoading: false,
			showEmptyText: false,
			tableHeaders: [
				{
					title: 'Address Hash',
					align: 'start',
					sortable: false,
					key: 'addresshash',
				},
				{title: 'Output Count', key: 'output_count'},
				{title: 'Unspent Output Count', key: 'unspent_output_count'},
			],
			deleteClusterDialog: {
				show: false,
				uid: '',
				size: -1,
			},
		};
	},
	computed: {
		isSearchable() {
			return this.addressHash && this.addressHash.trim().length > 0;
		},
	},
	created() {
		this.doLookup();
	},
	methods: {
		getClusterTypeLabel,
		setInfoMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'info', temporary: true, category: this.$route.name});
		},
		setWarningMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'warning', temporary: true, category: this.$route.name});
		},
		getQuery() {
			return {addressHash: this.addressHash.trim()};
		},
		async doLookup() {
			this.isLoading = true;
			this.showEmptyText = false;
			this.clusters = [];

			try {
				const response = await this.dakar.cluster.clusterLookupAddressHashGet(this.getQuery());

				if (response.clusters && response.clusters.length > 0) {
					const clusterMap = new Map();
					const clusters = [];

					// Add all clusters to array if they are not hmi and fmi
					response.clusters.forEach(d => {
						clusterMap.set(d.type, d);
						if (d.type !== CLUSTER_TYPE_HMI
              && d.type !== CLUSTER_TYPE_FMI) {
							clusters.push(d);
						}
					});

					// Insert hmi cluster into fmi cluster and add the composite cluster into the array
					if (clusterMap.has(CLUSTER_TYPE_FMI)) {
						const fmiCluster = clusterMap.get(CLUSTER_TYPE_FMI);
						if (clusterMap.has(CLUSTER_TYPE_HMI)) {
							fmiCluster.hmi = clusterMap.get(CLUSTER_TYPE_HMI);
						}

						clusters.push(fmiCluster);
					}

					this.clusters = clusters;
				} else {
					this.showEmptyText = true;
				}
			} catch (e) {
				handleError(this, e);
			}

			this.isLoading = false;
		},
		async downloadClusterSummary() {
			this.isClusterSummaryLoading = true;
			const fileName = this.addressHash;

			try {
				const response = await 	this.dakar.cluster.clusterSummaryAddressHashGet({addressHash: this.addressHash.trim()});

				// Looks hacky, but it is the only way with good UX
				const a = document.createElement('a');
				a.href = URL.createObjectURL(response);

				a.setAttribute(
					'download',
					`cluster_summary_${getCurrentDate()}_${fileName}.csv`,
				);
				a.click();
				a.remove();
			} catch (e) {
				handleError(this, e);
			}

			this.isClusterSummaryLoading = false;
		},
		deleteCluster(clusterUid, clusterSize) {
			if (!clusterUid || clusterSize <= 0) {
				return;
			}

			this.deleteClusterDialog.uid = clusterUid;
			this.deleteClusterDialog.size = clusterSize;
			this.deleteClusterDialog.show = true;
		},
	},
};
</script>

<style scoped>

</style>
