<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <template v-if="showExclusionChip">
    <v-chip
      v-tooltip="{'text': tooltipText, 'location':'bottom'}"
      rounded
      color="primary"
    >
      <template #append>
        <v-icon
          start
          :icon="mdiCloseCircle"
          @click="deleteExclusionDialog = true"
        />
      </template>
      Excluded
    </v-chip>
    <delete-address-exclusion-dialog
      v-model="deleteExclusionDialog"
      :address-hash="addressHash"
      @deleted="showExclusionChip = false"
    />
  </template>
</template>

<script setup>
import {mdiCloseCircle} from '@mdi/js';
import DeleteAddressExclusionDialog from '@/components/tools/addressExclusions/DeleteAddressExclusionDialog.vue';
import {
	computed, onMounted, ref,
} from 'vue';
import {
	getDakarClient, handleError, isAdminIdentity, isPrivilegedIdentity,
} from '@/utilities';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import {useLocalStore} from '@/pinia/local';

const props = defineProps({addressHash: {type: String, required: true}});

const route = useRoute();
const localStore = useLocalStore();
const context = {addMessage: useMsgStore().addMessage, $route: route};
const dakar = getDakarClient(route.params.blockchainMode);

const deleteExclusionDialog = ref(false);
const showExclusionChip = ref(false);

const tooltipText = 'This address is part of your address exclusion list. Click on the X to remove it from the list.';

// Computed
const session = computed({
	get() {
		return localStore.getSession;
	},
	set(value) {
		localStore.setSession(value);
	},
});

const isPrivilegedOrHigher = computed(() => isPrivilegedIdentity(session.value, route.params.blockchainMode)
	|| isAdminIdentity(session.value, route.params.blockchainMode));

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

	try {
		const response = await dakar.addressExclusion.exclusionsHashGet({hash: props.addressHash});
		showExclusionChip.value = response.isExclusion;
	} catch (e) {
		handleError(context, e);
	}
}
</script>

<style scoped>

</style>
