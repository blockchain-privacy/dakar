<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div>
    <v-expand-transition>
      <template v-if="flow.ui?.messages">
        <div>
          <ory-ui-message
            v-for="(msg,i) in flow.ui.messages"
            :key="i"
            :message="msg"
          />
        </div>
      </template>
    </v-expand-transition>
    <template v-if="embed">
      <v-tabs
        v-model="tab"
        align-tabs="center"
      >
        <v-tab
          v-for="(formNodes, i) in getForms"
          :key="`${formId}_${i}`"
          :value="`${formId}_${i}`"
          :to="{ name: route.name, params: { tabName: getFormGroupName(formNodes) }, query: route.query}"
        >
          {{ groupTitles.get(getFormGroupName(formNodes)) }}
        </v-tab>
      </v-tabs>
      <v-card
        class="mx-auto"
        variant="text"
      >
        <v-card-text>
          <v-window v-model="tab">
            <v-window-item
              v-for="(formNodes,i) in getForms"
              :key="`${formId}_${i}`"
              :value="`${formId}_${i}`"
            >
              <v-form
                :id="`${formId}_${i}`"
                :action="flow.ui.action"
                :method="flow.ui.method"
              >
                <ory-ui-node
                  v-for="(node, y) in formNodes"
                  :key="y"
                  :name="getNodeName(node)"
                  :node="node"
                  :submit-enabled="!disabledForms.includes(`${formId}_${i}`)"
                  @submit="propagateSubmitEvent(`${formId}_${i}`)"
                />
              </v-form>
            </v-window-item>
          </v-window>
        </v-card-text>
      </v-card>
    </template>
    <template v-else>
      <div
        v-for="(formNodes,i) in getForms"
        :key="`${formId}_${i}`"
      >
        <v-form
          :id="`${formId}_${i}`"
          :action="flow.ui.action"
          :method="flow.ui.method"
          autocomplete="on"
        >
          <ory-ui-node
            v-for="(node, y) in formNodes"
            :key="y"
            :name="getNodeName(node)"
            :node="node"
            :submit-enabled="!disabledForms.includes(`${formId}_${i}`)"
            @submit="(btnName) => propagateSubmitEvent(`${formId}_${i}`,btnName)"
          />
        </v-form>
        <v-divider
          v-if="getForms.length > 1 && i +1 < getForms.length"
          class="my-5"
        />
      </div>
    </template>
  </div>
</template>

<script setup>
import OryUiNode from './OryUiNode.vue';
import {getNodeName} from '@/components/user/ory/utils';
import {computed, ref} from 'vue';
import {useRoute} from 'vue-router';
import OryUiMessage from '@/components/user/ory/OryUiMessage.vue';

const route = useRoute();

const props = defineProps({
	flow: {type: Object, required: true},
	formId: {type: String, required: true},
	embed: {type: Boolean, required: false},
	// DisabledForms is an array of formIDs for which submitting is disabled
	disabledForms: {type: Array, require: false, default: () => []},
});

const emit = defineEmits(['submit']);

const groupTitles = new Map([
	['totp', 'Two-Factor Authentication'],
	['password', 'Password'],
	['profile', 'Profile'],
	['passkey', 'Passkey'],
	['webauthn', 'Webauthn'],
]);
const tab = ref(null);

// Computed
// getFormGroupNames returns a unique array including all form group names except the 'default' group
const getFormGroupNames = computed(() => {
	if (!props.flow || !props.flow.ui || !props.flow.ui.nodes) {
		return [];
	}

	// Find unique group names
	const groupNames = new Set();
	props.flow.ui.nodes.forEach(e => {
		if (e.group !== 'default') {
			groupNames.add(e.group);
		}
	});
	return Array.from(groupNames);
});

// GetForms returns an array of node sets ([[node1, node2, ...],[node10, node11, ...]]).
// This is needed because the initial set of nodes contained in the flow property can have
// more than one group. Nodes of the default group (e.g. csrf tokens) are included in
// each returned set.
const getForms = computed(() => {
	const forms = [];

	getFormGroupNames.value.forEach(e => {
		if (e !== 'default') {
			const formNodes = props.flow.ui.nodes.filter(d => d.group === 'default' || d.group === e);
			forms.push(formNodes);
		}
	});
	return forms;
});

// Functions
function propagateSubmitEvent(formID, btnName) {
	emit('submit', formID, btnName);
}

function getFormGroupName(formNodes) {
	const nodes = formNodes.filter(d => d.group !== 'default');
	if (nodes) {
		return nodes[0].group;
	}

	return '';
}

</script>

<style scoped>

</style>
