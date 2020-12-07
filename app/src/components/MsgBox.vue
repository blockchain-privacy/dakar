<template>
  <v-layout row>
    <v-flex style="z-index: 100">
      <v-alert xs6 :value="infoMsg && infoMsg !== ''" type="info"
               dense dismissible v-model="isInfoActive">
        {{ infoMsg }}
      </v-alert>
      <v-alert xs6 :value="errorMsg && errorMsg !== ''" type="error"
               dense dismissible v-model="isErrorActive">
        {{ errorMsg }}
      </v-alert>
      <v-alert xs6 :value="successMsg && successMsg !== ''" type="success"
               dense dismissible v-model="isSuccessActive">
        {{ successMsg }}
      </v-alert>
      <v-alert xs6 :value="warningMsg && warningMsg !== ''" type="warning"
               dense dismissible v-model="isWarningActive">
        {{ warningMsg }}
      </v-alert>
    </v-flex>
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
};
</script>

<style scoped>

</style>
