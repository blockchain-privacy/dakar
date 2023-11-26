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
        @destructed="removeMessage(msg.key)"
      />
      <messages
        v-if="numHiddenMessages > 0"
        key="hidden_message_display"
        type="info"
        class="mt-3"
        :temporary="false"
        :closable="false"
        :text="hiddenMessageText"
      />
    </transition-group>
  </div>
</template>

<script setup>
import Messages from './Messages.vue';
import {plural} from '@/utilities';
import {computed} from 'vue';
import {useStore} from 'vuex';

const store = useStore();

const 	maxNumberOfMessagesToDisplay = 3;

// Computed
const allMessages = computed(() => {
	// Show only limited number of messages
	const messages = [];
	const mapMessages = store.getters.getMessages;
	for (const [key, value] of mapMessages) {
		messages.push({key, value});
		if (messages.length + 1 > maxNumberOfMessagesToDisplay) {
			break;
		}
	}

	return messages;
});

const numHiddenMessages = computed(() => store.getters.getMessages.size - maxNumberOfMessagesToDisplay);
const hiddenMessageText = computed(() => {
	if (numHiddenMessages.value < 1) {
		return '';
	}

	return `${numHiddenMessages.value} additional ${plural('message', numHiddenMessages)}`;
});
function removeMessage(key) {
	store.dispatch('removeMessage', key);
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
