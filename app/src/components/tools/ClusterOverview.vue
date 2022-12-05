<template>
  <div>
    <v-card class="mx-auto elevation-4" max-width="1200">
      <v-toolbar dark flat color="primary" class="mb-1">
        <v-toolbar-title>
          <v-icon>{{ icon.mdiMerge }}</v-icon>
          Cluster Overview
        </v-toolbar-title>
        <v-spacer></v-spacer>
        <v-menu bottom left>
          <template v-slot:activator="{ on, attrs }">
            <v-btn dark icon v-bind="attrs" v-on="on">
              <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="addClusterDialog = true">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiFileImport }}</v-icon>
              </v-list-item-icon>
              <v-list-item-title>Import clusters</v-list-item-title>
            </v-list-item>
            <v-list-item :disabled="items.length === 0" @click="deleteAllClustersDialog = true">
              <v-list-item-icon>
                <v-icon>{{ icon.mdiDelete }}</v-icon>
              </v-list-item-icon>
              <v-list-item-title>Delete all clusters</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
      </v-toolbar>
      <v-card-text>
        <v-row>
          <v-col v-if="items.length > 0">
            <v-icon>{{ icon.mdiInformationOutline }}</v-icon>
            These <WikiTooltip description-url="addressCluster.md">clusters</WikiTooltip>
            have been created by you.
          </v-col>
          <v-col v-else>
            <div class="d-flex justify-center">
              <v-btn @click="addClusterDialog = true" text>
                <v-icon>{{ icon.mdiFileImport }}</v-icon>
                Import clusters
              </v-btn>
            </div>
          </v-col>
        </v-row>
      </v-card-text>
      <import-cluster v-model="addClusterDialog" @added="loadData"/>
      <delete-all-clusters v-model="deleteAllClustersDialog" @deleted="loadData"/>
      <delete-cluster v-model="deleteClusterDialog"
                      :cluster-uid="deleteClusterUid"
                      :num-addresses="deleteClusterSize"
                      @deleted="handleClusterDeletion"/>
    </v-card>
    <v-row v-if="items.length > 0" class="mt-2 mx-auto"
           style="max-width: 1200px; background-color: transparent">
      <v-col v-for="(item, i) in items" :key="i" cols="12" sm="6" md="4" lg="4">
        <v-card elevation="4">
          <v-list-item two-line>
            <v-list-item-title>
              {{ item.address_count.toLocaleString() }} Addresses
            </v-list-item-title>
            <v-list-item-subtitle class="text-right">
              {{ item.ts.toLocaleDateString() }}
            </v-list-item-subtitle>
            <v-menu
                bottom
                left>
              <template v-slot:activator="{ on, attrs }">
                <v-btn
                    icon
                    v-bind="attrs"
                    v-on="on">
                  <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
                </v-btn>
              </template>
              <v-list>
                <v-list-item @click="deleteItem(item.uid, item.address_count)">
                  <v-list-item-icon>
                    <v-icon>{{ icon.mdiDelete }}</v-icon>
                  </v-list-item-icon>
                  <v-list-item-title>Delete</v-list-item-title>
                </v-list-item>
              </v-list>
            </v-menu>
          </v-list-item>
          <v-list-item
              v-for="address in item.addresses"
              :key="address"
              :to="{ name: routes.addressRoute, params: { id: address }}">
            <v-list-item-content>
              {{ address }}
            </v-list-item-content>
          </v-list-item>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import {
  mdiMerge, mdiDelete, mdiDotsVertical, mdiFileImport, mdiInformationOutline,
} from '@mdi/js';
import { PAGE_TITLE, ROUTE_CLUSTER_OVERVIEW, ROUTE_NAME_ADDRESS_PAGE } from '../../constants';
import { doGet, handleError } from '../../utilities';
import ImportCluster from '../dialogs/ImportClusters.vue';
import DeleteCluster from '../dialogs/DeleteCluster.vue';
import DeleteAllClusters from '../dialogs/DeleteAllClusters.vue';
import WikiTooltip from '../wiki/WikiTooltip.vue';

export default {
  name: 'ClusterOverview',
  components: {
    WikiTooltip, DeleteAllClusters, DeleteCluster, ImportCluster,
  },
  data() {
    return {
      icon: {
        mdiMerge, mdiDelete, mdiDotsVertical, mdiFileImport, mdiInformationOutline,
      },
      routes: {
        addressRoute: ROUTE_NAME_ADDRESS_PAGE,
      },
      addClusterDialog: false,
      deleteClusterDialog: false,
      deleteAllClustersDialog: false,
      deleteClusterUid: '',
      deleteClusterSize: -1,
      items: [],
    };
  },
  methods: {
    loadData() {
      this.items = [];
      doGet(ROUTE_CLUSTER_OVERVIEW, this.$router, this.$store)
        .then((data) => {
          if (!data.success || data.clusters === undefined) throw new Error('could not get cluster data');

          if (data.clusters === null) {
            this.items = [];
            return;
          }

          // parse date
          data.clusters = data.clusters.map((d) => {
            d.ts = new Date(d.ts);
            return d;
          });

          // sort clusters by time stamp
          this.items = data.clusters.sort((a, b) => b.ts - a.ts);
        })
        .catch((e) => {
          handleError(this.$store, e);
        });
    },
    deleteItem(clusterUid, clusterSize) {
      this.deleteClusterUid = clusterUid;
      this.deleteClusterSize = clusterSize;
      this.deleteClusterDialog = true;
    },
    handleClusterDeletion(clusterUid) {
      this.items = this.items.filter((d) => d.uid !== clusterUid);
    },
  },
  mounted() {
    document.title = `Cluster Overview - ${PAGE_TITLE}`;

    this.loadData();
  },
};
</script>

<style scoped>

</style>
