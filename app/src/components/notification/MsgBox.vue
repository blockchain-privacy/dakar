<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="msgBox">
    <transition-group
      name="slide-x"
      mode="out-in"
      tag="span"
    >
      <messages
        v-for="msg in allMessages"
        :key="msg.key"
        class="mt-3"
        :type="msg.value.type"
        :temporary="msg.value.temporary"
        :text="msg.value.text"
        closable
        @destructed="removeMessage(msg.key)"
      />
      <messages
        v-if="numHiddenMessages > 0"
        key="hidden_message_display"
        type="info"
        class="mt-3"
        :temporary="false"
        :text="hiddenMessageText"
      />
    </transition-group>
  </div>
</template>

<script setup>
import Messages from './Messages.vue';
import {plural} from '@/utilities';
import {computed} from 'vue';
import {useMsgStore} from '@/pinia/msg';

const msgStore = useMsgStore();

const maxNumberOfMessagesToDisplay = 3;

// Computed
const allMessages = computed(() => {
	// Show only limited number of messages
	const messages = [];

	const mapMessages = msgStore.getMessages;
	for (const [key, value] of mapMessages) {
		messages.push({key, value});
		if (messages.length + 1 > maxNumberOfMessagesToDisplay) {
			break;
		}
	}

	return messages;
});

const numHiddenMessages = computed(() => msgStore.getMessages.size - maxNumberOfMessagesToDisplay);
const hiddenMessageText = computed(() => {
	if (numHiddenMessages.value < 1) {
		return '';
	}

	return `${numHiddenMessages.value} additional ${plural('message', numHiddenMessages.value)}`;
});
function removeMessage(key) {
	msgStore.removeMessage(key);
}

</script>

<style scoped>

.msgBox {
  z-index: 1006;
  position: absolute;
  right: 5px;
  top: 5px;
}

.slide-x-enter-active,
.slide-x-leave-active {
  transition: all 0.25s ease-out;
}

.slide-x-enter-from {
  opacity: 0;
  transform: translateX(50px);
}

.slide-x-leave-to {
  opacity: 0;
  transform: translateX(50px);
}

</style>
