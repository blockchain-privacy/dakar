<template>
  <div>
    <v-card flat v-if="!showEmptyText">
      <v-card-text>
        <v-icon>{{ icon.mdiInformationOutline }}</v-icon>
        The following clusters are attached to this address.
        New clusters can be created at the
        <router-link :to="{ name: clusterOverview}">cluster overview</router-link>
        page.
        <v-fade-transition>
          <v-row v-if="this.clusters.length > 0">
            <v-col class="d-flex justify-end align-center">
              <v-btn
                  elevation="0"
                  fab
                  color="primary"
                  @click="downloadClusterSummary">
                <v-icon>{{ icon.mdiFileDownloadOutline }}</v-icon>
              </v-btn>
            </v-col>
          </v-row>
        </v-fade-transition>
      </v-card-text>
    </v-card>
    <v-progress-linear class="mt-10" v-if="isLoading" indeterminate/>
    <v-card v-if="showEmptyText" flat class="my-3">
      <v-card-text class="text-subtitle-1" style="text-align: center">
        No clusters found
      </v-card-text>
    </v-card>
    <div v-if="this.clusters.length > 0">
      <v-card outlined v-for="(c, i) in clusters" :key="i" class="mx-3 my-3">
        <v-toolbar flat>
          <v-toolbar-title>
            {{ getClusterTypeLabel(c.type) }}
          </v-toolbar-title>
          <v-spacer></v-spacer>
          <v-chip outlined v-if="!$vuetify.breakpoint.xs">
            {{ c.addressCount }}
            {{ (c.addressCount === 1) ? 'Address' : 'Addresses' }}
          </v-chip>
          <v-chip outlined v-if="$vuetify.breakpoint.xs">
            {{ c.addressCount }}
          </v-chip>
          <v-btn v-if="c.type === 'custom'" icon
                 @click="deleteCluster(c.uid, c.addressCount)">
            <v-icon>{{ icon.mdiDelete }}</v-icon>
          </v-btn>
        </v-toolbar>
        <v-card-text>
          <attribution-tag v-for="(a, i) in c.attributions" :key="i" class="mr-2" :attribution="a"/>
        </v-card-text>
        <v-card-text v-if="c.txhash">
          <p class="text-subtitle-1">Last updated by</p>
          <ClusterDetails :tx-hash="c.txhash" :block-hash="c.bhash"
                          :block-id="c.bid" :timestamp="c.ts"/>
          <div v-if="c.hmi">
            <p class="text-subtitle-1">First included by</p>
            <ClusterDetails :tx-hash="c.hmi.txhash" :block-hash="c.hmi.bhash"
                            :block-id="c.hmi.bid" :timestamp="c.hmi.ts"/>
          </div>
        </v-card-text>
        <v-expansion-panels focusable flat
                            v-if="c.addresses && c.addresses.length > 0">
          <v-expansion-panel>
            <v-expansion-panel-header>
              Address Sample ({{ c.addresses.length }})
            </v-expansion-panel-header>
            <v-expansion-panel-content>
              <v-data-table
                  dense
                  :headers="tableHeaders"
                  :sort-by="['unspent_output_count']"
                  :items="c.addresses"
                  item-key="addresshash">
                <template v-slot:item.addresshash="{ item }">
                  <router-link :to="{ name: addressRoute, params: { id: item.addresshash }}">
                    {{ item.addresshash }}
                  </router-link>
                </template>
                <template v-slot:item.unspent_output_count="{ item }">
                  {{ item.output_count - item.spent_output_count }}
                </template>
              </v-data-table>
            </v-expansion-panel-content>
          </v-expansion-panel>
        </v-expansion-panels>
      </v-card>
    </div>
    <delete-cluster v-model="deleteClusterDialog.show"
                    :cluster-uid="deleteClusterDialog.uid"
                    :num-addresses="deleteClusterDialog.size"
                    @deleted="doLookup"/>
  </div>
</template>

