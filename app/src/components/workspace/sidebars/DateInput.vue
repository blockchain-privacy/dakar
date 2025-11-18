<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-menu
    v-model="menuModel"
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
    <v-date-picker
      v-model="model"
      hide-header
      :allowed-dates="isDateAllowed"
      show-adjacent-months
      :first-day-of-week="firstDayOfWeek"
    >
      <template #actions>
        <v-btn
          color="primary"
          @click="menuModel = false"
        >
          OK
        </v-btn>
      </template>
    </v-date-picker>
  </v-menu>
</template>
<script setup>
// DateInput was implemented because v-date-input (by vuetify) has usability issues:
// - it allows modifying the formatted text, but does not update the selected date
// - the input field can not set to be readonly, while still allowing the date picker to work
// Thus, this component allows selecting a date via the date picker. The selected date
// is displayed as formatted text in the readonly text field.
import {computed, ref} from 'vue';

defineProps({
	label: {type: String, required: false, default: 'Date'},
	rules: {type: Array, required: false, default: undefined},
	error: {type: Boolean, required: false},
});

// In 2009 Bitcoin was created
const earliestDate = new Date(2009, 0);
const latestDate = new Date();

function isDateAllowed(someDate) {
	return someDate <= latestDate && someDate >= earliestDate;
}

const model = defineModel({type: Date});
const menuModel = ref(false);

// Computed

// Firefox does not support getWeekInfo() yet.
// see: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Intl/Locale/getWeekInfo
// Default to Monday as the first day of the week if it could not be determined.
// eslint-disable-next-line no-warning-comments
// Todo: simplify this when firefox adds support.
const firstDayOfWeek = computed(() => new Intl.Locale(navigator.language)?.getWeekInfo?.().firstDay || 1);

</script>
<style scoped>

</style>
