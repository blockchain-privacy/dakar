<template>
  <v-menu
    v-model="menuModel"
    eager
    :close-on-content-click="false"
  >
    <template #activator="a">
      <v-text-field
        v-bind="a.props"
        readonly
        :label="label"
        :model-value="model?.toLocaleDateString()"
        hide-details
        class="me-2"
        :rules="rules"
        :error="error"
      />
    </template>
    <v-card>
      <!-- todo: to determine the first day of the week locale.weekInfo or getWeekInfo() can be used. Firefox does not support this yet
        (see: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Intl/Locale/getWeekInfo) -->
      <v-date-picker
        v-model="model"
        hide-header
        :allowed-dates="isDateAllowed"
        show-adjacent-months
        :first-day-of-week="1"
      />
      <v-card-actions>
        <v-btn
          class="ms-auto"
          variant="text"
          @click="menuModel = false"
        >
          OK
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>
<script setup>
// DateInput was implemented because v-date-input (by vuetify) has usability issues:
// - it allows modifying the formatted text, but does not update the selected date
// - the input field can not set to be readonly, while still allowing the date picker to work
// Thus, this component allows selecting a date via the date picker. The selected date
// is displayed as formatted text in the readonly text field.
import {ref} from 'vue';

defineProps({
	label: {type: String, required: false, default: 'Date'},
	rules: {type: Array, required: false, default: undefined},
	error: {type: Boolean, required: false, default: false},
});

// In 2009 Bitcoin was created
const earliestDate = new Date(2009, 0);
const latestDate = new Date();

function isDateAllowed(someDate) {
	return someDate <= latestDate && someDate >= earliestDate;
}

const model = defineModel({type: Date});
const menuModel = ref(false);

</script>
<style scoped>

</style>
