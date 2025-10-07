<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <side-bar
    v-model="model"
    title="Shortest Path"
    :icon="mdiChartTimelineVariant"
    max-width="648px"
    title-one-line
    disable-full-screen
  >
    <template #actions />
    <template #body>
      <v-card flat>
        <v-card-text>
          <div class="text-subtitle-1 mb-5">
            Find one of the shortest paths between two transactions. Multiple shortest paths can exist.
          </div>
          <div
            class="d-flex align-center flex-wrap"
            style="gap: 5px 20px"
          >
            <v-text-field
              :model-value="from"
              style="min-width: 200px"
              label="From"
              readonly
              autofocus
            />
            <v-text-field
              :model-value="to"
              style="min-width: 200px"
              label="To"
              readonly
            />
          </div>
          <div class="d-flex align-center flex-wrap">
            <v-radio-group
              v-model="anyDirection"
              inline
              label="Search direction:"
              :disabled="isLoading"
              hide-details
            >
              <v-radio
                label="Linear"
                :value="false"
              />
              <!-- eslint-disable vue/prefer-true-attribute-shorthand -->
              <v-radio
                label="Any"
                :value="true"
              />
            </v-radio-group>
            <v-checkbox
              v-model="includePrivacyTransactions"
              label="Traverse Classified Transactions"
              class="mx-5"
              :disabled="isLoading"
              hide-details
            />
            <v-btn
              class="ms-auto"
              color="primary"
              :loading="isLoading"
              @click="handleSearch"
            >
              Search
            </v-btn>
          </div>
          <v-divider
            v-if="resultTransactions.length > 0"
            class="my-3"
          />
          <v-timeline
            v-if="resultTransactions.length > 0"
            density="compact"
            side="end"
          >
            <v-timeline-item
              v-for="(tx) in resultTransactions"
              :key="tx.txhash"
              :dot-color="tx.txtype?'purple':'grey'"
              max-width="500px"
            >
              <template #opposite>
                <span
                  class="text-h5"
                  v-text="new Date(tx.bts).toLocaleString()"
                />
              </template>
              <transaction-item :tx="tx" />
            </v-timeline-item>
          </v-timeline>
        </v-card-text>
      </v-card>
    </template>
  </side-bar>
</template>

<script setup>
import SideBar from '@/components/common/SideBar.vue';
import {mdiChartTimelineVariant} from '@mdi/js';
import {getDakarClient, handleError} from '@/utilities';
import TransactionItem from '@/components/common/TransactionItem.vue';
import {onUpdated, ref} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const model = defineModel({type: Boolean});
const route = useRoute();
const msgStore = useMsgStore();
const context = {addMessage: msgStore.addMessage, $route: route};
const dakar = getDakarClient(route.params.blockchainMode);

const props = defineProps({
	from: {type: String, required: true},
	to: {type: String, required: true},
});

// V-model
const includePrivacyTransactions = ref(true);
const anyDirection = ref(false);
const isLoading = ref(false);
const resultTransactions = ref([]);

let oldFrom = '';
let oldTo = '';

// Hooks
onUpdated(() => {
	if (oldFrom === props.from && oldTo === props.to) {
		return;
	}

	oldFrom = props.from;
	oldTo = props.to;
	resultTransactions.value = [];
});

// Functions
function setInfoMessage(msg) {
	msgStore.addMessage({
		text: msg, type: 'info', temporary: true, category: route.name,
	});
}

async function handleSearch() {
	if (isLoading.value) {
		return;
	}

	msgStore.resetMessages();

	resultTransactions.value = [];
	await doLookup();
}

async function doLookup() {
	isLoading.value = true;

	try {
		const response = await dakar.tools.shortestTransactionPathPost({
			transactions: {
				to: props.to.trim(),
				from: props.from.trim(),
				includePrivacyTransactions: includePrivacyTransactions.value,
				anyDirection: anyDirection.value,
			},
		});

		if (response.msg) {
			setInfoMessage(response.msg);
		}

		if (response.transactions && response.transactions.length > 0) {
			if (props.from.trim() !== response.transactions[0].txhash) {
				response.transactions = response.transactions.reverse();
			}

			resultTransactions.value = response.transactions;
		}
	} catch (e) {
		handleError(context, e);
	}

	isLoading.value = false;
}

</script>

<style scoped>
.textBorder {
  border-style: solid;
  border-radius: 5px;
  border-width: 1px;
}
</style>
