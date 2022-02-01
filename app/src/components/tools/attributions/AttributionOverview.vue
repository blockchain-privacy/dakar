<template>
  <div class="my-2 mx-1">
    <v-card elevation-4>
      <v-card-text>
        <v-progress-linear v-if="loading" indeterminate/>
        <div v-else>
          <v-row>
            <v-col v-if="items.length > 0" class="d-flex">
              <div class="my-auto mr-auto">
                <v-icon>{{ icon.mdiInformationOutline }}</v-icon>
                These attributions have been created by you.
              </div>
              <v-menu bottom left>
                <template v-slot:activator="{ on, attrs }">
                  <v-btn icon v-bind="attrs" v-on="on">
                    <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
                  </v-btn>
                </template>
                <v-list>
                  <v-list-item @click="addAttributionDialog = true">
                    <v-list-item-icon>
                      <v-icon>{{ icon.mdiTagPlus }}</v-icon>
                    </v-list-item-icon>
                    <v-list-item-title>Import Attributions</v-list-item-title>
                  </v-list-item>
                  <v-list-item @click="deleteAllAttributionsDialog = true">
                    <v-list-item-icon>
                      <v-icon>{{ icon.mdiDelete }}</v-icon>
                    </v-list-item-icon>
                    <v-list-item-title>Delete All Attributions</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>
            </v-col>
            <v-col v-else>
              <div class="d-flex justify-center">
                <v-btn @click="addAttributionDialog = true" text>
                  <v-icon>{{ icon.mdiFileImport }}</v-icon>
                  Import attributions
                </v-btn>
              </div>
            </v-col>
          </v-row>
        </div>
      </v-card-text>
      <import-attribution v-model="addAttributionDialog" @added="loadData"/>
      <delete-all-attributions v-model="deleteAllAttributionsDialog" @deleted="loadData"/>
    </v-card>
    <v-row v-if="items.length > 0" class="mt-2">
      <v-col v-for="(item, i) in items" :key="i" cols="12" sm="6" md="4" lg="4">
        <attribution-details :attribution="item" @deleted="handleAttributionDeletion"/>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import {
  mdiMerge, mdiDelete, mdiDotsVertical, mdiFileImport, mdiTagPlus,
  mdiInformationOutline, mdiClose,
} from '@mdi/js';
import { PAGE_TITLE, ROUTE_ATTRIBUTION_OVERVIEW } from '../../../constants';
import { doGet, handleError } from '../../../utilities';
import ImportAttribution from '../../dialogs/ImportAttributions.vue';
import DeleteAllAttributions from '../../dialogs/DeleteAllAttributions.vue';
import AttributionDetails from './AttributionDetails.vue';

export default {
  name: 'AttributionOverview',
  components: {
    AttributionDetails, DeleteAllAttributions, ImportAttribution,
  },
  data() {
    return {
      icon: {
        mdiMerge,
        mdiDelete,
        mdiDotsVertical,
        mdiFileImport,
        mdiTagPlus,
        mdiInformationOutline,
        mdiClose,
      },
      loading: false,
      addAttributionDialog: false,
      deleteAllAttributionsDialog: false,
      items: [],
      fab: false,
    };
  },
  methods: {
    loadData() {
      this.loading = true;
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
        })
        .finally(() => {
          this.loading = false;
        });
    },
    handleAttributionDeletion(attributionUid) {
      this.items = this.items.filter((d) => d.uid !== attributionUid);
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
