<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <v-dialog
    v-model="model"
    max-width="350px"
    transition="fade-transition"
  >
    <v-card class="mx-auto">
      <v-card-title>{{ formTitle }}</v-card-title>
      <v-card-text>
        <alert
          :text="errorMsg"
          :type="msgType"
        />
        <v-row>
          <v-form
            ref="modifyIdentityForm"
            validate-on="submit"
          >
            <v-text-field
              v-model="shadowIdentity.email"
              label="E-mail"
              type="email"
              :rules="rules.emailRules"
              style="min-width: 250px"
              autofocus
            />
            <v-select
              v-model="shadowIdentity.state"
              class="mt-2"
              :rules="rules.stateRules"
              :items="states"
              label="State"
            />
            <named-divider
              title="Roles"
              :vertical-margin="3"
            />
            <p class="text-title-small mb-3">
              Applying a role will create a new user if no users currently exist in the system
            </p>
            <v-select
              v-model="shadowIdentity.roles.dakar_dash"
              :items="roles"
              label="Dakar Dash"
              clearable
            />
            <v-select
              v-model="shadowIdentity.roles.dakar_btc"
              :items="roles"
              label="Dakar BTC"
              clearable
            />
            <v-select
              v-model="shadowIdentity.roles.kratos_admin"
              :rules="rules.roleRules"
              :items="roles"
              label="Kratos Admin"
            />
            <template v-if="!createNewUser">
              <named-divider
                title="User UIDs"
                :vertical-margin="3"
              />
              <p class="text-title-small mb-3">
                Leaving these fields empty will result in no changes
              </p>
              <v-text-field
                v-model="shadowIdentity.dakarDashUser"
                label="Dakar Dash UID"
                style="min-width: 250px"
              />
              <v-text-field
                v-model="shadowIdentity.dakarBTCUser"
                label="Dakar BTC UID"
                style="min-width: 250px"
              />
            </template>
          </v-form>
        </v-row>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn
          color="red"
          @click="model = false"
        >
          Cancel
        </v-btn>
        <v-btn @click="saveIdentity">
          Save
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import {
	computed,
	inject,
	onMounted,
	ref,
} from 'vue';
import {emailRules} from '@/utilities';
import NamedDivider from '@/components/common/NamedDivider.vue';
import Alert from '@/components/common/Alert.vue';

const kratosAdmin = inject('kratosadmin');

const emit = defineEmits(['saved']);
const model = defineModel({type: Boolean});
const props = defineProps({
	identity: {type: Object, required: true},
	createNewUser: {type: Boolean, required: true},
});

const msgType = ref('error');
const errorMsg = ref('');
const isLoading = ref(false);
const shadowIdentity = ref({
	// eslint-disable-next-line camelcase
	id: '', email: '', roles: {dakar_dash: '', dakar_btc: '', kratos_admin: ''}, state: '', dakarDashUser: '', dakarBTCUser: '',
});
// Template ref
const modifyIdentityForm = ref(null);

const roles = [{title: 'Admin', value: 'admin'}, {title: 'Privileged', value: 'privileged'}];
const states = [{title: 'Active', value: 'active'}, {title: 'Inactive', value: 'inactive'}];
const rules = {
	roleRules: [
		v => (Boolean(v) && v.length > 0) || 'Role must be set',
	],
	stateRules: [
		v => (Boolean(v) && v.length > 0) || 'State must be set',
	],
	emailRules,
};

// Computed
const formTitle = computed(() => props.createNewUser ? 'Create Identity' : 'Edit Identity');

onMounted(() => {
	shadowIdentity.value = props.identity;
});

function setErrorMessage(msg) {
	msgType.value = 'error';
	errorMsg.value = msg;
}

function setInfoMessage(msg) {
	msgType.value = 'info';
	errorMsg.value = msg;
}

async function saveIdentity() {
	const {valid} = await modifyIdentityForm.value.validate();

	if (!valid) {
		return;
	}

	errorMsg.value = '';

	// Remove all object keys which have no attached value
	Object.keys(shadowIdentity.value.roles).forEach(key => {
		if (!shadowIdentity.value.roles[key]) {
			delete shadowIdentity.value.roles[key];
		}
	});

	isLoading.value = true;
	if (props.createNewUser) {
		try {
			const response = await kratosAdmin.identity.identitiesPost({
				identity: {
					email: shadowIdentity.value.email,
					roles: shadowIdentity.value.roles,
					state: shadowIdentity.value.state,
				},
			});
			if (response.msg) {
				setInfoMessage(response.msg);
			}

			emit('saved');
		} catch (error) {
			setErrorMessage(error);
		}
	} else {
		try {
			const response = await kratosAdmin.identity.identitiesPut({
				identity: {
					uid: shadowIdentity.value.id,
					email: shadowIdentity.value.email,
					roles: shadowIdentity.value.roles,
					state: shadowIdentity.value.state,
					dakarDashUser: shadowIdentity.value.dakarDashUser,
					dakarBTCUser: shadowIdentity.value.dakarBTCUser,
				},
			});

			if (response.msg) {
				setInfoMessage(response.msg);
			}

			emit('saved');
		} catch (error) {
			setErrorMessage(error);
		}
	}

	isLoading.value = false;
	model.value = false;
}

</script>

<style scoped>

</style>
