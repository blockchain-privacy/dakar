<template>
  <div>
    <ory-ui-node-input
        v-if="isUiNodeInputAttributes(node.attributes)"
        :attributes="node.attributes"
        :meta="node.meta"
        :id="id" @submit="propagateSubmitEvent"/>
    <v-img v-else-if="isUiNodeImageAttributes(node.attributes)"
           :src="node.attributes.src"
           :id="id"
           :width="node.attributes.witdh"
           :height="node.attributes.height"
    />
    <a v-else-if="isUiNodeAnchorAttributes(node.attributes)" :href="node.attributes.href">
      {{ node.attributes.title.text }}
    </a>
    <script
        v-else-if="isUiNodeScriptAttributes(node.attributes)"
        :src="node.attributes.src"
        :type="node.attributes.type"
        :integrity="node.attributes.integrity"
        :referrerpolicy="node.attributes.referrerpolicy"
        :crossorigin="node.attributes.crossorigin"
    ></script>
    <p v-else-if="isUiNodeTextAttributes(node.attributes)">{{ node.attributes.text.text }}</p>

    <v-btn v-else-if="node.type === 'submit'"></v-btn>
    <template v-if="node.messages">
      <ory-ui-message v-for="(msg,i) in node.messages" :key="i" :message="msg"/>
    </template>
  </div>
</template>

<script>
import {
  isUiNodeInputAttributes,
  isUiNodeImageAttributes,
  isUiNodeAnchorAttributes,
  isUiNodeScriptAttributes,
  isUiNodeTextAttributes,
} from '@ory/integrations/ui';
import OryUiNodeInput from './OryUiNodeInput.vue';
import OryUiMessage from './OryUiMessage.vue';

export default {
  name: 'OryUiNode',
  components: { OryUiMessage, OryUiNodeInput },
  props: {
    node: { type: Object, required: true },
    id: { type: String, required: true },
  },
  methods: {
    isUiNodeInputAttributes,
    isUiNodeImageAttributes,
    isUiNodeAnchorAttributes,
    isUiNodeScriptAttributes,
    isUiNodeTextAttributes,
    propagateSubmitEvent() {
      this.$emit('submit');
    },
  },
};
</script>

<style scoped>

</style>
