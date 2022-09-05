<template>
  <div>
    <div v-for="(formNodes,i) in getForms" :key="`${formId}_${i}`">
      <v-form
          :id="`${formId}_${i}`" :action="flow.ui.action" :method="flow.ui.method">
        <ory-ui-node
            v-for="node in formNodes"
            :key="getNodeId(node)"
            :id="getNodeId(node)"
            :node="node"
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
  methods: {
    getNodeId,
    setMessage(msg, msgType) {
      this.$store.dispatch('addMessage', { text: msg, type: msgType, temporary: true });
    },
    propagateSubmitEvent(formID) {
      this.$emit('submit', formID);
    },
    displayMessages() {
      if (!this.flow.ui || !this.flow.ui.messages) return;
      this.flow.ui.messages.forEach((msg) => this.setMessage(msg.text, msg.type));
    },
  },
  updated() {
    this.displayMessages();
  },
  mounted() {
    this.displayMessages();
  },
};
</script>

<style scoped>

</style>
