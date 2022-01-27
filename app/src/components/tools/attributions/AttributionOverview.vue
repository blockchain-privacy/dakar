<template>
  <v-card>
    <v-card-text>
      <v-speed-dial v-model="fab" style="top:2px" right direction="left" absolute
                    transition="slide-x-reverse-transition">
        <template v-slot:activator>
          <v-btn v-model="fab" color="primary" dark fab elevation="0">
            <v-icon v-if="fab">{{ icon.mdiClose }}</v-icon>
            <v-icon v-else> {{ icon.mdiDotsVertical }}</v-icon>
          </v-btn>
        </template>
        <v-btn fab small @click="addAttributionDialog = true">
          <v-icon>{{ icon.mdiTagPlus }}</v-icon>
        </v-btn>
        <v-btn fab dark small color="red" :disabled="items.length === 0"
               @click="deleteAllAttributionsDialog = true">
          <v-icon>{{ icon.mdiDelete }}</v-icon>
        </v-btn>
      </v-speed-dial>
      <v-row>
        <v-col v-if="items.length > 0">
          <v-icon>{{ icon.mdiInformationOutline }}</v-icon>
          These attributions have been created by you.
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
      <v-row v-if="items.length > 0">
        <v-col
            v-for="(item, i) in items"
            :key="i"
            cols="12"
            sm="6"
            md="4"
            lg="4">
          <v-card outlined>
            <v-toolbar flat>
              <v-chip label class="overflow-x-auto">
                {{ item.tag }}
              </v-chip>
              <v-spacer/>
              <div>
                {{ item.ts.toLocaleDateString() }}
              </div>
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
                  <v-list-item @click="deleteItem(item.uid, item.tag)">
                    <v-list-item-icon>
                      <v-icon>{{ icon.mdiDelete }}</v-icon>
                    </v-list-item-icon>
                    <v-list-item-title>Delete</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>
            </v-toolbar>
            <v-list-item
                :to="{ name: routes.addressRoute, params: { id: item.address }}">
              <v-list-item-content>
                {{ item.address }}
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-content>
                Source:
              </v-list-item-content>
            </v-list-item>
          </v-card>
        </v-col>
      </v-row>
    </v-card-text>
    <import-attribution v-model="addAttributionDialog" @added="loadData"/>
    <delete-all-attributions v-model="deleteAllAttributionsDialog" @deleted="loadData"/>
    <delete-attribution v-model="deleteAttributionDialog" :attribution-uid="deleteAttributionUid"
                        :tag="deleteAttributionTag" @deleted="handleAttributionDeletion"/>
  </v-card>
</template>

<script>
import {
  mdiMerge, mdiDelete, mdiDotsVertical, mdiFileImport, mdiTagText, mdiTagPlus,
  mdiInformationOutline, mdiClose,
} from '@mdi/js';
import { PAGE_TITLE, ROUTE_ATTRIBUTION_OVERVIEW, ROUTE_NAME_ADDRESS_PAGE } from '../../../constants';
import { doGet, handleError } from '../../../utilities';
import ImportAttribution from '../../dialogs/ImportAttributions.vue';
import DeleteAttribution from '../../dialogs/DeleteAttribution.vue';
import DeleteAllAttributions from '../../dialogs/DeleteAllAttributions.vue';

export default {
  name: 'AttributionOverview',
  components: { DeleteAllAttributions, DeleteAttribution, ImportAttribution },
  data() {
    return {
      icon: {
        mdiMerge,
        mdiDelete,
        mdiDotsVertical,
        mdiFileImport,
        mdiTagText,
        mdiTagPlus,
        mdiInformationOutline,
        mdiClose,
      },
      routes: {
        addressRoute: ROUTE_NAME_ADDRESS_PAGE,
      },
      addAttributionDialog: false,
      deleteAttributionDialog: false,
      deleteAttributionTag: '',
      deleteAllAttributionsDialog: false,
      deleteAttributionUid: '',
      items: [],
      fab: false,
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
    deleteItem(clusterUid, tag) {
      this.deleteAttributionUid = clusterUid;
      this.deleteAttributionTag = tag;
      this.deleteAttributionDialog = true;
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
