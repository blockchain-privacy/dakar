<template>
  <div>
    <template v-if="flow.ui.messages" >
      <ory-ui-message v-for="(msg,i) in flow.ui.messages" :key="i" :message="msg"/>
    </template>
    <v-form v-for="(formNodes,i) in getForms" :key="`${formId}_${i}`"
            :id="`${formId}_${i}`" :action="flow.ui.action" :method="flow.ui.method">
      <ory-ui-node
          v-for="node in formNodes"
          :key="getNodeId(node)"
          :id="getNodeId(node)"
          :node="node"
          @submit="propagateSubmitEvent(`${formId}_${i}`)"
      />
    </v-form>
  </div>
</template>

<script>
import { getNodeId } from '@ory/integrations/ui';
import OryUiNode from './OryUiNode.vue';
import OryUiMessage from './OryUiMessage.vue';

export default {
  name: 'OryFlow',
  components: { OryUiMessage, OryUiNode },
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
    propagateSubmitEvent(formID) {
      this.$emit('submit', formID);
    },
  },
};
</script>

<style scoped>

</style>
