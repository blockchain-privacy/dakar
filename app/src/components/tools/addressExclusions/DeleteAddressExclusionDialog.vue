<template>
  <v-dialog
    v-model="show"
    max-width="400px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>
        <span class="text-h5">Delete Address Exclusion</span>
      </v-card-title>
      <v-card-text>
        <div class="text-subtitle-1 text-break">
          Are you sure you want to delete the address <code>{{ addressHash }}</code>
          from the address exclusion list?
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
              @click="deleteAddressExclusion"
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
import {useStore} from 'vuex';
import {useRoute} from 'vue-router';

const dakar = inject('dakar');
const store = useStore();
const route = useRoute();

const props = defineProps({
	modelValue: {type: Boolean, required: true},
	addressHash: {type: String, required: true},
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
	store.dispatch('addMessage', {text: msg, type: 'error', temporary: false, category: route.name});
}

async function deleteAddressExclusion() {
	if (props.addressHash === '') {
		setPersistentErrorMessage('could not delete address exclusion');
		show.value = false;
		return;
	}

	isLoading.value = true;

	try {
		await dakar.addressExclusion.deleteAddressExclusionAddressHashGet({addressHash: props.addressHash});
		emit('deleted', props.addressHash);
	} catch (e) {
		setPersistentErrorMessage(e);
	}

	isLoading.value = false;
	show.value = false;
}

</script>

<style scoped>

</style>
