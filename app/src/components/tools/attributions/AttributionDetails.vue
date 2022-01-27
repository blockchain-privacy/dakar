<template>
  <v-card outlined>
    <v-toolbar flat>
      <v-chip label class="overflow-x-auto">
        {{ attribution.tag }}
      </v-chip>
      <v-spacer/>
      <div>
        {{ attribution.ts.toLocaleDateString() }}
      </div>
      <v-menu bottom left>
        <template v-slot:activator="{ on, attrs }">
          <v-btn icon v-bind="attrs" v-on="on">
            <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item @click="deleteItem(attribution.uid, attribution.tag)">
            <v-list-item-icon>
              <v-icon>{{ icon.mdiDelete }}</v-icon>
            </v-list-item-icon>
            <v-list-item-title>Delete</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-toolbar>
    <v-list-item :to="{ name: routes.addressRoute, params: { id: attribution.address }}">
      <v-list-item-content>
        {{ attribution.address }}
      </v-list-item-content>
    </v-list-item>
    <v-list-item>
      <v-list-item-content>
        Source:
      </v-list-item-content>
    </v-list-item>
    <delete-attribution v-model="deleteAttributionDialog" :attribution-uid="deleteAttributionUid"
                        :tag="deleteAttributionTag" @deleted="repeatDeletionSignal"/>
  </v-card>
</template>

<script>
import { mdiDelete, mdiDotsVertical } from '@mdi/js';
import { ROUTE_NAME_ADDRESS_PAGE } from '../../../constants';
import DeleteAttribution from '../../dialogs/DeleteAttribution.vue';

export default {
  name: 'AttributionDetails',
  components: { DeleteAttribution },
  props: {
    attribution: { type: Object, required: true },
  },
  data() {
    return {
      icon: { mdiDelete, mdiDotsVertical },
      routes: {
        addressRoute: ROUTE_NAME_ADDRESS_PAGE,
      },
      deleteAttributionDialog: false,
      deleteAttributionTag: '',
      deleteAttributionUid: '',
    };
  },
  methods: {
    deleteItem(clusterUid, tag) {
      this.deleteAttributionUid = clusterUid;
      this.deleteAttributionTag = tag;
      this.deleteAttributionDialog = true;
    },
    repeatDeletionSignal(attributionUid) {
      this.$emit('deleted', attributionUid);
    },
  },
};
</script>

<style scoped>

</style>
