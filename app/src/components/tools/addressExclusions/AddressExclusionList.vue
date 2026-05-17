<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div
    style="max-width: 1200px"
    class="mx-auto"
  >
    <v-card variant="text">
      <icon-title
        :title="`${title} Address Exclusions`"
        :icon="mdiPlaylistRemove"
        one-line
      >
        <v-menu location="bottom">
          <template #activator="item">
            <v-btn
              v-bind="item.props"
              icon
              variant="text"
            >
              <v-icon>{{ mdiDotsVertical }}</v-icon>
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="addAddressExclusions = true">
              <template #prepend>
                <v-icon>{{ mdiFileImport }}</v-icon>
              </template>
              <v-list-item-title>Import Address Exclusions</v-list-item-title>
            </v-list-item>
            <v-list-item
              :disabled="items.length === 0"
              @click="deleteAllExclusionsDialog = true"
            >
              <template #prepend>
                <v-icon>{{ mdiDelete }}</v-icon>
              </template>
              <v-list-item-title>Delete All Address Exclusions</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
      </icon-title>
      <v-card-text>
        <p class="text-body-large">
          <wiki-tooltip description-url="addressExclusions.md">
            Address exclusions
          </wiki-tooltip> allow to exclude outputs linked to addresses from being traversed by CoinJoin heuristics.
        </p>
        <v-progress-linear
          v-if="isLoading"
          indeterminate
          class="mt-2"
        />
        <alert
          v-else-if="failedLoading"
          text="Failed loading data. Please try again later."
        />
        <p
          v-else-if="items.length > 0"
          class="text-body-large"
        >
          The address exclusion list contains {{ Number(addressCount).toLocaleString() }} address exclusions.
          The list below is limited to 30 addresses.
        </p>
        <div
          v-else
          class="d-flex justify-center"
        >
          <v-btn
            variant="text"
            @click="addAddressExclusions = true"
          >
            <v-icon>{{ mdiFileImport }}</v-icon>
            Import Address Exclusions
          </v-btn>
        </div>
      </v-card-text>
    </v-card>
    <v-row
      v-if="items.length > 0"
      class="mt-3 mx-auto"
    >
      <div
        class="d-flex flex-wrap justify-center"
        style="gap: 20px 20px"
      >
        <v-card
          v-for="addressHash in items"
          :key="addressHash"
          min-width="300px"
          max-width="400px"
          style="flex:1"
        >
          <div
            class="d-flex"
            style="flex-wrap: nowrap;"
          >
            <div
              style="min-width: 100px"
              class="align-self-center flex-grow-0 flex-shrink-1"
            >
              <!-- Add padding so the list item covers the full height of the card,
              therefore make the mouse over highlight make nicer -->
              <v-list-item
                :to="{ name: ROUTE_NAME_ADDRESS_PAGE, params: { id: addressHash, blockchainMode: blockchainMode }}"
                class="py-3"
              >
                <v-list-item-title>
                  {{ addressHash }}
                </v-list-item-title>
              </v-list-item>
            </div>
            <div class="align-self-center ml-auto">
              <v-menu location="bottom">
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
                  <v-list-item @click="deleteItem(addressHash)">
                    <template #prepend>
                      <v-icon>{{ mdiDelete }}</v-icon>
                    </template>
                    <v-list-item-title>Delete</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>
            </div>
          </div>
        </v-card>
      </div>
    </v-row>
    <import-address-exclusions-dialog
      v-model="addAddressExclusions"
      :blockchain-mode="blockchainMode"
      :title="title"
      @added="loadData"
    />
    <delete-all-address-exclusions-dialog
      v-model="deleteAllExclusionsDialog"
      :count="addressCount"
      :blockchain-mode="blockchainMode"
      :title="title"
      @deleted="loadData"
    />
    <delete-address-exclusion-dialog
      v-model="deleteExclusionDialog"
      :address-hash="deleteAddressHash"
      :blockchain-mode="blockchainMode"
      :title="title"
      @deleted="handleExclusionDeletion"
    />
  </div>
</template>

<script setup>
import {
	mdiPlaylistRemove,
	mdiDelete,
	mdiDotsVertical,
	mdiFileImport,
} from '@mdi/js';
import {onMounted, ref} from 'vue';
import ImportAddressExclusionsDialog from './ImportAddressExclusionsDialog.vue';
import DeleteAddressExclusionDialog from './DeleteAddressExclusionDialog.vue';
import DeleteAllAddressExclusionsDialog from './DeleteAllAddressExclusionsDialog.vue';
import {PAGE_TITLE, ROUTE_NAME_ADDRESS_PAGE} from '@/constants';
import {getDakarClient} from '@/utilities';
import IconTitle from '@/components/common/IconTitle.vue';
import WikiTooltip from '@/components/wiki/WikiTooltip.vue';
import Alert from '@/components/common/Alert.vue';

const props = defineProps({
	title: {type: String, required: true},
	blockchainMode: {type: String, required: true},
});

const dakar = getDakarClient(props.blockchainMode);

const addAddressExclusions = ref(false);
const deleteExclusionDialog = ref(false);
const deleteAllExclusionsDialog = ref(false);
const isLoading = ref(false);
const failedLoading = ref(false);
const deleteAddressHash = ref('');
const items = ref([]);
const addressCount = ref(-1);

// Hooks
onMounted(async () => {
	document.title = `Address Exclusions - ${PAGE_TITLE}`;
	await loadData();
});

// Functions
async function loadData() {
	items.value = [];
	addressCount.value = -1;
	isLoading.value = true;
	failedLoading.value = false;

	try {
		const response = await dakar.addressExclusion.exclusionsGet();

		if (response.addresses) {
			items.value = response.addresses;
			addressCount.value = response.addressCount;
		}
	} catch {
		failedLoading.value = true;
	}

	isLoading.value = false;
}

function deleteItem(addressHash) {
	deleteAddressHash.value = addressHash;
	deleteExclusionDialog.value = true;
}

function handleExclusionDeletion(addressHash) {
	addressCount.value -= 1;
	items.value = items.value.filter(d => d !== addressHash);
}

</script>

<style scoped>

</style>