<script>
import { mdiDelete, mdiFileDownloadOutline, mdiInformationOutline } from '@mdi/js';
import {
  CLUSTER_TYPE_FMI,
  CLUSTER_TYPE_HMI,
  ROUTE_CLUSTER_LOOKUP,
  ROUTE_CLUSTER_SUMMARY,
  ROUTE_NAME_ADDRESS_PAGE,
  ROUTE_NAME_BLOCK_PAGE,
  ROUTE_NAME_CLUSTER_OVERVIEW,
  ROUTE_NAME_TRANSACTION_PAGE,
} from '../../../constants';
import {
  doPost, doPostBlob, getClusterTypeLabel, getCurrentDate, handleError,
} from '../../../utilities';
import ClusterDetails from './ClusterDetails.vue';
import DeleteCluster from '../../dialogs/DeleteCluster.vue';
import AttributionTag from '../../tools/attributions/AttributionTag.vue';

export default {
  name: 'ClusterLookup',
  components: { AttributionTag, ClusterDetails, DeleteCluster },
  props: {
    addressHash: { type: String, required: true },
  },
  data() {
    return {
      icon: {
        mdiFileDownloadOutline, mdiDelete, mdiInformationOutline,
      },
      blockRoute: ROUTE_NAME_BLOCK_PAGE,
      txRoute: ROUTE_NAME_TRANSACTION_PAGE,
      addressRoute: ROUTE_NAME_ADDRESS_PAGE,
      clusterOverview: ROUTE_NAME_CLUSTER_OVERVIEW,
      // v-model
      isLoading: false,
      clusters: [],
      isClusterSummaryLoading: false,
      showEmptyText: false,
      tableHeaders: [
        {
          text: 'Addresshash',
          align: 'start',
          sortable: false,
          value: 'addresshash',
        },
        { text: 'Output count', value: 'output_count' },
        { text: 'Unspent output count', value: 'unspent_output_count' },
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
  methods: {
    getClusterTypeLabel,
    setInfoMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'info', temporary: true });
    },
    setWarningMessage(msg) {
      this.$store.dispatch('addMessage', { text: msg, type: 'warning', temporary: true });
    },
    getQuery() {
      return { addressHash: this.addressHash.trim() };
    },
    doLookup() {
      this.isLoading = true;
      this.showEmptyText = false;
      this.clusters = [];
      doPost(ROUTE_CLUSTER_LOOKUP, this.$router, this.$store, this.getQuery())
        .then((data) => {
          if (data.success === undefined) throw Error('error searching for clusters');
          if (data.success === false) {
            throw new Error(data.msg);
          }
          if (data.success === true && data.msg !== undefined) {
            this.setInfoMessage(data.msg);
          }

          if (data.clusters && data.clusters.length > 0) {
            const clusterMap = new Map();
            const clusters = [];

            // add all clusters to array if they are not hmi and fmi
            data.clusters.forEach((d) => {
              clusterMap.set(d.type, d);
              if (d.type !== CLUSTER_TYPE_HMI
                    && d.type !== CLUSTER_TYPE_FMI) clusters.push(d);
            });

            // insert hmi cluster into fmi cluster and add the composite cluster into the array
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
        })
        .catch((e) => {
          handleError(this.$store, e);
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
    downloadClusterSummary() {
      this.isClusterSummaryLoading = true;
      const body = { addressHash: this.addressHash.trim() };
      const fileName = this.addressHash;

      doPostBlob(ROUTE_CLUSTER_SUMMARY, this.$router, this.$store, body)
        .then((blob) => {
          // looks hacky, but it is the only way with good UX
          const a = document.createElement('a');
          a.href = URL.createObjectURL(blob);

          a.setAttribute(
            'download',
            `cluster_summary_${getCurrentDate()}_${fileName}.csv`,
          );
          a.click();
          a.remove();
        })
        .catch((error) => {
          handleError(this.$store, error);
        })
        .finally(() => {
          this.isClusterSummaryLoading = false;
        });
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
  created() {
    this.doLookup();
  },
};
</script>

<style scoped>

</style>
