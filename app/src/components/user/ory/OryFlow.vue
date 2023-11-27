<template>
  <div v-if="embed">
    <v-tabs
      v-model="tab"
      align-tabs="center"
    >
      <v-tab
        v-for="(formNodes, i) in getForms"
        :key="`${formId}_${i}`"
        :value="`${formId}_${i}`"
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
  </div>
  <div v-else>
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
  </div>
</template>

<script setup>
import OryUiNode from './OryUiNode.vue';
import {getNodeName} from '@/components/user/ory/utils';
import {computed, onMounted, ref, watch} from 'vue';
import {useStore} from 'vuex';
import {useRoute} from 'vue-router';

const store = useStore();
const route = useRoute();

const props = defineProps({
	flow: {type: Object, required: true},
	formId: {type: String, required: true},
	embed: {type: Boolean, required: false, default: false},
	// DisabledForms is an array of formIDs for which submitting is disabled
	disabledForms: {type: Array, require: false, default: () => []},
});

const emit = defineEmits(['submit']);

const groupTitles = new Map([
	['totp', 'Two-Factor Authentication'],
	['password', 'Password'],
	['profile', 'Profile'],
]);
const tab = ref(null);

watch(props.flow, () => {
	displayMessages();
});

onMounted(() => {
	displayMessages();
});

// Computed
// GetForms returns an array of node sets ([[node1, node2, ...],[node10, node11, ...]]).
// This is needed because the initial set of nodes contained in the flow property can have
// more than one group. Nodes of the default group (e.g. csrf tokens) are included in
// each returned set.
const getForms = computed(() => {
	const forms = [];

	if (!props.flow || !props.flow.ui || !props.flow.ui.nodes) {
		return forms;
	}

	// Find unique group names
	const groupNames = new Set();
	props.flow.ui.nodes.forEach(e => groupNames.add(e.group));

	groupNames.forEach(e => {
		if (e !== 'default') {
			const formNodes = props.flow.ui.nodes.filter(d => d.group === 'default' || d.group === e);
			forms.push(formNodes);
		}
	});
	return forms;
});

// Functions
function setMessage(msg, msgType) {
	store.dispatch('addMessage', {text: msg, type: msgType, temporary: false, category: route.name});
}

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

function displayMessages() {
	if (!props.flow.ui || !props.flow.ui.messages) {
		return;
	}

	props.flow.ui.messages.forEach(msg => setMessage(msg.text, msg.type));
}

</script>

<style scoped>

</style>
