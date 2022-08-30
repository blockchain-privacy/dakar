<template>
  <div>
    <template v-if="flow.ui.messages" >
      <ory-ui-message v-for="(msg,i) in flow.ui.messages" :key="i" :message="msg"/>
    </template>
    <v-form :id="formId" :action="flow.ui.action" :method="flow.ui.method">
      <ory-ui-node
          v-for="node in flow.ui.nodes"
          :key="getNodeId(node)"
          :id="getNodeId(node)"
          :node="node"
          @submit="propagateSubmitEvent"
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
  methods: {
    getNodeId,
    propagateSubmitEvent() {
      this.$emit('submit');
    },
  },
};
</script>

<style scoped>

</style>
