<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-card min-width="350px">
    <v-card-title class="d-flex align-center">
      <attribution-tag :attribution="attribution" />
      <v-spacer />
      <div class="text-subtitle-2">
        {{ attribution.ts.toLocaleDateString() }}
      </div>
      <v-menu
        v-if="!attribution.isPublic"
        location="bottom"
      >
        <template #activator="item">
          <v-btn
            v-bind="item.props"
            icon
            variant="plain"
          >
            <v-icon>{{ mdiDotsVertical }}</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item @click="deleteItem(attribution.uid, attribution.tag, attribution.isPublic)">
            <template #prepend>
              <v-icon>{{ mdiDelete }}</v-icon>
            </template>
            <v-list-item-title>Delete</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-card-title>
    <v-divider />
    <v-list-item :to="{ name: ROUTE_NAME_ADDRESS_PAGE, params: { id: attribution.address, blockchainMode: blockchainMode }}">
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
      v-model="deleteAttributionDialogModel"
      :attribution-uid="deleteAttributionUid"
      :tag="deleteAttributionTag"
      :public="deleteAttributionPublic"
      :blockchain-mode="blockchainMode"
      :title="title"
      @deleted="repeatDeletionSignal"
    />
  </v-card>
</template>

<script setup>
import {mdiDelete, mdiDotsVertical} from '@mdi/js';
import {ROUTE_NAME_ADDRESS_PAGE} from '@/constants';
import DeleteAttributionDialog from './DeleteAttributionDialog.vue';
import AttributionTag from './AttributionTag.vue';
import {ref} from 'vue';

defineProps({
	attribution: {type: Object, required: true},
	blockchainMode: {type: String, required: true},
	title: {type: String, required: true},
});

const emit = defineEmits(['deleted']);

const deleteAttributionDialogModel = ref(false);
const deleteAttributionTag = ref('');
const deleteAttributionUid = ref('');
const deleteAttributionPublic = ref(false);

// Functions
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

// Functions
function deleteItem(clusterUid, tag, isPublic) {
	deleteAttributionUid.value = clusterUid;
	deleteAttributionTag.value = tag;
	deleteAttributionPublic.value = isPublic;
	deleteAttributionDialogModel.value = true;
}

function repeatDeletionSignal(attributionUid) {
	emit('deleted', attributionUid);
}
</script>

<style scoped>

</style>
