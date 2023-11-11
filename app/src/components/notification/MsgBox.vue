<template>
  <div class="msgBox">
    <transition-group
      name="slide-x"
      mode="out-in"
      tag="span"
    >
      <Messages
        v-for="msg in messages"
        :key="msg.key"
        class="mt-3"
        :type="msg.value.type"
        :temporary="msg.value.temporary"
        :text="msg.value.text"
        @destructed="removeMessage(msg.key)"
      />
      <Messages
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

<script>
import Messages from './Messages.vue';
import {plural} from '@/utilities';

export default {
	name: 'MsgBox',
	components: {Messages},
	data() {
		return {
			maxNumberOfMessagesToDisplay: 3,
		};
	},
	computed: {
		messages() {
			// Show only limited number of messages
			const messages = [];
			const mapMessages = this.$store.getters.getMessages;
			for (const [key, value] of mapMessages) {
				messages.push({key, value});
				if (messages.length + 1 > this.maxNumberOfMessagesToDisplay) {
					break;
				}
			}

			return messages;
		},
		numHiddenMessages() {
			return this.$store.getters.getMessages.size - this.maxNumberOfMessagesToDisplay;
		},
		hiddenMessageText() {
			if (this.numHiddenMessages < 1) {
				return '';
			}

			return `${this.numHiddenMessages} additional ${plural('message', this.numHiddenMessages)}`;
		},
	},
	methods: {
		removeMessage(key) {
			this.$store.dispatch('removeMessage', key);
		},
	},
};
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
