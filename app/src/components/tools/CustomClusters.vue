<template>
  <v-card
      class="mx-auto elevation-4" max-width="1200">
    <v-toolbar
        dark
        flat
        color="primary"
        class="mb-1">
      <v-toolbar-title>
        <v-icon>{{ icon.mdiMerge }}</v-icon>
        Custom Clusters
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-menu
          bottom
          left>
        <template v-slot:activator="{ on, attrs }">
          <v-btn
              dark
              icon
              v-bind="attrs"
              v-on="on">
            <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item @click="addClusterDialog = true">
            <v-list-item-title>Add Cluster</v-list-item-title>
          </v-list-item>
          <v-list-item :disabled="items.length === 0" @click="deleteAllClustersDialog = true">
            <v-list-item-title>Delete All Clusters</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-toolbar>
    <v-card-text>
      <v-row>
        <v-col v-if="items.length === 0">
          <p class="text-subtitle-1 text-center">No Clusters</p>
        </v-col>
        <v-col
            v-else
            v-for="(item, i) in items"
            :key="i"
            cols="12"
            sm="6"
            md="4"
            lg="4">
          <v-card outlined>
            <v-card-title class="subheading font-weight-bold">
              {{ item.addresses.length }} Addresses
              <v-spacer></v-spacer>
              <v-btn icon outlined @click="deleteItem(item.uid, item.addresses.length)">
                <v-icon>{{ icon.mdiDelete }}</v-icon>
              </v-btn>
            </v-card-title>
            <v-divider></v-divider>
            <v-virtual-scroll
                bench="5"
                :items="item.addresses"
                max-width="600px"
                max-height="300px"
                item-height="64">
              <template v-slot:default="{ item }">
                <v-list-item :key="item" :to="{ name: routes.addressRoute, params: { id: item }}">
                  <v-list-item-content>
                    <v-list-item-title>
                      {{ item }}
                    </v-list-item-title>
                  </v-list-item-content>
                </v-list-item>
              </template>
            </v-virtual-scroll>
          </v-card>
        </v-col>
      </v-row>
    </v-card-text>
    <add-cluster v-model="addClusterDialog"/>
    <delete-all-clusters v-model="deleteAllClustersDialog"/>
    <delete-cluster v-model="deleteClusterDialog"
                    :cluster-uid="deleteClusterUid"
                    :num-addresses="deleteClusterSize"/>
  </v-card>
</template>

<script>
import { mdiMerge, mdiDelete, mdiDotsVertical } from '@mdi/js';
import { PAGE_TITLE, ROUTE_CLUSTER_OVERVIEW, ROUTE_NAME_ADDRESS_PAGE } from '../../constants';
import { doGet, handleError } from '../../utilities';
import AddCluster from './AddCluster.vue';
import DeleteCluster from './DeleteCluster.vue';
import DeleteAllClusters from './DeleteAllClusters.vue';

export default {
  name: 'CustomClusters',
  components: { DeleteAllClusters, DeleteCluster, AddCluster },
  data() {
    return {
      icon: { mdiMerge, mdiDelete, mdiDotsVertical },
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
      doGet(ROUTE_CLUSTER_OVERVIEW, this.$router, this.$store)
        .then((d) => {
          if (!d.success || d.clusters === undefined) throw new Error('could not get cluster data');

          if (d.clusters === null) {
            this.items = [];
            return;
          }

          this.items = d.clusters;
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
  },
  mounted() {
    document.title = `Custom Clusters - ${PAGE_TITLE}`;

    this.loadData();
  },
};
</script>

<style scoped>

</style>
