<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div>
    <ory-ui-node-input
      v-if="isUiNodeInputAttributes(node.attributes)"
      :name="name"
      :submit-enabled="submitEnabled"
      :attributes="node.attributes"
      :meta="node.meta"
      :messages="node.messages"
      @submit="propagateSubmitEvent"
    />
    <v-img
      v-else-if="isUiNodeImageAttributes(node.attributes)"
      class="ma-3"
      :src="node.attributes.src"
      :width="node.attributes.witdh"
      :height="node.attributes.height"
      :alt="node.meta.label.text"
    />
    <a
      v-else-if="isUiNodeAnchorAttributes(node.attributes)"
      :href="node.attributes.href"
    >
      {{ node.attributes.title.text }}
    </a>
    <component
      :is="'script'"
      v-else-if="isUiNodeScriptAttributes(node.attributes)"
      :src="node.attributes.src"
      type="application/javascript"
      :integrity="node.attributes.integrity"
      :referrerpolicy="node.attributes.referrerpolicy"
      :crossorigin="node.attributes.crossorigin"
    />
    <template v-else-if="isUiNodeTextAttributes(node.attributes)">
      <p
        v-if="node.meta?.label?.text"
        class="text-center text-subtitle-1"
      >
        {{ `${node.meta.label.text}` }}
      </p>
      <p
        v-if="node.attributes?.text?.text"
        class="text-center text-subtitle-1"
      >
        {{ node.attributes.text.text }}
      </p>
    </template>
    <v-btn v-else-if="node.type === 'submit'" />
    <template v-if="node.messages && !isUiNodeInputAttributes(node.attributes)">
      <ory-ui-message
        v-for="(msg,i) in node.messages"
        :key="i"
        :message="msg"
      />
    </template>
  </div>
</template>

<script setup>
import {
	isUiNodeInputAttributes,
	isUiNodeImageAttributes,
	isUiNodeAnchorAttributes,
	isUiNodeScriptAttributes,
	isUiNodeTextAttributes,
} from '@/components/user/ory/utils';
import OryUiNodeInput from './OryUiNodeInput.vue';
import OryUiMessage from './OryUiMessage.vue';

defineProps({
	node: {type: Object, required: true},
	name: {type: String, required: true},
	submitEnabled: {type: Boolean, require: false},
});

const emit = defineEmits(['submit']);

function propagateSubmitEvent(btnName) {
	emit('submit', btnName);
}

</script>

<style scoped>

</style>
