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
      :absolute="true"
    />
  </v-alert>
</template>

<script>
export default {
	name: 'Messages',
	props: {
		type: {type: String, required: true},
		temporary: {type: Boolean, default: false},
		text: {type: String, required: true},
		closable: {type: Boolean, required: false, default: true},
	},
	emits: ['destructed'],
	data() {
		return {
			showMessage: true,
			progressValue: 100,
			interval: null,
		};
	},
	mounted() {
		this.startTimer();
	},
	methods: {
		hideMessage() {
			if (this.interval) {
				clearInterval(this.interval);
			}

			this.showMessage = false;
			this.$emit('destructed');
		},
		startTimer() {
			if (!this.temporary) {
				return;
			}

			if (this.interval) {
				clearInterval(this.interval);
			}

			this.startProgressLoop();
		},
		stopTimer() {
			if (!this.temporary || !this.interval) {
				return;
			}

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
};
</script>

<style scoped>

.msg {
  word-break: break-word;
  box-shadow: 0 2px 4px -1px rgba(0, 0, 0, 0.2),
  0 4px 5px 0 rgba(0, 0, 0, 0.14),
  0 1px 10px 0 rgba(0, 0, 0, 0.12);
}

/* :deep for deep selection, 0 height and width for removing
the border which 'prominent' introduces */
:deep(.v-alert__icon) {
  height: 0 !important;
  width: 0 !important;
}
</style>
