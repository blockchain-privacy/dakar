<template>
  <div>
    <v-card
        class="mx-auto elevation-4"
        max-width="1000">
      <v-toolbar color="primary" dark flat>
        <v-toolbar-title>
          <v-icon>{{ icon.mdiMerge }}</v-icon>
          Cluster Lookup
        </v-toolbar-title>
      </v-toolbar>
      <v-card-text>
        <div class="text-subtitle-1" v-if="!isJointLookup">
          Find clusters connected to an address.
        </div>
        <div class="text-subtitle-1" v-if="isJointLookup">
          Find clusters which are connected to both addresses.
        </div>
        <v-row>
          <v-col>
            <v-text-field label="Address"
                          v-model="a1"
                          :disabled="isLoading"
                          @keydown.enter="handleSearch"
                          autofocus/>
          </v-col>
          <v-col v-if="isJointLookup">
            <v-text-field label="Second Address"
                          v-model="a2"
                          :disabled="isLoading"
                          @keydown.enter="handleSearch"/>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-switch v-model="isJointLookup" label="Find joint clusters" :disabled="isLoading"/>
          </v-col>
          <v-col class="d-flex justify-end align-center">
            <v-btn
                color="primary"
                :disabled="!isSearchable"
                :loading="isLoading"
                @click="handleSearch">
              Search
            </v-btn>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
    <div v-if="this.clusters.length > 0">
      <v-card outlined v-for="(c, i) in clusters" :key="i"
              class="mx-auto mt-3 elevation-4" max-width="1000">
        <v-toolbar flat>
          <v-toolbar-title>
            {{ getClusterTypeLabel(c.cluster_type) }}
          </v-toolbar-title>
          <v-spacer></v-spacer>
          <v-chip outlined v-if="!$vuetify.breakpoint.xs" color="primary">
            {{ c.cluster_address_count }}
            {{ (c.cluster_address_count === 1) ? 'Address' : 'Addresses' }}
          </v-chip>
          <v-chip outlined v-if="$vuetify.breakpoint.xs">
            {{ c.cluster_address_count }}
          </v-chip>
        </v-toolbar>
        <v-card-text>
          <p class="text-subtitle-1">Last updated by</p>
          <ClusterDetails :tx-hash="c.txhash"  :block-hash="c.bhash"
                          :block-id="c.bid" :timestamp="c.ts" />
          <div v-if="isJointLookup && c.hmi">
            <p class="text-subtitle-1">First included by</p>
            <ClusterDetails :tx-hash="c.hmi.txhash"  :block-hash="c.hmi.bhash"
                            :block-id="c.hmi.bid" :timestamp="c.hmi.ts" />
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
              <v-list dense>
                <v-row>
                  <v-col v-for="(a, i) in c.cluster_addresses" :key="i">
                    <v-list-item :to="{ name: addressRoute, params: { id: a.addresshash }}">
                      <v-list-item-content>
                        <v-list-item-title>
                          {{ a.addresshash }}
                        </v-list-item-title>
                      </v-list-item-content>
                    </v-list-item>
                  </v-col>
                </v-row>
              </v-list>
            </v-expansion-panel-content>
          </v-expansion-panel>
        </v-expansion-panels>
      </v-card>
    </div>
  </div>
</template>

<script>
import { mdiMerge } from '@mdi/js';
import {
  PAGE_TITLE, ROUTE_CLUSTER_LOOKUP, ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_BLOCK_PAGE,
  ROUTE_NAME_TRANSACTION_PAGE, CLUSTER_TYPE_HMI, CLUSTER_TYPE_FMI,
} from '../../constants';
import { doPost, getClusterTypeLabel, handleError } from '../../utilities';
import ClusterDetails from './ClusterDetails.vue';

export default {
  name: 'ClusterLookup',
  components: { ClusterDetails },
  data() {
    return {
      icon: {
        mdiMerge,
      },
      blockRoute: ROUTE_NAME_BLOCK_PAGE,
      txRoute: ROUTE_NAME_TRANSACTION_PAGE,
      addressRoute: ROUTE_NAME_ADDRESS_PAGE,
      // v-model
      a1: '',
      a2: '',
      isJointLookup: false,
      isLoading: false,
      clusters: [],
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
    handleSearch() {
      if (this.isLoading || !this.isSearchable) {
        return;
      }

      this.$store.dispatch('resetMessages');

      this.clusters = [];
      this.doLookup();
    },
    doLookup() {
      this.isLoading = true;
      const body = { a1: this.a1.trim() };
      if (this.isJointLookup) {
        body.a2 = this.a2.trim();
      }

      doPost(ROUTE_CLUSTER_LOOKUP, this.$router, this.$store, body)
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
          }
        })
        .catch((e) => {
          handleError(this.$store, e);
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
  },
  mounted() {
    document.title = `Cluster Lookup - ${PAGE_TITLE}`;
  },
};
</script>

<style scoped>

</style>
