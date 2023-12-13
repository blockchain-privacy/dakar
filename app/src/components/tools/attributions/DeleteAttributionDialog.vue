<template>
  <v-dialog
    v-model="show"
    max-width="400px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>
        <span class="text-h5">Delete Attribution</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1">
          Are you sure you want to delete the attribution <code>{{ tag }}</code>?
        </div>
        <v-row class="mt-4">
          <v-col class="d-flex justify-end align-center">
            <v-btn
              variant="text"
              :disabled="isLoading"
              @click="show = false"
            >
              Cancel
            </v-btn>
            <v-btn
              variant="text"
              :loading="isLoading"
              color="red"
              @click="deleteAttribution"
            >
              Delete
            </v-btn>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {computed, inject, ref} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();

const props = defineProps({
	modelValue: {type: Boolean, required: true},
	attributionUid: {type: String, required: true},
	tag: {type: String, required: true},
	public: {type: Boolean, required: true},
});

const emit = defineEmits(['update:modelValue', 'deleted']);

const isLoading = ref(false);

// Computed
const show = computed({
	get() {
		return props.modelValue;
	},
	set(value) {
		emit('update:modelValue', value);
	},
});

// Functions
function setPersistentErrorMessage(msg) {
	msgStore.addMessage({text: msg, type: 'error', temporary: false, category: route.name});
}

function setInfoMessage(msg) {
	msgStore.addMessage({text: msg, type: 'info', temporary: true, category: route.name});
}

async function deleteAttribution() {
	if (props.attributionUid === '') {
		setPersistentErrorMessage('could not delete attribution');
		show.value = false;
		return;
	}

	isLoading.value = true;

	try {
		const response = props.public
			? await dakar.attribution.deletePublicAttributionAttributionUidGet({attributionUid: props.attributionUid})
			: await dakar.attribution.deletePrivateAttributionAttributionUidGet({attributionUid: props.attributionUid});

		if (response.msg) {
			setInfoMessage(response.msg);
		}

		emit('deleted', props.attributionUid);
	} catch (e) {
		setPersistentErrorMessage(e);
	}

	isLoading.value = false;
	show.value = false;
}
</script>

<style scoped>

</style>
