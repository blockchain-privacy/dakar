<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <side-bar
    v-model="model"
    title="Shortest Path"
    :icon="mdiChartTimelineVariant"
    max-width="700px"
    title-one-line
    disable-full-screen
  >
    <template #body>
      <v-card flat>
        <v-card-text>
          <div class="text-body-large mb-5">
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
              @click="doLookup"
            >
              Search
            </v-btn>
          </div>
          <v-divider
            v-if="resultTransactions.length > 0"
            class="my-3"
          />
          <alert
            :text="errorMsg"
            :type="alertType"
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
              max-width="600px"
            >
              <template #opposite>
                <span
                  class="text-headline-small"
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
import {mdiChartTimelineVariant} from '@mdi/js';
import {onUpdated, ref} from 'vue';
import {useRoute} from 'vue-router';
import SideBar from '@/components/common/SideBar.vue';
import {getDakarClient} from '@/utilities';
import TransactionItem from '@/components/common/TransactionItem.vue';
import Alert from '@/components/common/Alert.vue';

const model = defineModel({type: Boolean});
const route = useRoute();
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
const errorMsg = ref('');
const alertType = ref('info');

let oldFrom = '';
let oldTo = '';

// Hooks
onUpdated(() => {
	if (oldFrom === props.from && oldTo === props.to) {
		return;
	}

	errorMsg.value = '';
	oldFrom = props.from;
	oldTo = props.to;
	resultTransactions.value = [];
});

// Functions
function setInfoMessage(msg) {
	alertType.value = 'info';
	errorMsg.value = msg;
}

function setErrorMessage(msg) {
	alertType.value = 'error';
	errorMsg.value = msg;
}

async function doLookup() {
	if (isLoading.value) {
		return;
	}

	resultTransactions.value = [];
	isLoading.value = true;
	errorMsg.value = '';

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
				response.transactions = response.transactions.toReversed();
			}

			resultTransactions.value = response.transactions;
		}
	} catch (error) {
		setErrorMessage(error.message);
	}

	isLoading.value = false;
}

</script>

<style scoped>

</style>
