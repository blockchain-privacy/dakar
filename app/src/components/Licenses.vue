<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-container fluid>
    <alert :text="errorMsg" />
    <v-row
      v-for="type in licenseTypes"
      :key="type.text"
      class="align-center justify-center"
    >
      <v-col
        md="12"
        xl="8"
      >
        <v-card>
          <v-card-title>
            {{ type.text }} Packages
          </v-card-title>
          <v-card-text>
            <v-table v-if="type.ref.value">
              <thead>
                <tr>
                  <th class="text-left">
                    Package
                  </th>
                  <th class="text-left">
                    License
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in type.ref.value"
                  :key="item.name"
                >
                  <td>{{ item.name }}</td>
                  <td v-if="licenseText.has(item.license)">
                    <v-btn
                      variant="tonal"
                      @click="openDialog(item.name, item.license)"
                    >
                      {{ item.license }}
                    </v-btn>
                  </td>
                  <td v-else>
                    {{ item.license }}
                  </td>
                </tr>
              </tbody>
            </v-table>
            <v-skeleton-loader
              v-else
              type="article@3"
            />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    <v-dialog v-model="dialogModel">
      <v-card
        max-width="700px"
        class="mx-auto"
      >
        <v-card-title>
          {{ licenseDialogTitle }} - {{ licenseDialogLicense }}
        </v-card-title>
        <v-card-text>
          {{ licenseDialogText }}
        </v-card-text>
        <v-card-actions>
          <v-btn @click="dialogModel = false">
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup>
import {onMounted, ref} from 'vue';
import {PAGE_TITLE} from '@/constants';
import Alert from '@/components/common/Alert.vue';

const frontendLicenses = ref(null);
const backendLicenses = ref(null);
const errorMsg = ref('');
const licenseDialogTitle = ref('');
const licenseDialogLicense = ref('');
const licenseDialogText = ref('');
const dialogModel = ref(false);

const licenseText = new Map([
	['AGPL-3.0-only', '/licenses/AGPLv3'],
	['AGPL-3.0', '/licenses/AGPLv3'],
	['Apache-2.0', '/licenses/Apache2'],
	['BSD-2-Clause', '/licenses/BSD2'],
	['BSD-3-Clause', '/licenses/BSD3'],
	['ISC', '/licenses/ISC'],
	['MIT', '/licenses/MIT'],
	['MPL-2.0', '/licenses/MPL2'],
	['OFL-1.1', '/licenses/OFL'],
]);

const licenseTypes = [{text: 'JavaScript', ref: frontendLicenses}, {text: 'Golang', ref: backendLicenses}];

// Hooks
onMounted(async () => {
	document.title = `Licenses - ${PAGE_TITLE}`;

	await fetchBackendLicenses('/backend_licenses.json', backendLicenses);
	await fetchBackendLicenses('/frontend_licenses.json', frontendLicenses);
});

// Functions
async function fetchBackendLicenses(url, licenseRef) {
	errorMsg.value = '';
	try {
		const response = await fetch(url);
		licenseRef.value = await response.json();
	} catch (error) {
		errorMsg.value = error.message;
	}
}

async function openDialog(title, license) {
	dialogModel.value = true;
	licenseDialogTitle.value = title;
	licenseDialogLicense.value = license;

	if (!licenseText.has(license)) {
		return;
	}

	try {
		const response = await fetch(licenseText.get(license));
		licenseDialogText.value = await response.text();
	} catch (error) {
		errorMsg.value = error.message;
	}
}

</script>

<style scoped>

</style>
