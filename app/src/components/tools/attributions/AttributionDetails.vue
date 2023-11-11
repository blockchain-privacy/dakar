<template>
  <v-card min-width="350px">
    <v-card-title class="d-flex align-center">
      <attribution-tag :attribution="attribution" />
      <v-spacer />
      <div class="text-subtitle-2">
        {{ attribution.ts.toLocaleDateString() }}
      </div>
      <v-menu
        v-if="!attribution.isPublic ||
          (attribution.isPublic && isAdminIdentity(session))"
        location="bottom"
      >
        <template #activator="{ props }">
          <v-btn
            icon
            v-bind="props"
            variant="plain"
          >
            <v-icon>{{ icon.mdiDotsVertical }}</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item @click="deleteItem(attribution.uid, attribution.tag, attribution.isPublic)">
            <template #prepend>
              <v-icon>{{ icon.mdiDelete }}</v-icon>
            </template>
            <v-list-item-title>Delete</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-card-title>
    <v-divider />
    <v-list-item :to="{ name: routes.addressRoute, params: { id: attribution.address }}">
      {{ attribution.address }}
    </v-list-item>
    <v-list-item v-if="attribution.description">
      Description: {{ attribution.description }}
    </v-list-item>
    <v-list-item v-if="attribution.source">
      Source: <a
        v-if="isValidHttpUrl(attribution.source)"
        :href="attribution.source"
        target="_blank"
      >{{ attribution.source }}</a>
      <template v-else>
        {{ attribution.source }}
      </template>
    </v-list-item>
    <v-list-item v-if="attribution.category">
      Category: {{ attribution.category }}
    </v-list-item>
    <delete-attribution-dialog
      v-model="deleteAttributionDialog"
      :attribution-uid="deleteAttributionUid"
      :tag="deleteAttributionTag"
      :public="deleteAttributionPublic"
      @deleted="repeatDeletionSignal"
    />
  </v-card>
</template>

<script>
import {mdiDelete, mdiDotsVertical} from '@mdi/js';
import {ROUTE_NAME_ADDRESS_PAGE} from '@/constants';
import DeleteAttributionDialog from './DeleteAttributionDialog.vue';
import AttributionTag from './AttributionTag.vue';
import {isAdminIdentity} from '@/utilities';

// Credit: https://stackoverflow.com/questions/5717093/check-if-a-javascript-string-is-a-url/43467144#43467144
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
	components: {AttributionTag, DeleteAttributionDialog},
	props: {
		attribution: {type: Object, required: true},
	},
	emits: ['deleted'],
	data() {
		return {
			icon: {mdiDelete, mdiDotsVertical},
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
		session() {
			return this.$store.getters.getSession;
		},
	},
	methods: {
		isAdminIdentity,
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
