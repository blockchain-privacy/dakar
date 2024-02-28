<template>
  <side-bar
    v-model="model"
    :title="title"
    :icon="sideBarIcon"
    max-width="700px"
    :title-one-line="false"
  >
    <template #actions>
      <div class="overflow-auto">
        <v-chip
          :rounded="true"
          class="me-2"
          variant="tonal"
          :prepend-icon="mdiDelete"
          @click="emitDeleteEntity"
        >
          Delete
        </v-chip>
        <template v-if="!isLoading && entityData">
          <template v-if="type === 'heuristic' || (type === 'transaction' && entityData[0]?.privacytype >= 0)">
            <v-chip
              v-if="type === 'heuristic' || isDestination(entityData[0].privacytype)"
              :rounded="true"
              color="primary"
              variant="tonal"
              class="me-2"
              :prepend-icon="mdiShapeCirclePlus"
              @click="handleAddHeuristicClick"
            >
              Add Heuristic
            </v-chip>
          </template>
          <fingerprint-chip
            v-if="type === 'transaction' && entityData[0] && isDestination(entityData[0].privacytype)"
            :transaction-hash="identifier"
            class="me-2"
          />
          <privacy-chip
            v-if="type === 'transaction' && entityData[0]?.privacytype >= 0"
            :privacy-type="entityData[0].privacytype"
          />
          <exclusion-chip
            v-else-if="type === 'cluster' && entityData?.addresshash"
            :address-hash="entityData.addresshash"
          />
          <v-chip
            v-else-if="type === 'heuristic' && entityData?.clusterCount > 0"
            :rounded="true"
            color="primary"
            variant="tonal"
            :prepend-icon="mdiFileDownloadOutline"
            @click="downloadReport"
          >
            Download
          </v-chip>
        </template>
      </div>
    </template>
    <template #body>
      <fade-transition>
        <v-skeleton-loader
          v-if="isLoading"
          class="mx-auto"
          type="list-item-three-line, list-item-three-line, list-item-three-line"
        />
        <template v-else>
          <template v-if="type === 'transaction' && entityData?.length">
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
          <address-view
            v-else-if="entityData && type === 'cluster'"
            :address-data="entityData"
            :show-title-bar="false"
          />
          <heuristic-details
            v-else-if="entityData.heuristicUid && type === 'heuristic'"
            :heuristic-data="entityData"
          />
          <div v-else>
            Type not recognized
          </div>
        </template>
      </fade-transition>
    </template>
  </side-bar>
  <v-dialog
    v-model="routeGuardDialogModel"
    max-width="300px"
    :contained="true"
    :no-click-animation="true"
  >
    <v-card>
      <v-card-text class="d-flex align-center flex-column">
        <v-btn
          class="mx-auto"
          variant="text"
          size="x-large"
          :to="routeGuardTo"
          target="_blank"
          @click="routeGuardDialogModel = false"
        >
          <v-icon
            :icon="mdiOpenInNew"
            class="me-2"
          />
          <div
            class="shorten"
            style="max-width: 200px; text-transform: none !important;"
          >
            Go to {{ routeGuardId }}
          </div>
        </v-btn>
        <named-divider
          title="Or"
          style="width:100%"
          :vertical-margin="2"
        />
        <v-btn
          class="mx-auto"
          variant="text"
          size="x-large"
          style="text-transform: none !important;"
          :disabled="disableAddingNodes"
          @click="handleRouteGuardDialogAdd"
        >
          <v-icon
            :icon="mdiPlus"
            class="me-2"
          />
          Add to Workspace
        </v-btn>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {
	mdiCardBulletedOutline,
	mdiChartBar,
	mdiDelete,
	mdiFileDownloadOutline,
	mdiOpenInNew,
	mdiPlus,
	mdiShapeCirclePlus,
	mdiTransfer,
} from '@mdi/js';
import SideBar from '@/components/common/SideBar.vue';
import {
	computed, inject, onUpdated, ref,
} from 'vue';
import Transaction from '@/components/explorer/transaction/Transaction.vue';
import AddressView from '@/components/explorer/address/Address.vue';
import {onBeforeRouteLeave, useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';
import PrivacyChip from '@/components/common/PrivacyChip.vue';
import FadeTransition from '@/components/common/FadeTransition.vue';
import ExclusionChip from '@/components/explorer/address/ExclusionChip.vue';
import {useCacheStore} from '@/pinia/cache';
import HeuristicDetails from '@/components/workspace/HeuristicDetails.vue';
import {getCurrentDate, isDestination} from '@/utilities';
import {ROUTE_NAME_ADDRESS_PAGE, ROUTE_NAME_TRANSACTION_PAGE} from '@/constants';
import NamedDivider from '@/components/common/NamedDivider.vue';
import FingerprintChip from '@/components/explorer/transaction/FingerprintChip.vue';

const props = defineProps({
	identifier: {type: String, required: true},
	type: {type: String, required: true},
	auxiliaryData: {type: Object, required: false, default: null},
	disableAddingNodes: {type: Boolean, required: true},
});
const emit = defineEmits(['addHeuristic', 'addNode', 'deleteEntity']);
const model = defineModel({type: Boolean});

const dakar = inject('dakar');
const route = useRoute();
const msgStore = useMsgStore();
const cacheStore = useCacheStore();

const isLoading = ref(true);
const entityData = ref();

let oldIdentifier = null;
let routeGuardTo = null;
let routeGuardId = '';
const routeGuardDialogModel = ref(false);

// Computed
const title = computed(() => {
	switch (props.type) {
		case 'transaction':
			return `Transaction ${props.identifier}`;
		case 'cluster':
			return `Address ${props.identifier}`;
		case 'heuristic':
			return 'Heuristic Properties';
		default:
			return 'unknown entity type';
	}
});

// Hooks
onUpdated(async () => {
	if (props.identifier && props.identifier !== oldIdentifier) {
		isLoading.value = true;
		oldIdentifier = props.identifier;
		// Check if value is in cache, otherwise get data from backend
		const cacheValue = cacheStore.getValue(props.identifier);
		if (cacheValue !== undefined) {
			entityData.value = cacheValue;
		} else if (props.type === 'transaction') {
			await getTransactionData();
		} else if (props.type === 'cluster') {
			await getAddressData();
		} else if (props.type === 'heuristic') {
			await getHeuristicData();
		}

		isLoading.value = false;
	}
});

onBeforeRouteLeave(to => {
	// Don't activate route guard if sidbar is closed
	if (!model.value) {
		return true;
	}

	if ((to.name === ROUTE_NAME_TRANSACTION_PAGE || to.name === ROUTE_NAME_ADDRESS_PAGE) && to.params?.id) {
		routeGuardTo = to;
		routeGuardId = to.params.id;
		routeGuardDialogModel.value = true;
		return false;
	}

	return true;
});

// Computed
const sideBarIcon = computed(() => {
	switch (props.type) {
		case 'transaction':
			return mdiTransfer;
		case 'cluster':
			return mdiCardBulletedOutline;
		case 'heuristic':
			return mdiChartBar;
		default:
			return mdiShapeCirclePlus;
	}
});

// Functions
async function getTransactionData() {
	if (props.identifier === '') {
		return;
	}

	entityData.value = null;
	try {
		const response = await dakar.data.blockchainTransactionsHashGet({hash: props.identifier});
		entityData.value = response.transactions;
		cacheStore.setValue(props.identifier, response.transactions);
	} catch (e) {
		setErrorMessage(e);
	}
}

async function getAddressData() {
	if (props.identifier === '') {
		return;
	}

	entityData.value = null;

	try {
		const response = await dakar.data.blockchainAddressesHashGet({hash: props.identifier});
		entityData.value = response.address;
		cacheStore.setValue(props.identifier, response.address);
	} catch (e) {
		setErrorMessage(e);
	}
}

async function getHeuristicData() {
	if (props.identifier === '') {
		return;
	}

	entityData.value = null;

	const tmp = {
		heuristicParameter: props.auxiliaryData.heuristicParameter,
		heuristicExcludeAddresses: props.auxiliaryData.heuristicExcludeAddresses,
		heuristicExcludeSpendingGaps: props.auxiliaryData.heuristicExcludeSpendingGaps,
		heuristicCustomClusters: props.auxiliaryData.heuristicClusterTypes?.length > 0,
		heuristicTypeTitle: props.auxiliaryData.displayType,
		clusterCount: props.auxiliaryData.heuristicClusterCount,
		heuristicUid: props.auxiliaryData.uid,
		heuristicTimestamp: new Date(props.auxiliaryData.heuristicTs),
		clusters: [],
	};

	// Check if data has to be loaded from backend
	if (!tmp.clusterCount) {
		entityData.value = tmp;
		return;
	}

	try {
		const response = await dakar.heuristic.heuristicDetailsPost({heuristic: {uid: props.identifier}});

		if (!response.heuristic) {
			throw new Error('response contains no heuristics');
		}

		tmp.clusters = response.heuristic.clusters;
		entityData.value = tmp;
		cacheStore.setValue(props.identifier, tmp);
	} catch (e) {
		setErrorMessage(e);
	}
}

function setErrorMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'error', temporary: true, category: route.name,
	});
}

async function downloadReport() {
	try {
		const response = await dakar.heuristic.heuristicsReportUidGet({uid: entityData.value.heuristicUid});
		// Looks hacky, but it is the only way with good UX
		const a = document.createElement('a');
		a.href = URL.createObjectURL(response);

		a.setAttribute(
			'download',
			`heuristic_report_${getCurrentDate()}_${entityData.value.heuristicUid}.csv`,
		);
		a.click();
		a.remove();
	} catch (e) {
		setErrorMessage(e);
	}
}

function emitDeleteEntity() {
	emit('deleteEntity', props.identifier);
	model.value = false;
}

function handleAddHeuristicClick() {
	emit('addHeuristic');
}

function handleRouteGuardDialogAdd() {
	emit('addNode', routeGuardId);
	routeGuardDialogModel.value = false;
}

</script>

<style scoped>
.shorten {
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
  margin-right: 2px;
}
</style>
