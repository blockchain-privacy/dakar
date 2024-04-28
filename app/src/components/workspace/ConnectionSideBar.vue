<template>
  <side-bar
    v-model="model"
    title="Connections"
    :icon="mdiArrowLeftRight"
    max-width="648px"
  >
    <template #body>
      <v-card flat>
        <v-card-text>
          <p>Source: {{ connectionSourceUID }}</p>
          <p>Target: {{ connectionTargetUID }}</p>
        </v-card-text>
      </v-card>
    </template>
  </side-bar>
</template>

<script setup>
import {mdiArrowLeftRight} from '@mdi/js';
import SideBar from '@/components/common/SideBar.vue';
import {onUpdated, ref} from 'vue';

const props = defineProps({
	connection: {type: Object, required: true},
});

const model = defineModel({type: Boolean});

let oldConnection = null;
const connectionSourceUID = ref('');
const connectionTargetUID = ref('');

// Hooks

onUpdated(() => {
	if (props.connection?.target?.uid && props.connection.source?.uid) {
		const sourceUID = props.connection.source.uid;
		const targetUID = props.connection.target.uid;

		if (oldConnection && (sourceUID === oldConnection.source.uid && targetUID === oldConnection.target.uid)) {
			return;
		}

		oldConnection = props.connection;
		connectionSourceUID.value = sourceUID;
		connectionTargetUID.value = targetUID;
		getConnectionData();
	}
});

// Functions

function getConnectionData() {
	if (!connectionSourceUID.value || !connectionTargetUID.value) {
		return;
	}

	console.log('pulling data');
}

</script>

<style scoped>

</style>
