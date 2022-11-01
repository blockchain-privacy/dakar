<template>
  <div v-if="embed">
    <v-card class="mx-auto elevation-4 my-5" max-width="700"
            v-for="(formNodes,i) in getForms" :key="`${formId}_${i}`">
      <v-toolbar color="primary" dark flat>
        <v-toolbar-title>
          {{ groupTitles.get(getFormGroupName(formNodes)) }}
        </v-toolbar-title>
      </v-toolbar>
      <v-card>
        <v-card-text>
          <v-form
              :id="`${formId}_${i}`" :action="flow.ui.action" :method="flow.ui.method">
            <ory-ui-node
                v-for="node in formNodes"
                :key="getNodeId(node)"
                :id="getNodeId(node)"
                :node="node"
                :submit-enabled="!disabledForms.includes(`${formId}_${i}`)"
                @submit="propagateSubmitEvent(`${formId}_${i}`)"
            />
          </v-form>
        </v-card-text>
      </v-card>
    </v-card>
  </div>
  <div v-else>
    <div v-for="(formNodes,i) in getForms" :key="`${formId}_${i}`">
      <v-form
          :id="`${formId}_${i}`" :action="flow.ui.action" :method="flow.ui.method">
        <ory-ui-node
            v-for="node in formNodes"
            :key="getNodeId(node)"
            :id="getNodeId(node)"
            :node="node"
            :submit-enabled="!disabledForms.includes(`${formId}_${i}`)"
            @submit="propagateSubmitEvent(`${formId}_${i}`)"
        />
      </v-form>
      <v-divider v-if="getForms.length > 1 && i +1 < getForms.length" class="my-5"></v-divider>
    </div>
  </div>
</template>

<script>
import { getNodeId } from '@ory/integrations/ui';
import OryUiNode from './OryUiNode.vue';

export default {
  name: 'OryFlow',
  components: { OryUiNode },
  props: {
    flow: { type: Object, required: true },
    formId: { type: String, required: true },
    embed: { type: Boolean, required: false, default: false },
    // disabledForms is an array of formIDs for which submitting is disabled
    disabledForms: { type: Array, require: false, default: () => [] },
  },
  computed: {
    // getForms returns an array of node sets ([[node1, node2, ...],[node10, node11, ...]]).
    // This is needed because the initial set of nodes contained in the flow property can have
    // more than one group. Nodes of the default group (e.g. csrf tokens) are included in
    // each returned set.
    getForms() {
      const forms = [];

      if (!this.flow || !this.flow.ui || !this.flow.ui.nodes) return forms;

      // find unique group names
      const groupNames = new Set();
      this.flow.ui.nodes.forEach((e) => groupNames.add(e.group));

      groupNames.forEach((e) => {
        if (e !== 'default') {
          const formNodes = this.flow.ui.nodes.filter((d) => d.group === 'default' || d.group === e);
          forms.push(formNodes);
        }
      });
      return forms;
    },
  },
  data() {
    return {
      groupTitles: new Map([
        ['totp', 'Two-Factor Authentication'],
        ['password', 'Change Password'],
        ['profile', 'Change Profile'],
      ]),
    };
  },
  methods: {
    getNodeId,
    setMessage(msg, msgType) {
      this.$store.dispatch('addMessage', { text: msg, type: msgType, temporary: true });
    },
    propagateSubmitEvent(formID) {
      this.$emit('submit', formID);
    },
    getFormGroupName(formNodes) {
      const nodes = formNodes.filter((d) => d.group !== 'default');
      if (nodes) {
        return nodes[0].group;
      }

      return '';
    },
    displayMessages() {
      if (!this.flow.ui || !this.flow.ui.messages) return;
      this.flow.ui.messages.forEach((msg) => this.setMessage(msg.text, msg.type));
    },
  },
  mounted() {
    this.displayMessages();
  },
  watch: {
    flow() {
      this.displayMessages();
    },
  },
};
</script>

<style scoped>

</style>
