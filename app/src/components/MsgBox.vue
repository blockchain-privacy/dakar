<template>
  <v-layout>
    <div class="msgBox">
      <v-alert xs6 :value="infoMsg && infoMsg !== ''" type="info"
               class="floating" max-width="400px" prominent outlined text
               dismissible v-model="isInfoActive">
        {{ infoMsg }}
      </v-alert>

      <v-alert xs6 :value="errorMsg && errorMsg !== '' &&  isValidError(errorMsg)" type="error"
               class="floating" max-width="400px" prominent outlined text border="left"
               dismissible v-model="isErrorActive">
        {{ errorMsg }}
      </v-alert>
      <v-alert xs6 :value="successMsg && successMsg !== ''" type="success"
               class="floating" max-width="400px" prominent outlined text border="left"
               dismissible v-model="isSuccessActive">
        {{ successMsg }}
      </v-alert>
      <v-alert xs6 :value="warningMsg && warningMsg !== ''" type="warning"
               class="floating" max-width="400px" prominent outlined text border="left"
               dismissible v-model="isWarningActive">
        {{ warningMsg }}
      </v-alert>
    </div>
  </v-layout>
</template>

<script>
export default {
  name: 'MsgBox',
  computed: {
    errorMsg() {
      return this.$store.getters.getErrorMsg;
    },
    infoMsg() {
      return this.$store.getters.getInfoMsg;
    },
    successMsg() {
      return this.$store.getters.getSuccessMsg;
    },
    warningMsg() {
      return this.$store.getters.getWarningMsg;
    },
    isErrorActive: {
      get() {
        return this.$store.getters.isErrorActive;
      },
      set(value) {
        this.$store.dispatch('setErrorActive', value);
      },
    },
    isInfoActive: {
      get() {
        return this.$store.getters.isInfoActive;
      },
      set(value) {
        this.$store.dispatch('setInfoActive', value);
      },
    },
    isSuccessActive: {
      get() {
        return this.$store.getters.isSuccessActive;
      },
      set(value) {
        this.$store.dispatch('setSuccessActive', value);
      },
    },
    isWarningActive: {
      get() {
        return this.$store.getters.isWarningActive;
      },
      set(value) {
        this.$store.dispatch('setWarningActive', value);
      },
    },
  },
  methods: {
    isValidError(err) {
      return err.toString() !== '';
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

.floating {
  background-color: white !important;
}

/* >>> for deep selection, 0 height and width for removing
the border which 'prominent' introduces */
>>> .v-alert__icon {
  height: 0 !important;
  width: 0 !important;
}

</style>
