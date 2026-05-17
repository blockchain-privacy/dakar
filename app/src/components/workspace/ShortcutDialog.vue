<!-- SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="700px"
  >
    <v-card class="mx-auto pb-2">
      <v-card-title>Shortcut Overview</v-card-title>
      <v-card-text>
        <div class="text-body-large">
          <v-table>
            <thead>
              <tr>
                <th class="text-left">
                  Shortcut
                </th>
                <th class="text-left">
                  Description
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="item in shortcuts"
                :key="item.hotkey"
              >
                <td><v-hotkey :keys="item.hotkey" /></td>
                <td>{{ item.description }}</td>
              </tr>
            </tbody>
          </v-table>
        </div>
      </v-card-text>
      <v-card-actions>
        <v-btn
          class="ml-auto"
          text="Close"
          @click="model = false"
        />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
const model = defineModel({type: Boolean});

const shortcuts = [
	{
		hotkey: 'cmd+k',
		description: 'Open the \'Add Entities\' dialog',
	},
	{
		hotkey: 'cmd+A',
		description: 'Select all nodes',
	},
	{
		hotkey: 'esc',
		description: 'Deselect all nodes',
	},
	{
		hotkey: 'delete',
		description: 'Delete the selected nodes',
	},
];

</script>

<style scoped>

</style>
