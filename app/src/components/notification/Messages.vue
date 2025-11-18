<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-alert
    v-model="showMessage"
    style="background-color: rgb(var(--v-theme-surface))"
    density="compact"
    :type="type"
    width="300px"
    class="msg"
    variant="outlined"
    border="start"
    :closable="closable"
    :text="text"
    @mouseenter="stopTimer"
    @mouseleave="startTimer"
    @update:model-value="hideMessage"
  >
    <v-progress-linear
      v-if="temporary"
      style="bottom:0;top:unset"
      :model-value="progressValue"
      :color="type"
      absolute
    />
  </v-alert>
</template>

<script setup>
import {onBeforeUnmount, onMounted, ref} from 'vue';

const props = defineProps({
	type: {type: String, required: true},
	temporary: {type: Boolean, required: false},
	text: {type: String, required: true},
	closable: {type: Boolean, required: false},
});

const emit = defineEmits(['destructed']);
const showMessage = ref(true);
const progressValue = ref(100);
const interval = ref(null);

function hideMessage() {
	if (interval.value) {
		clearInterval(interval.value);
	}

	showMessage.value = false;
	emit('destructed');
}

function startTimer() {
	if (!props.temporary) {
		return;
	}

	if (interval.value) {
		clearInterval(interval.value);
	}

	startProgressLoop();
}

function stopTimer() {
	if (!props.temporary || !interval.value) {
		return;
	}

	progressValue.value = 100;
	clearInterval(interval.value);
}

function startProgressLoop() {
	// 15 seconds and 10 steps: 15000 / 10
	const timeout = 150;
	interval.value = setInterval(() => {
		if (progressValue.value === 0) {
			hideMessage();
			return;
		}

		progressValue.value -= 1;
	}, timeout);
}

// Hooks
onMounted(() => {
	startTimer();
});

onBeforeUnmount(() => {
	clearInterval(interval.value);
});

</script>

<style scoped>

.msg {
  word-break: break-word;
  box-shadow: 0 2px 4px -1px rgba(0, 0, 0, 0.2),
  0 4px 5px 0 rgba(0, 0, 0, 0.14),
  0 1px 10px 0 rgba(0, 0, 0, 0.12);
}

</style>
