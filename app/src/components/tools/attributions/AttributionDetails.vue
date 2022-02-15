<template>
  <v-card>
    <v-toolbar flat>
      <attribution-tag :attribution="attribution"/>
      <v-spacer/>
      <div>
        {{ attribution.ts.toLocaleDateString() }}
      </div>
      <v-menu bottom left v-if="!attribution.isPublic ||
       (attribution.isPublic && isAdminUser(userData))">
        <template v-slot:activator="{ on, attrs }">
          <v-btn icon v-bind="attrs" v-on="on">
            <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item @click="deleteItem(attribution.uid, attribution.tag, attribution.isPublic)">
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
    <v-list-item v-if="attribution.description">
      <v-list-item-content>
        Description: {{ attribution.description }}
      </v-list-item-content>
    </v-list-item>
    <v-list-item v-if="attribution.source">
      <v-list-item-content>
        Source: <a :href="attribution.source" target="_blank"
                   v-if="isValidHttpUrl(attribution.source)">{{ attribution.source }}</a>
        <template v-else>{{ attribution.source }}</template>
      </v-list-item-content>
    </v-list-item>
    <v-list-item v-if="attribution.category">
      <v-list-item-content>
        Category: {{ attribution.category }}
      </v-list-item-content>
    </v-list-item>
    <delete-attribution v-model="deleteAttributionDialog" :attribution-uid="deleteAttributionUid"
                        :tag="deleteAttributionTag" @deleted="repeatDeletionSignal"
                        :public="this.deleteAttributionPublic"/>
  </v-card>
</template>

<script>
import { mdiDelete, mdiDotsVertical } from '@mdi/js';
import { ROUTE_NAME_ADDRESS_PAGE } from '../../../constants';
import DeleteAttribution from '../../dialogs/DeleteAttribution.vue';
import AttributionTag from './AttributionTag.vue';
import { isAdminUser } from '../../../utilities';

// credit: https://stackoverflow.com/questions/5717093/check-if-a-javascript-string-is-a-url/43467144#43467144
function isValidHttpUrl(string) {
  let url;

  try {
    url = new URL(string);
  } catch (_) {
    return false;
  }

  return url.protocol === 'http:' || url.protocol === 'https:';
}

export default {
  name: 'AttributionDetails',
  components: { AttributionTag, DeleteAttribution },
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
      deleteAttributionPublic: false,
      uuid: '',
    };
  },
  computed: {
    userData: {
      get() {
        return this.$store.getters.getActiveUser;
      },
      set(value) {
        this.$store.dispatch('setActiveUser', value);
      },
    },
  },
  methods: {
    isAdminUser,
    isValidHttpUrl,
    deleteItem(clusterUid, tag, isPublic) {
      this.deleteAttributionUid = clusterUid;
      this.deleteAttributionTag = tag;
      this.deleteAttributionPublic = isPublic;
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
