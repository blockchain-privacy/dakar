<template>
  <v-card
      class="mx-auto elevation-4" max-width="1200">
    <v-toolbar dark flat color="primary" class="mb-1">
      <v-toolbar-title>
        <v-icon>{{ icon.mdiTagText }}</v-icon>
        Attribution Overview
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
          <v-list-item @click="addAttributionDialog = true">
            <v-list-item-icon>
              <v-icon>{{ icon.mdiTagPlus }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Import attributions</v-list-item-title>
          </v-list-item>
          <v-list-item :disabled="items.length === 0" @click="deleteAllClustersDialog = true">
            <v-list-item-icon>
              <v-icon>{{ icon.mdiDelete }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Delete all attributions</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-toolbar>
    <v-card-text>
      <v-row>
        <v-col v-if="items.length === 0">
          <p class="text-subtitle-1 text-center">No Attributions</p>
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
            <v-list-item two-line>
              <v-list-item-subtitle class="text-right">
                {{ item.ts.toLocaleDateString() }}
              </v-list-item-subtitle>
            </v-list-item>
            <v-list-item
                :to="{ name: routes.addressRoute, params: { id: item.address }}">
              <v-list-item-content>
                <v-list-item-title>
                  {{ item.address }}
                </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-title>
                  <v-chip label>
                    {{ item.tag }}
                  </v-chip>
                </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-card-actions>
              <v-btn text @click="deleteItem(item.uid, item.address_count)">
                <v-icon>{{ icon.mdiDelete }}</v-icon>
                Delete
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-col>
      </v-row>
    </v-card-text>
    <import-attribution v-model="addAttributionDialog" @added="loadData"/>
    <!--    <delete-all-clusters v-model="deleteAllClustersDialog" @deleted="loadData" />-->
    <!--    <delete-cluster v-model="deleteClusterDialog"-->
    <!--                    :cluster-uid="deleteClusterUid"-->
    <!--                    :num-addresses="deleteClusterSize"-->
    <!--                    @deleted="handleClusterDeletion"/>-->
  </v-card>
</template>

<script>
import {
  mdiMerge, mdiDelete, mdiDotsVertical, mdiFileImport, mdiTagText, mdiTagPlus,
} from '@mdi/js';
import { PAGE_TITLE, ROUTE_ATTRIBUTION_OVERVIEW, ROUTE_NAME_ADDRESS_PAGE } from '../../constants';
import { doGet, handleError } from '../../utilities';
import ImportAttribution from '../dialogs/ImportAttributions.vue';

export default {
  name: 'AttributionOverview',
  components: { ImportAttribution },
  data() {
    return {
      icon: {
        mdiMerge, mdiDelete, mdiDotsVertical, mdiFileImport, mdiTagText, mdiTagPlus,
      },
      routes: {
        addressRoute: ROUTE_NAME_ADDRESS_PAGE,
      },
      addAttributionDialog: false,
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
      doGet(ROUTE_ATTRIBUTION_OVERVIEW, this.$router, this.$store)
        .then((data) => {
          if (!data.success || data.attributions === undefined) throw new Error('could not get attribution data');

          if (data.attributions === null) {
            this.items = [];
            return;
          }

          // parse date
          data.attributions = data.attributions.map((d) => {
            d.ts = new Date(d.ts);
            return d;
          });

          // sort attributions by time stamp
          this.items = data.attributions.sort((a, b) => b.ts - a.ts);
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
    document.title = `Attribution Overview - ${PAGE_TITLE}`;

    this.loadData();
  },
};
</script>

<style scoped>

</style>
