<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="350px"
    contained
    no-click-animation
  >
    <v-card>
      <v-card-text class="d-flex align-center flex-column">
        <v-btn
          class="mx-auto text-title-large"
          variant="text"
          :to="to"
          target="_blank"
          @click="model = false"
        >
          <v-icon
            :icon="mdiOpenInNew"
            start
          />
          <div
            class="shorten"
            style="max-width: 270px"
          >
            Go to {{ to.params.id }}
          </div>
        </v-btn>
        <named-divider
          title="Or"
          style="width:100%"
          :vertical-margin="2"
        />
        <v-btn
          class="mx-auto text-title-large"
          variant="text"
          :disabled="disableAddingNodes"
          @click="handleRouteGuardDialogAdd"
        >
          <v-icon
            :icon="mdiPlus"
            start
          />
          Add to Workspace
        </v-btn>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>
<script setup>

import {mdiOpenInNew, mdiPlus} from '@mdi/js';
import NamedDivider from '@/components/common/NamedDivider.vue';

const emit = defineEmits(['addEntities']);
const model = defineModel({type: Boolean});
const props = defineProps({
	to: {type: Object, required: true},
	disableAddingNodes: {type: Boolean, required: true},
});

// Functions
function handleRouteGuardDialogAdd() {
	emit('addEntities', [props.to.params.id]);
	model.value = false;
}

</script>
<style scoped>

.shorten {
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
  margin-right: 2px;
}
</style>
