<template>
  <template v-if="showExclusionChip">
    <v-chip
      :rounded="true"
      color="primary"
    >
      <template #append>
        <v-icon
          class="ms-1"
          :icon="mdiCloseCircle"
          @click="deleteExclusionDialog = true"
        />
      </template>
      <span id="address_excluded">
        Excluded
      </span>
      <v-tooltip
        activator="#address_excluded"
        location="bottom"
      >
        This address is part of your address exclusion list.
        Click on the X to remove it from the list.
      </v-tooltip>
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
import {inject, onMounted, ref} from 'vue';
import {handleError} from '@/utilities';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const props = defineProps({addressHash: {type: String, required: true}});

const dakar = inject('dakar');
const route = useRoute();
const context = {addMessage: useMsgStore().addMessage, $route: route};

onMounted(() => {
	getExclusionStatus();
});

const deleteExclusionDialog = ref(false);
const showExclusionChip = ref(false);

// Functions
async function getExclusionStatus() {
	if (props.addressHash === '') {
		return;
	}

	try {
		const response = await dakar.addressExclusion.addressExclusionStatusAddressHashGet({addressHash: props.addressHash});
		showExclusionChip.value = response.isExclusion;
	} catch (e) {
		handleError(context, e);
	}
}
</script>

<style scoped>

</style>
