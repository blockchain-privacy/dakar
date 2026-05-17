<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <template v-if="showExclusionChip">
    <v-chip
      v-tooltip="{'text': tooltipText, 'location':'bottom', 'open-delay': 400}"
      rounded
      :color="color"
    >
      <template
        v-if="!isError"
        #close
      >
        <v-icon
          :icon="mdiCloseCircle"
          @click.stop="deleteExclusionDialog = true"
        />
      </template>
      {{ title }}
    </v-chip>
    <delete-address-exclusion-dialog
      v-model="deleteExclusionDialog"
      :address-hash="addressHash"
      :blockchain-mode="route.params.blockchainMode"
      @deleted="showExclusionChip = false"
    />
  </template>
</template>

<script setup>
import {mdiCloseCircle} from '@mdi/js';
import {
	computed,
	onMounted,
	ref,
} from 'vue';
import {useRoute} from 'vue-router';
import {storeToRefs} from 'pinia';
import DeleteAddressExclusionDialog from '@/components/tools/addressExclusions/DeleteAddressExclusionDialog.vue';
import {
	getDakarClient,
	isAdminIdentity,
	isPrivilegedIdentity,
} from '@/utilities';
import {useLocalStore} from '@/pinia/local';

const props = defineProps({addressHash: {type: String, required: true}});

const route = useRoute();
const dakar = getDakarClient(route.params.blockchainMode);
const {session} = storeToRefs(useLocalStore());

const deleteExclusionDialog = ref(false);
const showExclusionChip = ref(false);
const isError = ref(false);

const tooltipTextNoError = 'This address is part of your address exclusion list. Click on the X to remove it from the list.';
const tooltipTextError = 'Error fetching the address exclusion status';

// Computed
const isPrivilegedOrHigher = computed(() => isPrivilegedIdentity(session.value, route.params.blockchainMode)
	|| isAdminIdentity(session.value, route.params.blockchainMode));

const color = computed(() =>	isError.value ? 'error' : 'primary');
const tooltipText = computed(() => isError.value ? tooltipTextError : tooltipTextNoError);
const title = computed(() => isError.value ? 'Error' : 'Excluded');

// Hooks
onMounted(() => {
	if (isPrivilegedOrHigher.value) {
		getExclusionStatus();
	}
});

// Functions
async function getExclusionStatus() {
	if (props.addressHash === '') {
		return;
	}

	isError.value = false;

	try {
		const response = await dakar.addressExclusion.exclusionsHashGet({hash: props.addressHash});
		showExclusionChip.value = response.isExclusion;
	} catch {
		isError.value = true;
	}
}
</script>

<style scoped>

</style>
