<template>
  <v-layout>
    <div class="msgBox">
      <Messages
          v-for="(msg, i) in messages" :key="i"
          :type="msg.type" :temporary="msg.temporary">
        {{ msg.text }}
      </Messages>
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

</style>
