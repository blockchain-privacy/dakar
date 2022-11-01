<template>
  <div class="msgBox">
    <transition-group name="component-fade" mode="out-in">
      <!-- todo Vue3 supports iterating over Maps with v-for
            https://github.com/vuejs/vue/issues/6644 -->
      <Messages
          v-for="msg in messages" :key="msg.key"
          @destructed="removeMessage(msg.key)"
          :type="msg.type" :temporary="msg.temporary">
        {{ msg.text }}
      </Messages>
    </transition-group>
  </div>
</template>

<script>
import Messages from './Messages.vue';

export default {
  name: 'MsgBox',
  components: { Messages },
  computed: {
    messages() {
      return this.$store.getters.getMessages;
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
