<template>
  <div>
    <v-card flat v-if="!showEmptyText">
      <v-card-text>
        <div class="text-subtitle-1" v-if="isJointLookup">
          Find clusters which are connected to both addresses.
        </div>
        <v-text-field
            v-if="isJointLookup"
            label="Address"
            v-model="a2"
            :disabled="isLoading"
            @keydown.enter="doLookup"/>
        <v-row>
          <v-col>
            <v-switch v-model="isJointLookup" label="Find joint clusters"
                      :disabled="isLoading" @change="handleSwitchChange"/>
          </v-col>
          <v-col class="d-flex justify-end align-center" v-if="isJointLookup">
            <v-btn
                color="primary"
                :disabled="!isSearchable"
                :loading="isLoading"
                @click="doLookup">
              Search
            </v-btn>
          </v-col>
        </v-row>
        <v-expand-transition>
          <div v-if="this.clusters.length > 0">
            <v-divider/>
            <v-row v-if="this.clusters.length > 0">
              <v-col class="d-flex justify-end align-center">
                <v-btn :loading="isClusterSummaryLoading"
                       color="primary" outlined @click="downloadClusterSummary"
                       class="mx-auto mt-3">
                  <v-icon>{{ icon.mdiFileDownloadOutline }}</v-icon>
                  Cluster Summary
                </v-btn>
              </v-col>
            </v-row>
          </div>
        </v-expand-transition>
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
            {{ getClusterTypeLabel(c.cluster_type) }}
          </v-toolbar-title>
          <v-spacer></v-spacer>
          <v-chip outlined v-if="!$vuetify.breakpoint.xs">
            {{ c.cluster_address_count }}
            {{ (c.cluster_address_count === 1) ? 'Address' : 'Addresses' }}
          </v-chip>
          <v-chip outlined v-if="$vuetify.breakpoint.xs">
            {{ c.cluster_address_count }}
          </v-chip>
          <v-btn v-if="c.cluster_type === 'custom'" icon outlined
                 @click="deleteCluster(c.uid, c.cluster_address_count)">
            <v-icon>{{ icon.mdiDelete }}</v-icon>
          </v-btn>
        </v-toolbar>
        <v-card-text v-if="c.txhash">
          <p class="text-subtitle-1">Last updated by</p>
          <ClusterDetails :tx-hash="c.txhash" :block-hash="c.bhash"
                          :block-id="c.bid" :timestamp="c.ts"/>
          <div v-if="!isJointLookup && c.hmi">
            <p class="text-subtitle-1">First included by</p>
            <ClusterDetails :tx-hash="c.hmi.txhash" :block-hash="c.hmi.bhash"
                            :block-id="c.hmi.bid" :timestamp="c.hmi.ts"/>
          </div>
        </v-card-text>
        <v-divider v-if="c.cluster_addresses && c.cluster_addresses.length > 0"/>
        <v-expansion-panels focusable flat
                            v-if="c.cluster_addresses && c.cluster_addresses.length > 0">
          <v-expansion-panel>
            <v-expansion-panel-header>
              Address Sample ({{ c.cluster_addresses.length }})
            </v-expansion-panel-header>
            <v-expansion-panel-content>
              <v-data-table
                  dense
                  :headers="tableHeaders"
                  :sort-by="['unspent_output_count']"
                  :items="c.cluster_addresses"
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
import { mdiFileDownloadOutline, mdiDelete } from '@mdi/js';
import {
  ROUTE_CLUSTER_LOOKUP, ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_BLOCK_PAGE,
  ROUTE_NAME_TRANSACTION_PAGE, CLUSTER_TYPE_HMI, CLUSTER_TYPE_FMI, ROUTE_CLUSTER_SUMMARY,
} from '../../../constants';
import {
  doPost, doPostBlob, getClusterTypeLabel, getCurrentDate, handleError,
} from '../../../utilities';
import ClusterDetails from './ClusterDetails.vue';
import DeleteCluster from '../../dialogs/DeleteCluster.vue';

export default {
  name: 'ClusterLookup',
  components: { ClusterDetails, DeleteCluster },
  props: {
    a1: { type: String, required: true },
  },
  data() {
    return {
      icon: {
        mdiFileDownloadOutline, mdiDelete,
      },
      blockRoute: ROUTE_NAME_BLOCK_PAGE,
      txRoute: ROUTE_NAME_TRANSACTION_PAGE,
      addressRoute: ROUTE_NAME_ADDRESS_PAGE,
      // v-model
      a2: '',
      isJointLookup: false,
      isLoading: false,
      clusters: [],
      isClusterSummaryLoading: false,
      showEmptyText: false,
      resultsAreFromJointLookup: false,
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
      const isA1Set = this.a1 && this.a1.trim().length > 0;

      if (!isA1Set) {
        return false;
      }

      if (this.isJointLookup) {
        return this.a2 && this.a2.trim().length > 0 && this.a2.trim() !== this.a1.trim();
      }
      return true;
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
      const query = { a1: this.a1.trim() };
      if (this.isJointLookup) {
        query.a2 = this.a2.trim();
      }

      return query;
    },
    doLookup() {
      this.isLoading = true;
      this.showEmptyText = false;

      this.resultsAreFromJointLookup = this.isJointLookup;

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
              clusterMap.set(d.cluster_type, d);
              if (d.cluster_type !== CLUSTER_TYPE_HMI
                    && d.cluster_type !== CLUSTER_TYPE_FMI) clusters.push(d);
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
      const body = { a1: this.a1.trim() };
      if (this.isJointLookup) {
        body.a2 = this.a2.trim();
      }

      let fileName = this.a1;

      if (this.isJointLookup) {
        fileName += `_${this.a2}`;
      }

      doPostBlob(ROUTE_CLUSTER_SUMMARY, this.$router, this.$store, body)
        .then((blob) => {
          // looks hacky, but it is the only way with good UX
          const a = document.createElement('a');
          a.href = URL.createObjectURL(blob);

          a.setAttribute('download',
            `cluster_summary_${getCurrentDate()}_${fileName}.csv`);
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
    handleSwitchChange() {
      // only do lookup if user switched back to non-joint lookup and results are from joint lookup
      if (this.resultsAreFromJointLookup && !this.isJointLookup) this.doLookup();
    },
  },
  created() {
    this.doLookup();
  },
};
</script>

<style scoped>

</style>
