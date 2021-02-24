<template>
  <v-layout>
    <div class="msgBox">
      <transition-group name="component-fade" mode="out-in">
        <Messages
            v-for="(msg, i) in messages" :key="msg.text + i"
            :type="msg.type" :temporary="msg.temporary">
          {{ msg.text }}
        </Messages>
      </transition-group>
    </div>
  </v-layout>
</template>

<script>
import Messages from './Messages.vue';

export default {
  name: 'MsgBox',
  components: { Messages },
  computed: {
    messages: {
      get() {
        return this.$store.getters.getMessages;
      },
      set(value) {
        this.$store.dispatch('addMessage', value);
      },
    },
  },
};
</script>

<style scoped>

.msgBox {
  z-index: 100;
  position: absolute;
  right: 5px;
  top: 5px;
}

.component-fade-enter-active, .component-fade-leave-active {
  transition: opacity 0.2s ease;
}

.component-fade-enter, .component-fade-leave-to {
  opacity: 0;
}

</style>
