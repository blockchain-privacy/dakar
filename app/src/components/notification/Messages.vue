<template>
  <v-alert
      xs6 :type="type" width="300px"
      :class="`msg ${$vuetify.theme.dark?'dark-msg':'light-msg'}`"
      prominent outlined border="left"
      @mouseenter="stopTimer"
      @mouseleave="startTimer"
      @input="hideMessage"
      dismissible v-model="showMessage">
    <v-progress-linear v-if="temporary" absolute bottom :value="progressValue" :color="type"/>
    <slot/>
  </v-alert>
</template>

<script>
export default {
  name: 'Messages',
  props: {
    type: { type: String, required: true },
    temporary: { type: Boolean, default: false },
  },
  data() {
    return {
      showMessage: true,
      progressValue: 100,
      interval: null,
    };
  },
  methods: {
    hideMessage() {
      if (this.interval) clearInterval(this.interval);
      this.showMessage = false;
      this.$emit('destructed');
    },
    startTimer() {
      if (!this.temporary) return;
      if (this.interval) clearInterval(this.interval);
      this.startProgressLoop();
    },
    stopTimer() {
      if (!this.temporary || !this.interval) return;
      this.progressValue = 100;
      clearInterval(this.interval);
    },
    startProgressLoop() {
      // 15 seconds and 10 steps: 15000 / 10
      const timeout = 150;
      this.interval = setInterval(() => {
        if (this.progressValue === 0) {
          this.hideMessage();
          return;
        }
        this.progressValue -= 1;
      }, timeout);
    },
  },
  mounted() {
    this.startTimer();
  },
};
</script>

<style scoped>

.msg {
  word-break: break-word;
  box-shadow: 0 2px 4px -1px rgba(0, 0, 0, 0.2),
  0 4px 5px 0 rgba(0, 0, 0, 0.14),
  0 1px 10px 0 rgba(0, 0, 0, 0.12);
}

.dark-msg {
  background-color: #1E1E1E !important;
}

.light-msg {
  background-color: white !important;
}

/* >>> for deep selection, 0 height and width for removing
the border which 'prominent' introduces */
>>> .v-alert__icon {
  height: 0 !important;
  width: 0 !important;
}
</style>
