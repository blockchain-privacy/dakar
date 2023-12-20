<template>
  <side-bar
    v-model="inputVal"
    :title="title"
    :icon="mdiShapeSquareRoundedPlus"
    max-width="648px"
  >
    <template #actions>
      <template v-if="entityData && type === 'transaction' && entityData[0].privacytype >= 0">
        <privacy-chip :privacy-type="entityData[0].privacytype" />
      </template>
    </template>
    <template #body>
      <fade-transition>
        <v-skeleton-loader
          v-if="isLoading"
          class="mx-auto"
          type="list-item-three-line, list-item-three-line, list-item-three-line"
        />
        <template v-else>
          <template v-if="entityData && type === 'transaction'">
            <!-- duplicate transaction hashes can exist -> loop through all results
            (e.g. d5d27987d2a3dfc724e359870c6644b40e497bdc0589a033220fe15429d88599 in Bitcoin) -->
            <template
              v-for="t in entityData"
              :key="t.txhash+t.bid"
            >
              <transaction
                :tx="t"
                :show-heuristic-editor-link="false"
                :show-fingerprint-link="true"
                show-details
                :embed="false"
                :show-title-bar="false"
              />
            </template>
          </template>
          <div v-else>
            Type not recognized
          </div>
        </template>
      </fade-transition>
    </template>
  </side-bar>
</template>

<script setup>
import {mdiShapeSquareRoundedPlus} from '@mdi/js';
import SideBar from '@/components/heuristic/SideBar.vue';
import {computed, inject, onUpdated, ref} from 'vue';
import Transaction from '@/components/explorer/transaction/Transaction.vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import PrivacyChip from '@/components/common/PrivacyChip.vue';
import FadeTransition from '@/components/common/FadeTransition.vue';

const props = defineProps({
	modelValue: {type: Boolean, required: true},
	identifier: {type: String, required: true},
	type: {type: String, required: true},
});

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();

const isLoading = ref(true);
const entityData = ref();

const emit = defineEmits(['update:modelValue']);

// Computed
const inputVal = computed({
	get() {
		return props.modelValue;
	},
	set(val) {
		emit('update:modelValue', val);
	},
});

const title = computed(() => {
	switch (props.type) {
		case 'transaction':
			return `Transaction ${props.identifier}`;
		case 'cluster':
			return `Address ${props.identifier}`;
		default:
			return 'unknown entity type';
	}
});

// Hooks
onUpdated(async () => {
	if (props.type === 'transaction' && props.identifier) {
		entityData.value = null;
		await getTransactionData();
	}
});

// Functions
async function getTransactionData() {
	if (props.identifier === '') {
		return;
	}

	isLoading.value = true;
	try {
		const response = await dakar.data.txHashGet({hash: props.identifier});
		entityData.value = response.payload;
	} catch (e) {
		setErrorMessage(e);
	}

	isLoading.value = false;
}

function setErrorMessage(msg) {
	msgStore.addMessage({text: msg, type: 'error', temporary: true, category: route.name});
}

</script>

<style scoped>

</style>
